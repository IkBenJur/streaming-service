# TODO

Grouped by urgency, then difficulty within each group.

---

## Critical — bugs / correctness

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 4 | Graceful shutdown: in-flight transcodes won't cancel on shutdown — process can hang | Medium | `videoTranscoder/service.go` — store a server-lifetime context on `VideoTranscoder`, set at construction time from `main`'s signal context |

---

## High — missing core features

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 7 | **Status endpoint**: `GET /video/:id/status` → `{ status, progress }` | Easy | Needed by any client to know when a video is ready to stream |
| 10 | Crash recovery: on startup, re-queue videos stuck in `processing` status | Medium | If the server restarts mid-transcode these videos are permanently stuck — query for them and resubmit to the transcoder |

---

## Medium — tests

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 14 | Integration test for full upload → transcode → stream flow | Hard | Needs a real ffmpeg binary and postgres instance — consider a `//go:build integration` tag |

---

## Medium — code quality

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 15 | Clean up partial HLS segments from disk when transcode fails | Easy | `videoTranscoder/service.go` — currently leaves orphaned files |
| 16 | Fix `WriteSucces` typo → `WriteSuccess` | Trivial | `internal/json/json.go:14` |
| 17 | Fix `go.mod`: `go 1.26.2` doesn't exist, and all deps are marked `// indirect` — run `go mod tidy` | Trivial | `go.mod:3` |
| 19 | Replace wildcard CORS with a configurable allowlist | Easy | `api.go:24` |

---

## Low — future features

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 20 | Multiple transcode quality levels (1080p, 720p, 480p) using HLS adaptive bitrate | Hard | Would produce multiple renditions in the m3u8 manifest — the current fMP4 pipeline is already the right foundation |
| 21 | Chunked / resumable upload protocol | Hard | Design doc already identifies this — useful for large files on bad connections |
| 22 | Auth middleware (JWT or session) | Hard | Required before any production use |
| 23 | Video metadata endpoint (`GET /video/:id`) — title, description, duration, status | Medium | Design doc lists this; schema doesn't have title/description yet (migration needed) |
