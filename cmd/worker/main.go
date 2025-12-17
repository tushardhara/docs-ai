package main

import (
	"context"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"

	"cgap/api"
	"cgap/internal/embedding"
	"cgap/internal/postgres"
	"cgap/internal/queue"
)

func main() {
	log.Println("cgap worker starting...")

	// Load configuration from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Initialize PostgreSQL storage with pgx
	store, err := postgres.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize postgres store: %v", err)
	}
	defer store.Close()

	// Initialize Redis client (supports host:port and redis:// URI)
	redisClient := redis.NewClient(redisOptionsFromEnv())
	defer redisClient.Close()

	// Initialize Redis queue consumer
	consumer := queue.NewConsumer(redisClient)

	// Build embedder for ingestion
	embedder := buildEmbedder()

	// Start job processor
	go func() {
		log.Println("Starting job consumer...")
		ctx := context.Background()
		for {
			// Get next task
			task, err := consumer.Process(ctx)
			if err != nil {
				log.Printf("Consumer error: %v", err)
				continue
			}

			// Nil task means timeout - no task available
			if task == nil {
				continue
			}

			// Route task to appropriate handler based on task.Type
			log.Printf("Processing task: type=%s, id=%s", task.Type, task.ID)

			switch task.Type {
			case "ingest":
				if err := handleIngest(ctx, store, embedder, redisClient, task.ID, task.Payload); err != nil {
					log.Printf("ingest error (task=%s): %v", task.ID, err)
					// Mark failed
					_ = markJobFailed(ctx, redisClient, task.ID, err)
				} else {
					log.Printf("ingest completed (task=%s)", task.ID)
					// Mark completed
					_ = markJobCompleted(ctx, redisClient, task.ID)
				}
			default:
				// Not handled yet
			}
		}
	}()

	log.Println("Worker ready")

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker...")
	log.Println("Worker stopped")
}

// buildEmbedder constructs an embedder from environment configuration.
func buildEmbedder() embedding.Embedder {
	provider := os.Getenv("EMBEDDING_PROVIDER")
	model := os.Getenv("EMBEDDING_MODEL")
	if provider == "" {
		provider = "google"
	}
	switch provider {
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			log.Println("OPENAI_API_KEY not set; embeddings may fail")
		}
		return embedding.NewOpenAIEmbedder(key, model)
	case "http":
		return embedding.NewHTTPEmbedder(os.Getenv("EMBEDDING_ENDPOINT"), model, os.Getenv("EMBEDDING_API_KEY"), os.Getenv("EMBEDDING_AUTH_HEADER"))
	case "mock":
		return embedding.NewMockEmbedder(768)
	default: // google
		key := os.Getenv("GEMINI_API_KEY")
		if key == "" {
			log.Println("GEMINI_API_KEY not set; embeddings may fail")
		}
		return embedding.NewGoogleEmbedder(key, model)
	}
}

// redisOptionsFromEnv parses REDIS_URL and returns go-redis options.
// Supports formats:
//   - host:port
//   - redis://[:password@]host:port[/db]?db=...
//   - rediss:// (TLS not configured here)
func redisOptionsFromEnv() *redis.Options {
	raw := os.Getenv("REDIS_URL")
	if raw == "" {
		return &redis.Options{Addr: "localhost:6379"}
	}
	if strings.HasPrefix(raw, "redis://") || strings.HasPrefix(raw, "rediss://") {
		u, err := neturl.Parse(raw)
		if err != nil {
			return &redis.Options{Addr: raw}
		}
		opt := &redis.Options{Addr: u.Host}
		if u.User != nil {
			opt.Username = u.User.Username()
			if pw, ok := u.User.Password(); ok {
				opt.Password = pw
			}
		}
		if dbStr := strings.TrimPrefix(u.Path, "/"); dbStr != "" {
			if n, err := strconv.Atoi(dbStr); err == nil {
				opt.DB = n
			}
		}
		if qdb := u.Query().Get("db"); qdb != "" {
			if n, err := strconv.Atoi(qdb); err == nil {
				opt.DB = n
			}
		}
		return opt
	}
	return &redis.Options{Addr: raw}
}

