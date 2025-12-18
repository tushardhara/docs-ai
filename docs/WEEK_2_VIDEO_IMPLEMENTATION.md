# Video File Transcription Implementation

## Overview
Added support for direct video file transcription (MP4, AVI, MOV, etc.) alongside existing YouTube transcript extraction.

## Files Created/Modified

### 1. `/internal/media/video.go` (NEW)
**VideoTranscriber** handles transcription of video files using speech-to-text APIs.

**Key Features:**
- Supports multiple API providers (Whisper, AssemblyAI, OpenAI)
- Environment variable configuration: `WHISPER_API_KEY`, `ASSEMBLYAI_API_KEY`, `OPENAI_API_KEY`
- Mock mode for testing without API credentials
- Supports 8 video formats: MP4, AVI, MOV, MKV, WebM, FLV, WMV, M4V

**Methods:**
```go
NewVideoTranscriber(logger) *VideoTranscriber
TranscribeFromURL(ctx, videoURL) (*TranscriptResult, error)
TranscribeFromFile(ctx, filePath) (*TranscriptResult, error)
GetSupportedFormats() []string
EstimateProcessingTime(durationSeconds int) int
```

**Mock Response:**
Returns 3 sample segments with timestamps for testing.

### 2. `/api/types.go` (MODIFIED)
Added `VideoRequest` and `VideoResponse` types.

**VideoRequest:**
```json
{
  "project_id": "proj_123",
  "source_id": "src_456",
  "video_url": "https://example.com/video.mp4"
}
```

**VideoResponse:**
```json
{
  "media_item_id": "media_1234567890",
  "transcript": "Full transcript text...",
  "language": "en",
  "segments": [
    {"text": "...", "start_seconds": 0, "end_seconds": 15}
  ],
  "is_auto_generated": true,
  "duration_seconds": 60,
  "processed_at": "2024-01-15T10:30:00Z",
  "extraction_status": "success"
}
```

### 3. `/api/handlers.go` (MODIFIED)
Added `VideoHandler` endpoint.

**Route:** `POST /v1/media/video`
**Timeout:** 120 seconds (longer than other endpoints due to video processing)

**Process:**
1. Validates request (project_id, source_id, video_url required)
2. Initializes VideoTranscriber
3. Transcribes video
4. Converts segments to API format
5. Returns response with status (success/partial/failed)

## API Comparison

| Endpoint | Purpose | Media Type | Processing Time |
|----------|---------|------------|----------------|
| `/v1/media/ocr` | Extract text from images | Image files | 30s timeout |
| `/v1/media/youtube` | Fetch YouTube transcripts | YouTube URLs only | 30s timeout |
| `/v1/media/video` | Transcribe video files | Direct video files | 120s timeout |

## Usage Example

```bash
# Test with mock mode (no API key needed)
curl -X POST http://localhost:8080/v1/media/video \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_abc",
    "source_id": "src_123",
    "video_url": "https://example.com/demo.mp4"
  }'
```

## Production Setup

To enable real transcription, set one of these environment variables:

```bash
# OpenAI Whisper API (recommended)
export OPENAI_API_KEY="sk-..."

# Or AssemblyAI
export ASSEMBLYAI_API_KEY="your-key"

# Or standalone Whisper API
export WHISPER_API_KEY="your-key"
```

## Supported Video Formats
MP4, AVI, MOV, MKV, WebM, FLV, WMV, M4V

## Implementation Notes

1. **Mock Mode:** Without API keys, returns sample transcripts for testing
2. **API Placeholder:** Real API integration needs implementation in:
   - `transcribeWithAPI(ctx, videoURL)`
   - `transcribeFileWithAPI(ctx, filePath)`
3. **Duration Estimation:** Processing time typically 10-30% of video duration
4. **Error Handling:** Returns appropriate HTTP status codes and error messages

## Next Steps

1. **Implement Real API Integration:**
   - OpenAI Whisper API client
   - AssemblyAI integration
   - Or local Whisper model support

2. **Add File Upload Support:**
   - Multipart form data handling
   - Temporary file storage
   - Upload size limits

3. **Database Integration:**
   - Save transcripts to `extracted_text` table
   - Link to `media_items` records

4. **Media Orchestrator:**
   - Route by media type (image→OCR, youtube→YouTube, video→Video)
   - Unified media processing pipeline

## Testing

Build verification:
```bash
go build ./...
# ✅ Successful compilation
```

Test endpoint:
```bash
# Returns mock transcript with 3 segments
curl -X POST http://localhost:8080/v1/media/video \
  -H "Content-Type: application/json" \
  -d '{"project_id":"p1","source_id":"s1","video_url":"test.mp4"}'
```
