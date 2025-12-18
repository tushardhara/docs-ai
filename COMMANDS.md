# CGAP Quick Command Reference

## 📖 Documentation

```bash
# View all documentation index
cat INDEX.md

# Current status
cat WEEK_1_STATUS.md

# Next week's tasks
cat WEEK_2_PREVIEW.md

# Full sprint overview
cat SPRINT_OVERVIEW.txt
```

## 🔨 Build & Test

```bash
# Build entire project
cd /Users/tushar.dhara/docs-ai && go build ./...

# View all migrations
ls -la db/migrations/

# Check migration files
cat db/migrations/001_add_media_items_table.sql
cat db/migrations/002_add_extracted_text_table.sql
```

## 🗄️ Database (When Running)

```bash
# Set environment
export DATABASE_URL="postgres://cgap:cgap_dev_password@localhost:5432/cgap?sslmode=disable"

# Check migration status
goose -dir db/migrations postgres "$DATABASE_URL" status

# Apply migrations
goose -dir db/migrations postgres "$DATABASE_URL" up

# Rollback latest
goose -dir db/migrations postgres "$DATABASE_URL" down

# Connect to database
docker-compose exec db psql -U cgap -d cgap

# Verify tables exist (in psql):
\dt media_items
\dt extracted_text
\d media_items
\d extracted_text

# View table structure (in psql):
SELECT * FROM media_items LIMIT 1;
SELECT * FROM extracted_text LIMIT 1;
```

## 🐳 Docker

```bash
# Start all services
docker-compose up -d

# Check services
docker-compose ps

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## 📝 Project Structure

```
/Users/tushar.dhara/docs-ai/
├── db/
│   ├── migrations/
│   │   ├── 001_add_media_items_table.sql
│   │   ├── 002_add_extracted_text_table.sql
│   │   ├── migrations.go
│   │   └── run.sh
│   └── schema.sql (existing)
├── internal/
│   ├── postgres/
│   │   └── store.go (updated for migrations)
│   └── ... (existing)
├── cmd/
│   ├── api/
│   ├── worker/
│   └── ... (existing)
├── INDEX.md (read first!)
├── SPRINT_OVERVIEW.txt
├── WEEK_1_DELIVERY.txt
├── WEEK_1_STATUS.md
├── WEEK_2_PREVIEW.md
├── TODO_MVP.md
└── COMMANDS.md (this file)
```

## 🎯 Workflow (Week 2 Start)

```bash
# 1. Pull latest
cd /Users/tushar.dhara/docs-ai
git pull

# 2. Read Week 2 preview
cat WEEK_2_PREVIEW.md

# 3. Set up Google Vision API credentials
# (Instructions in WEEK_2_PREVIEW.md)

# 4. Create media handlers
mkdir -p internal/media
touch internal/media/ocr.go
touch internal/media/youtube.go
touch internal/media/handler.go

# 5. Build and test
go build ./...

# 6. Run tests
go test ./...
```

## 💡 Pro Tips

### When stuck on a task
1. Check INDEX.md for relevant docs
2. Search docs: `grep -r "keyword" .`
3. Check git history: `git log --oneline | head -20`

### For database issues
```bash
# Reset migrations (WARNING: deletes data!)
export DATABASE_URL="postgres://cgap:cgap_dev_password@localhost:5432/cgap?sslmode=disable"
goose -dir db/migrations postgres "$DATABASE_URL" reset

# Reapply
goose -dir db/migrations postgres "$DATABASE_URL" up
```

### For build issues
```bash
# Clean go cache
go clean -cache

# Tidy dependencies
go mod tidy

# Download missing deps
go mod download
```

## 📊 Checking Progress

```bash
# Verify Week 1 complete
echo "Week 1 Tasks:"
echo "✅ go build ./... passes"
cd /Users/tushar.dhara/docs-ai && go build ./... && echo "✅ PASS" || echo "❌ FAIL"

echo ""
echo "✅ Migration files exist"
ls db/migrations/001* db/migrations/002* && echo "✅ PASS" || echo "❌ FAIL"

echo ""
echo "✅ Documentation created"
ls WEEK_1_* INDEX.md SPRINT_OVERVIEW.txt && echo "✅ PASS" || echo "❌ FAIL"
```

## 🚀 Ready for Week 2?

```bash
# Run this checklist
echo "Week 1 Completion Checklist:"
echo ""

# 1. Build
cd /Users/tushar.dhara/docs-ai
go build ./... && echo "✅ Build passes" || echo "❌ Build fails"

# 2. Migrations exist
[ -f db/migrations/001_add_media_items_table.sql ] && echo "✅ Migration 001 exists" || echo "❌ Missing"
[ -f db/migrations/002_add_extracted_text_table.sql ] && echo "✅ Migration 002 exists" || echo "❌ Missing"

# 3. Documentation
[ -f INDEX.md ] && echo "✅ INDEX.md created" || echo "❌ Missing"
[ -f WEEK_2_PREVIEW.md ] && echo "✅ Week 2 preview ready" || echo "❌ Missing"

echo ""
echo "If all ✅, Week 1 is complete. Ready for Week 2!"
```

## 📚 Documentation URLs/Paths

| Doc | Path | Use |
|-----|------|-----|
| Index | INDEX.md | Start here |
| Status | WEEK_1_STATUS.md | Current progress |
| Timeline | SPRINT_OVERVIEW.txt | Visual roadmap |
| Week 2 | WEEK_2_PREVIEW.md | Next tasks |
| Checklist | TODO_MVP.md | Full sprint |
| Commands | COMMANDS.md | This file |

---

**Status: All files ready. Build passes. Ready for Week 2! 🚀**