// handleIngest performs a minimal ingestion: fetch content from URL(s),
// create document and one-or-more chunks, embed and store in Postgres.
func handleIngest(ctx context.Context, store *postgres.Store, emb embedding.Embedder, rdb *redis.Client, jobID string, payload any) error {
	// Decode payload into API DTO
	mp, ok := payload.(map[string]any)
	if !ok {
		return nil // ignore unknown payload format
	}

	// Quick, robust decode: only fields we need
	p := api.IngestTaskPayload{}
	if v, ok := mp["project_id"].(string); ok {
		p.ProjectID = v
	}
	if v, ok := mp["fail_fast"].(bool); ok {
		p.FailFast = v
	}
	if src, ok := mp["source"].(map[string]any); ok {
		p.Source.Type, _ = src["type"].(string)
		p.Source.URL, _ = src["url"].(string)
		if crawl, ok := src["crawl"].(map[string]any); ok {
			cs := &api.CrawlSpec{}
			if v, ok := crawl["mode"].(string); ok {
				cs.Mode = v
			}
			if v, ok := crawl["start_url"].(string); ok {
				cs.StartURL = v
			}
			if v, ok := crawl["sitemap_url"].(string); ok {
				cs.SitemapURL = v
			}
			if v, ok := crawl["scope"].(string); ok {
				cs.Scope = v
			}
			if v, ok := crawl["max_depth"].(float64); ok {
				cs.MaxDepth = int(v)
			}
			if v, ok := crawl["max_pages"].(float64); ok {
				cs.MaxPages = int(v)
			}
			if v, ok := crawl["respect_robots"].(bool); ok {
				cs.RespectRobots = v
			}
			if v, ok := crawl["concurrency"].(float64); ok {
				cs.Concurrency = int(v)
			}
			if v, ok := crawl["delay_ms"].(float64); ok {
				cs.DelayMS = int(v)
			}
			if v, ok := crawl["allow"].([]any); ok {
				for _, it := range v {
					if s, ok := it.(string); ok {
						cs.Allow = append(cs.Allow, s)
					}
				}
			}
			if v, ok := crawl["deny"].([]any); ok {
				for _, it := range v {
					if s, ok := it.(string); ok {
						cs.Deny = append(cs.Deny, s)
					}
				}
			}
			p.Source.Crawl = cs
		}
		if files, ok := src["files"].(map[string]any); ok {
			if urls, ok := files["urls"].([]any); ok {
				for _, u := range urls {
					if s, ok := u.(string); ok {
						if p.Source.Files == nil {
							p.Source.Files = &api.FileSpec{}
						}
						p.Source.Files.URLs = append(p.Source.Files.URLs, s)
					}
				}
			}
			p.Source.Files.Format, _ = files["format"].(string)
		}
	}

	if p.ProjectID == "" {
		return nil
	}

	urls := make([]string, 0, 32)
	if p.Source.Type == "crawl" && p.Source.Crawl != nil {
		// Build URL set based on crawl mode
		list, err := buildCrawlURLList(ctx, p.Source.Crawl)
		if err != nil {
			return err
		}
		urls = append(urls, list...)
	} else {
		if p.Source.URL != "" {
			urls = append(urls, p.Source.URL)
		}
		if p.Source.Files != nil && len(p.Source.Files.URLs) > 0 {
			urls = append(urls, p.Source.Files.URLs...)
		}
	}
	if len(urls) == 0 {
		return nil
	}

	pool := store.Pool()
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Resolve project slug -> UUID if needed
	pid := p.ProjectID
	if !looksLikeUUID(pid) {
		if err := pool.QueryRow(ctx, `SELECT id FROM projects WHERE slug = $1`, pid).Scan(&pid); err != nil {
			return err
		}
	}

	// Initialize running status
	_ = markJobRunning(ctx, rdb, jobID, pid, len(urls))

	// Concurrency settings: default to 4; use crawl.concurrency when provided
	maxWorkers := 4
	if p.Source.Crawl != nil && p.Source.Crawl.Concurrency > 0 {
		maxWorkers = p.Source.Crawl.Concurrency
	}
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if maxWorkers > 16 {
		maxWorkers = 16
	}

	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var firstErr error
	var once sync.Once

	for _, u := range urls {
		u := u // capture
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if workCtx.Err() != nil {
				return
			}
			// politeness delay per-request if configured
			if p.Source.Crawl != nil && p.Source.Crawl.DelayMS > 0 {
				select {
				case <-time.After(time.Duration(p.Source.Crawl.DelayMS) * time.Millisecond):
				case <-workCtx.Done():
					return
				}
			}
			if err := processURL(workCtx, pool, httpClient, emb, pid, p.Source, u); err != nil {
				log.Printf("ingest: error processing %s: %v", u, err)
				if p.FailFast {
					once.Do(func() {
						firstErr = err
						cancel()
					})
				}
			}
			_ = incJobProcessed(ctx, rdb, jobID, 1)
		}()
	}
	wg.Wait()
	if p.FailFast && firstErr != nil {
		return firstErr
	}
	return nil
}

