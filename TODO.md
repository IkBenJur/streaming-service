# TODO

Grouped by urgency, then difficulty within each group.

---

## Critical — bugs / correctness

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 3 | `LoadVideoStatuses` accepts `*Queries` instead of `Querier` interface — breaks the interface abstraction used everywhere else | Easy | `video_status.go:26` |
| 4 | Goroutines in `Submit` create `context.Background()` instead of inheriting the server context — in-flight transcodes won't cancel on shutdown, process can hang | Medium | `videoTranscoder/service.go:36` |
| 5 | `determineTranscodeDurationInUs` returns `0, nil` when no video stream is found — progress will divide by zero | Easy | `videoTranscoder/service.go:173` |

---

## High — missing core features

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 6 | **Streaming endpoints**: serve HLS playlist and segments | Medium | `GET /video/:id/playlist.m3u8` and `GET /video/:id/:segment` — files already exist on disk post-transcode, just need to be served with correct `Content-Type` headers (`application/vnd.apple.mpegurl` and `video/mp4`) |
| 7 | **Status endpoint**: `GET /video/:id/status` → `{ status, progress }` | Easy | Needed by any client to know when a video is ready to stream |
| 8 | Return the video ID from `POST /upload-video` instead of just `"file uploaded"` | Easy | `handler.go:62` — client has no way to reference the video it just uploaded |
| 9 | Load config from environment variables (DB connection string, port, worker count) | Easy | `main.go:25` has a `// TODO` for this — hardcoded values block any deployment |
| 10 | Crash recovery: on startup, re-queue videos stuck in `processing` status | Medium | If the server restarts mid-transcode these videos are permanently stuck — query for them and resubmit to the transcoder |

---

## Medium — tests

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 11 | Unit tests for `validateFileAndGetFileExtension` | Easy | Pure function, no deps — good first test to write |
| 12 | Unit tests for `UploadVideo` handler using mock `VideoProcessingService` and `Transcoder` | Medium | Interfaces are already in place, just needs table-driven tests for happy path, bad extension, save failure |
| 13 | Unit tests for streaming endpoints (once built) | Medium | Test correct `Content-Type`, 404 for missing video/segment |
| 14 | Integration test for full upload → transcode → stream flow | Hard | Needs a real ffmpeg binary and postgres instance — consider a `//go:build integration` tag |

---

## Medium — code quality

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 15 | Clean up partial HLS segments from disk when transcode fails | Easy | `videoTranscoder/service.go` — currently leaves orphaned files |
| 16 | Fix `WriteSucces` typo → `WriteSuccess` | Trivial | `internal/json/json.go:14` |
| 17 | Fix `go.mod`: `go 1.26.2` doesn't exist, and all deps are marked `// indirect` — run `go mod tidy` | Trivial | `go.mod:3` |
| 18 | Add `.gitignore` entries for `go/files/raw/` and `go/files/hls/` | Easy | Raw video files and HLS segments shouldn't be in the repo |
| 19 | Replace wildcard CORS with a configurable allowlist | Easy | `api.go:24` — the `// TODO SETUP CORS` comment |

---

## Low — future features

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 20 | Multiple transcode quality levels (1080p, 720p, 480p) using HLS adaptive bitrate | Hard | Would produce multiple renditions in the m3u8 manifest — the current fMP4 pipeline is already the right foundation |
| 21 | Chunked / resumable upload protocol | Hard | Design doc already identifies this — useful for large files on bad connections |
| 22 | Auth middleware (JWT or session) | Hard | Required before any production use |
| 23 | Video metadata endpoint (`GET /video/:id`) — title, description, duration, status | Medium | Design doc lists this; schema doesn't have title/description yet (migration needed) |
