# Week 2 - Next Steps

## ✅ Completed
1. **OCR Handler** - `internal/media/ocr.go`
   - ✅ `GoogleVisionOCR` struct implementing `OCRHandler` interface
   - ✅ `ExtractFromURL()` and `ExtractFromFile()` methods
   - ✅ Mock mode for testing (no API key needed)
   - ✅ Language detection

2. **YouTube Handler Skeleton** - `internal/media/youtube.go`
   - ✅ `YouTubeVideoProcessor` struct
   - ✅ Video download and frame extraction logic
   - ✅ Fixed all unused parameters

3. **API Integration** - `api/handlers.go` & `api/types.go`
   - ✅ `POST /v1/media/ocr` endpoint
   - ✅ OCRRequest/OCRResponse types
   - ✅ Build passes: `go build ./...`

---

## 📋 Next Steps (In Order)

### Step 1: Simplify YouTube Handler
**Current Issue:** `youtube.go` tries to download full videos and extract frames via FFmpeg (heavy dependencies)

**Better Approach:** Use YouTube Transcript API which is simpler and doesn't require video download

**Tasks:**
- Replace `YouTubeVideoProcessor` with simpler `YouTubeTranscriptFetcher`
- Use REST API calls to fetch transcripts (no heavy external dependencies)
- Extract video ID from URL
- Handle auto-generated vs manual captions
- Implement language detection

**Files to modify:** `internal/media/youtube.go`

### Step 2: Create YouTube API Handler
**Task:** Add `POST /v1/media/youtube` endpoint

**Endpoint:**
```bash
POST /v1/media/youtube
Content-Type: application/json

{
  "project_id": "project-123",
  "source_id": "source-456", 
  "video_url": "https://www.youtube.com/watch?v=..."
}
```

**Response:**
```json
{
  "media_item_id": "media_123",
  "transcript": "Full transcript text...",
  "language": "en",
  "segments": [
    {"timestamp": 0, "text": "First segment..."},
    {"timestamp": 30, "text": "Second segment..."}
  ],
  "is_auto_generated": false,
  "processed_at": "2024-12-18T10:30:45Z"
}
```

**Files to modify:** `api/handlers.go`, `api/types.go`

### Step 3: Create Media Orchestrator
**Task:** Route media items to appropriate handler based on type

**Logic:**
```go
func ProcessMediaItem(ctx context.Context, mediaItem *MediaItem) (*ExtractedContent, error) {
    switch mediaItem.Type {
    case "image":
        return ocrHandler.ExtractFromURL(ctx, mediaItem.URL)
    case "video", "youtube":
        videoID := extractVideoID(mediaItem.URL)
        return youtubeHandler.GetTranscript(ctx, videoID)
    default:
        return nil, fmt.Errorf("unsupported media type: %s", mediaItem.Type)
    }
}
```

**Files to create:** `internal/media/orchestrator.go`

### Step 4: Integration Tests
**Test the full flow:**
```bash
# Test OCR endpoint
curl -X POST http://localhost:8080/v1/media/ocr \
  -H "Content-Type: application/json" \
  -d '{"project_id": "test", "image_url": "https://..."}'

# Test YouTube endpoint
curl -X POST http://localhost:8080/v1/media/youtube \
  -H "Content-Type: application/json" \
  -d '{"project_id": "test", "video_url": "https://youtube.com/watch?v=..."}'
```

---

## 🎯 Priority Order
1. **HIGH** - Simplify YouTube handler (currently over-engineered)
2. **HIGH** - Add YouTube API endpoint
3. **MEDIUM** - Create media orchestrator
4. **MEDIUM** - Integration tests
5. **LOW** - Database integration (save to `extracted_text` table)

---

## 📊 Current Status
- **Build:** ✅ Passing (`go build ./...`)
- **OCR:** ✅ Functional (mock mode)
- **YouTube:** ⚠️ Needs simplification
- **API:** ✅ OCR endpoint working
- **Integration:** 📋 Not yet started

---

## 🚀 After Week 2
- Week 3: Extension endpoint with hybrid search
- Week 4: Browser extension UI and demo