// processURL fetches, normalizes, chunks, embeds, and stores a single URL.
func processURL(ctx context.Context, pool *pgxpool.Pool, httpClient *http.Client, emb embedding.Embedder, projectID string, src api.SourceSpec, u string) error {
	// Fetch content
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("fetch failed: %s status=%d", u, resp.StatusCode)
		return nil
	}

	text := string(body)
	// naive text cleanup for markdown/plain; leave HTML as-is for now
	if strings.HasSuffix(strings.ToLower(u), ".md") || src.Files != nil && (src.Files.Format == "markdown" || src.Files.Format == "md") {
		// keep as-is
	} else if src.Files != nil && (src.Files.Format == "txt" || src.Files.Format == "text") {
		// keep as-is
	} else {
		// basic fallback: strip tags lightly
		text = stripBasicHTML(text)
	}

	// Upsert document (by project_id + uri)
	var docID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO documents (project_id, uri, title)
		VALUES ($1, $2, COALESCE($3, 'Untitled'))
		ON CONFLICT (project_id, uri) DO UPDATE SET title = COALESCE(EXCLUDED.title, documents.title)
		RETURNING id
	`, projectID, u, "").Scan(&docID); err != nil {
		return err
	}

	// Very simple chunking: split by blank lines, cap to first N chunks
	parts := splitIntoParagraphs(text)
	if len(parts) == 0 {
		return nil
	}
	const maxChunks = 20
	if len(parts) > maxChunks {
		parts = parts[:maxChunks]
	}

	// Insert chunks and embeddings
	for i, t := range parts {
		var chunkID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO chunks (document_id, ord, text)
			VALUES ($1, $2, $3)
			RETURNING id
		`, docID, i, t).Scan(&chunkID); err != nil {
			return err
		}

		vec, err := emb.Embed(ctx, t)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO chunk_embeddings (chunk_id, embedding)
			VALUES ($1, $2)
			ON CONFLICT (chunk_id) DO UPDATE SET embedding = EXCLUDED.embedding
		`, chunkID, pgvector.NewVector(vec)); err != nil {
			return err
		}
	}
	return nil
}

// buildCrawlURLList expands a CrawlSpec to a list of URLs to fetch.
func buildCrawlURLList(ctx context.Context, cs *api.CrawlSpec) ([]string, error) {
	if cs == nil {
		return nil, nil
	}
	mode := cs.Mode
	if mode == "" {
		mode = "crawl"
	}
	switch mode {
	case "single":
		if cs.StartURL == "" {
			return nil, nil
		}
		if cs.RespectRobots && !isAllowedByRobots(ctx, cs.StartURL) {
			return nil, nil
		}
		return []string{cs.StartURL}, nil
	case "sitemap":
		if cs.SitemapURL == "" {
			return nil, nil
		}
		urls, _ := parseSitemap(ctx, cs.SitemapURL)
		urls = filterURLs(urls, cs)
		if cs.MaxPages > 0 && len(urls) > cs.MaxPages {
			urls = urls[:cs.MaxPages]
		}
		return urls, nil
	case "crawl":
		if cs.StartURL == "" {
			return nil, nil
		}
		return crawlBFS(ctx, cs)
	default:
		return nil, nil
	}
}

// parseSitemap parses a simple sitemap.xml (urlset) and returns URL list.
func parseSitemap(ctx context.Context, sitemapURL string) ([]string, error) {
	visited := map[string]bool{}
	var out []string
	client := &http.Client{Timeout: 30 * time.Second}
	var fetch func(string, int) error
	fetch = func(url string, depth int) error {
		if visited[url] || depth > 3 {
			return nil
		}
		visited[url] = true
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil
		}
		type urlEntry struct {
			Loc string `xml:"loc"`
		}
		type urlSet struct {
			URLs []urlEntry `xml:"url"`
		}
		type sitemapEntry struct {
			Loc string `xml:"loc"`
		}
		type sitemapIndex struct {
			Maps []sitemapEntry `xml:"sitemap"`
		}
		u := urlSet{}
		if err := xml.Unmarshal(data, &u); err == nil && len(u.URLs) > 0 {
			for _, e := range u.URLs {
				if e.Loc != "" {
					out = append(out, strings.TrimSpace(e.Loc))
				}
			}
			return nil
		}
		si := sitemapIndex{}
		if err := xml.Unmarshal(data, &si); err == nil && len(si.Maps) > 0 {
			limit := 0
			for _, e := range si.Maps {
				if e.Loc == "" {
					continue
				}
				if limit >= 50 { // cap nested sitemaps
					break
				}
				limit++
				_ = fetch(strings.TrimSpace(e.Loc), depth+1)
				if len(out) > 5000 { // safety cap
					break
				}
			}
		}
		return nil
	}
	_ = fetch(sitemapURL, 0)
	return dedup(out), nil
}

// crawlBFS performs a simple single-threaded BFS crawl with dedup and limits.
func crawlBFS(ctx context.Context, cs *api.CrawlSpec) ([]string, error) {
	maxDepth := cs.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 2
	}
	maxPages := cs.MaxPages
	if maxPages <= 0 {
		maxPages = 200
	}

	start := cs.StartURL
	base, err := neturl.Parse(start)
	if err != nil {
		return []string{start}, nil
	}

	q := []string{start}
	depth := map[string]int{start: 0}
	seen := map[string]bool{start: true}
	out := make([]string, 0, maxPages)
	client := &http.Client{Timeout: 20 * time.Second}

	for len(q) > 0 && len(out) < maxPages {
		u := q[0]
		q = q[1:]
		d := depth[u]
		if cs.RespectRobots && !isAllowedByRobots(ctx, u) {
			continue
		}
		if !withinScope(u, base, cs.Scope) {
			continue
		}
		if !passesAllowDeny(u, cs.Allow, cs.Deny) {
			continue
		}

		out = append(out, u)
		if d >= maxDepth {
			continue
		}

		// Fetch page and extract links
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			continue
		}
		links := extractLinks(string(body), u)
		for _, link := range links {
			if !withinScope(link, base, cs.Scope) {
				continue
			}
			if !passesAllowDeny(link, cs.Allow, cs.Deny) {
				continue
			}
			if seen[link] {
				continue
			}
			seen[link] = true
			depth[link] = d + 1
			q = append(q, link)
			if len(seen) >= maxPages*3 {
				break
			}
		}
		if cs.DelayMS > 0 {
			time.Sleep(time.Duration(cs.DelayMS) * time.Millisecond)
		}
	}
	if maxPages > 0 && len(out) > maxPages {
		out = out[:maxPages]
	}
	return out, nil
}

// isAllowedByRobots checks Disallow rules for User-agent: *
func isAllowedByRobots(ctx context.Context, rawURL string) bool {
	u, err := neturl.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return true
	}
	robotsURL := u.Scheme + "://" + u.Host + "/robots.txt"
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return true
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return true
	}
	data, _ := io.ReadAll(resp.Body)
	return robotsAllow(string(data), u.Path)
}

// filterURLs applies scope, allow/deny and robots to a list derived from sitemap.
func filterURLs(in []string, cs *api.CrawlSpec) []string {
	if len(in) == 0 {
		return in
	}
	// Determine base from StartURL, else from first URL in list, else from sitemap URL
	var base *neturl.URL
	if cs.StartURL != "" {
		if u, err := neturl.Parse(cs.StartURL); err == nil {
			base = u
		}
	}
	if base == nil {
		if u, err := neturl.Parse(in[0]); err == nil {
			base = u
		}
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, raw := range in {
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}
		u := raw
		if base != nil {
			if pu, err := neturl.Parse(raw); err == nil && !pu.IsAbs() {
				u = base.ResolveReference(pu).String()
			}
		}
		if base != nil && !withinScope(u, base, cs.Scope) {
			continue
		}
		if !passesAllowDeny(u, cs.Allow, cs.Deny) {
			continue
		}
		if cs.RespectRobots && !isAllowedByRobots(context.Background(), u) {
			continue
		}
		out = append(out, u)
	}
	return out
}

// robotsAllow parses minimal robots.txt rules for UA "*" and checks path.
func robotsAllow(txt string, p string) bool {
	uaStar := false
	allows := make([]string, 0, 8)
	disallows := make([]string, 0, 8)
	lines := strings.Split(txt, "\n")
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		low := strings.ToLower(s)
		if strings.HasPrefix(low, "user-agent:") {
			v := strings.TrimSpace(strings.TrimPrefix(low, "user-agent:"))
			uaStar = (v == "*")
			continue
		}
		if !uaStar {
			continue
		}
		if strings.HasPrefix(low, "disallow:") {
			v := strings.TrimSpace(strings.TrimPrefix(s, "Disallow:"))
			if v != "" {
				disallows = append(disallows, v)
			}
			continue
		}
		if strings.HasPrefix(low, "allow:") {
			v := strings.TrimSpace(strings.TrimPrefix(s, "Allow:"))
			if v != "" {
				allows = append(allows, v)
			}
			continue
		}
	}
	bestAllow := 0
	for _, a := range allows {
		if strings.HasPrefix(p, a) {
			if len(a) > bestAllow {
				bestAllow = len(a)
			}
		}
	}
	bestDis := 0
	for _, d := range disallows {
		if strings.HasPrefix(p, d) {
			if len(d) > bestDis {
				bestDis = len(d)
			}
		}
	}
	if bestAllow >= bestDis {
		return true
	}
	if bestDis > 0 {
		return false
	}
	return true
}

// withinScope returns true if link is within the configured scope relative to base.
func withinScope(link string, base *neturl.URL, scope string) bool {
	u, err := neturl.Parse(link)
	if err != nil {
		return false
	}
	if !u.IsAbs() {
		u = base.ResolveReference(u)
	}
	switch scope {
	case "prefix":
		b := base.String()
		if !strings.HasSuffix(b, "/") {
			b += "/"
		}
		s := u.String()
		return strings.HasPrefix(s, b)
	case "domain":
		// naive fallback: host suffix match (registrable domain requires publicsuffix)
		return strings.HasSuffix(u.Host, hostSuffix(base.Host))
	default: // host
		return normalizeHost(u.Host) == normalizeHost(base.Host)
	}
}

func normalizeHost(h string) string { return strings.TrimPrefix(strings.ToLower(h), "www.") }
func hostSuffix(h string) string {
	h = normalizeHost(h)
	if i := strings.LastIndex(h, "."); i > 0 {
		p := h[:i]
		s := h[i+1:]
		if j := strings.LastIndex(p, "."); j > 0 {
			return p[j+1:] + "." + s
		}
	}
	return h
}

// passesAllowDeny applies allow/deny regex patterns if provided.
func passesAllowDeny(u string, allow, deny []string) bool {
	// Deny has precedence
	for _, pat := range deny {
		if pat == "" {
			continue
		}
		if reMatch(pat, u) {
			return false
		}
	}
	if len(allow) == 0 {
		return true
	}
	for _, pat := range allow {
		if pat == "" {
			continue
		}
		if reMatch(pat, u) {
			return true
		}
	}
	return false
}

func reMatch(pat, s string) bool {
	// Treat pattern as regex; if invalid, fallback to substring
	re, err := regexp.Compile(pat)
	if err != nil {
		return strings.Contains(s, pat)
	}
	return re.MatchString(s)
}

// extractLinks pulls href values from anchor tags and resolves relative links.
func extractLinks(html, baseURL string) []string {
	hrefRe := regexp.MustCompile(`(?i)<a[^>]+href=["']([^"']+)["']`)
	m := hrefRe.FindAllStringSubmatch(html, -1)
	if len(m) == 0 {
		return nil
	}
	base, err := neturl.Parse(baseURL)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(m))
	for _, g := range m {
		raw := strings.TrimSpace(g[1])
		if raw == "" {
			continue
		}
		// skip fragments and mailto
		if strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "mailto:") || strings.HasPrefix(raw, "javascript:") {
			continue
		}
		u, err := neturl.Parse(raw)
		if err != nil {
			continue
		}
		if !u.IsAbs() {
			u = base.ResolveReference(u)
		}
		// drop anchors
		u.Fragment = ""
		// normalize index-like paths (optional)
		u.Path = path.Clean(u.Path)
		out = append(out, u.String())
	}
	return dedup(out)
}

func dedup(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// --- Job status helpers (Redis-backed) ---

func jobKey(id string) string { return "cgap:job:" + id }

func markJobRunning(ctx context.Context, rdb *redis.Client, jobID, projectID string, total int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	return rdb.HSet(ctx, jobKey(jobID), map[string]any{
		"job_id":     jobID,
		"project_id": projectID,
		"status":     "running",
		"processed":  0,
		"total":      total,
		"started_at": now,
		"updated_at": now,
	}).Err()
}

func incJobProcessed(ctx context.Context, rdb *redis.Client, jobID string, delta int) error {
	if delta == 0 {
		return nil
	}
	if err := rdb.HIncrBy(ctx, jobKey(jobID), "processed", int64(delta)).Err(); err != nil {
		return err
	}
	return rdb.HSet(ctx, jobKey(jobID), "updated_at", time.Now().UTC().Format(time.RFC3339)).Err()
}

func markJobCompleted(ctx context.Context, rdb *redis.Client, jobID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	return rdb.HSet(ctx, jobKey(jobID), map[string]any{
		"status":      "completed",
		"finished_at": now,
		"updated_at":  now,
	}).Err()
}

func markJobFailed(ctx context.Context, rdb *redis.Client, jobID string, err error) error {
	now := time.Now().UTC().Format(time.RFC3339)
	return rdb.HSet(ctx, jobKey(jobID), map[string]any{
		"status":      "failed",
		"error":       err.Error(),
		"finished_at": now,
		"updated_at":  now,
	}).Err()
}

func setJobTotal(ctx context.Context, rdb *redis.Client, jobID string, total int) error {
	return rdb.HSet(ctx, jobKey(jobID), map[string]any{
		"total":      total,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}).Err()
}

func splitIntoParagraphs(s string) []string {
	// split on blank lines and trim
	blocks := strings.Split(s, "\n\n")
	out := make([]string, 0, len(blocks))
	for _, b := range blocks {
		t := strings.TrimSpace(b)
		if t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 && strings.TrimSpace(s) != "" {
		// fallback to single chunk
		return []string{strings.TrimSpace(s)}
	}
	return out
}

func stripBasicHTML(s string) string {
	// very naive: remove <...> tags and collapse whitespace
	b := make([]rune, 0, len(s))
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
			continue
		case '>':
			inTag = false
			continue
		}
		if !inTag {
			b = append(b, r)
		}
	}
	t := strings.ReplaceAll(string(b), "\r", "")
	t = strings.ReplaceAll(t, "\t", " ")
	// collapse multiple newlines
	for strings.Contains(t, "\n\n\n") {
		t = strings.ReplaceAll(t, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(t)
}

// looksLikeUUID reports whether s matches canonical UUID v1-5 format.
var uuidReWorker = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func looksLikeUUID(s string) bool { return uuidReWorker.MatchString(s) }
