# HLS Streaming Implementation Plan

## Summary

The current streaming service implements a basic custom byte-range protocol: the client requests a video chunk via `?start=` and `?end=` query params, and the server responds with `206 Partial Content`. No standard player supports this out of the box, and it lacks adaptive bitrate, progress tracking, or crash recovery.

The goal is to replace this with a proper **HLS (HTTP Live Streaming)** pipeline using **fragmented MP4 (fMP4)** segments, which is what modern streaming services (YouTube, Netflix) use. fMP4 is compatible with both HLS and DASH, making it a good long-term foundation.

---

## Architecture

```
POST /upload-video
  → save raw file to disk
  → insert video row in DB (status: "pending")
  → go transcodeVideo(videoID, filePath)   ← non-blocking
  → return videoID to client

transcodeVideo (goroutine)
  → update DB status: "processing"
  → exec FFmpeg, parse stderr for progress
  → write init.mp4 + segment_000.m4s, segment_001.m4s, ...
  → write playlist.m3u8
  → update DB status: "ready" (or "failed" on error)

GET /video/:id/status         → poll until ready
GET /video/:id/playlist.m3u8  → serve manifest
GET /video/:id/:segment       → serve init.mp4 / segment_NNN.m4s
```

---

## Implementation Steps

### 1. Database

Set up a `videos` table:

```sql
CREATE TABLE videos (
    id          UUID PRIMARY KEY,
    status      TEXT NOT NULL,  -- pending, processing, ready, failed
    progress    INT,            -- 0-100, updated by goroutine
    file_path   TEXT,
    error       TEXT,
    created_at  TIMESTAMP
);
```

### 2. Upload endpoint

- Accept the file via multipart form
- Save the raw file to disk (e.g. `./files/raw/:id.mp4`)
- Insert a DB row with `status: "pending"`
- Spawn `go transcodeVideo(id, path)`
- Return the `videoID` immediately — do not block

### 3. Chunked upload (for large files)

Replace single-part upload with a three-step protocol to avoid memory pressure and support resumable uploads:

```
POST /upload/init              → returns uploadID
PUT  /upload/:id/chunk/:n      → stream each chunk (~5MB) straight to disk
POST /upload/:id/complete      → assemble chunks, kick off transcode
```

### 4. Transcode goroutine

Run FFmpeg to produce fMP4 segments and a manifest:

```bash
ffmpeg -i input.mp4 \
  -c:v libx264 -c:a aac \
  -f hls \
  -hls_time 6 \
  -hls_playlist_type vod \
  -hls_segment_type fmp4 \
  -hls_segment_filename "files/:id/segment_%03d.m4s" \
  "files/:id/playlist.m3u8"
```

Parse FFmpeg's stderr (`-progress pipe:2`) to extract progress and update the DB in real time.

Handle failure:

```go
err := cmd.Run()
if err != nil {
    // delete partial segments from disk
    // update DB: status = "failed", error = err.Error()
    // optionally retry once
}
```

### 5. Worker pool (concurrency limit)

Cap simultaneous FFmpeg processes to avoid overwhelming the server:

```go
var sem = make(chan struct{}, 4) // max 4 concurrent transcode jobs

func transcodeVideo(id, path string) {
    sem <- struct{}{}
    defer func() { <-sem }()
    // run ffmpeg...
}
```

### 6. Crash recovery

On server startup, re-queue any videos stuck in `"processing"` status — these were interrupted by a restart:

```go
// on startup
rows := db.Query(`SELECT id, file_path FROM videos WHERE status = 'processing'`)
for rows.Next() {
    go transcodeVideo(row.ID, row.FilePath)
}
```

### 7. Manifest and segment endpoints

```
GET /video/:id/playlist.m3u8
  → serve files/:id/playlist.m3u8 with Content-Type: application/vnd.apple.mpegurl

GET /video/:id/:segment
  → serve files/:id/:segment
  → .m4s files: Content-Type: video/mp4
  → init.mp4:   Content-Type: video/mp4
```

### 8. Status endpoint

```
GET /video/:id/status
→ { "status": "processing", "progress": 42 }
→ { "status": "ready" }
→ { "status": "failed", "error": "..." }
```

---

## File layout on disk

```
files/
  raw/
    <videoID>.mp4          ← original upload
  <videoID>/
    playlist.m3u8          ← HLS manifest
    init.mp4               ← fMP4 init segment
    segment_000.m4s
    segment_001.m4s
    ...
```

---

## Resources

- [HLS spec (Apple)](https://developer.apple.com/documentation/http-live-streaming) — the authoritative HLS reference
- [FFmpeg HLS muxer docs](https://ffmpeg.org/ffmpeg-formats.html#hls-1) — all `-hls_*` flags explained
- [FFmpeg fMP4/CMAF guide](https://ffmpeg.org/ffmpeg-formats.html#toc-Options-51) — `hls_segment_type fmp4` specifics
- [tus resumable upload protocol](https://tus.io/) — open spec for chunked uploads, has Go server implementations
- [HLS.js](https://github.com/video-dev/hls.js) — JavaScript HLS player for non-Safari browsers
- [Go `os/exec` docs](https://pkg.go.dev/os/exec) — for spawning and managing FFmpeg processes
- [pgx](https://github.com/jackc/pgx) — idiomatic Go Postgres driver if you go with Postgres
