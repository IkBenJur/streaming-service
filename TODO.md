# TODO

Grouped by urgency, then difficulty within each group.

## High — missing core features

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 7 | **Status endpoint**: `GET /video/:id/status` → `{ status, progress }` | Easy | Needed by any client to know when a video is ready to stream |

---

## Medium — tests

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 14 | Integration test for full upload → transcode → stream flow | Hard | Needs a real ffmpeg binary and postgres instance — consider a `//go:build integration` tag |

---

## Medium — code quality

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 15 | Clean up local temp files on transcode or upload failure | Easy | `videoTranscoder/service.go:76-104` — on failure the downloaded raw file and partial HLS folder are left on disk; add cleanup with `defer` or explicit cleanup in the error paths |
| 16 | Fix `WriteSucces` typo → `WriteSuccess` | Trivial | `internal/json/json.go:26` |
| 19 | Replace wildcard CORS with a configurable allowlist | Easy | `api.go:24` |

---

## Medium — security

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 24 | **Auth on upload endpoint** | Hard | `POST /videos/create-and-get-upload-url` is open to anyone; gate it behind the auth middleware (item 22) |
| 25 | **Switch presigned PUT → presigned POST with `content-length-range`** | Medium | S3 presigned POST supports a policy condition that caps upload size; presigned PUT has no equivalent — S3 will reject oversized uploads at the edge. Requires changing `GenerateRawUploadUrl` in `storage/s3.go` and the upload logic in the frontend |
| 26 | **Enforce `Content-Type` on presigned upload URL** | Easy | Pass `ContentType: "video/mp4"` or `"video/webm"` in `PutObjectInput` so S3 rejects non-video uploads; client must send a matching `Content-Type` header |
| 27 | **Rate limit upload URL generation per user/IP** | Medium | Add a Gin middleware that caps how many upload URLs a single user (or IP before auth) can request per hour; prevents quota abuse by legitimate accounts |
| 28 | **S3 lifecycle policy to expire unprocessed raw files** | Easy | Add a lifecycle rule on the `raw/` prefix to delete objects older than 24h; limits damage from abuse that slips through and keeps storage costs down |

---

## Low — future features

| # | Task | Difficulty | Notes |
|---|------|------------|-------|
| 20 | Multiple transcode quality levels (1080p, 720p, 480p) using HLS adaptive bitrate | Hard | Would produce multiple renditions in the m3u8 manifest — the current fMP4 pipeline is already the right foundation |
| 21 | Chunked / resumable upload protocol | Hard | Design doc already identifies this — useful for large files on bad connections |
| 22 | Auth middleware (JWT or session) | Hard | Required before any production use |
| 23 | Video metadata endpoint (`GET /video/:id`) — title, description, duration, status | Medium | Design doc lists this; schema doesn't have title/description yet (migration needed) |
