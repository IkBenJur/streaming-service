# S3 Bucket Implementation Plan

## Context

Replace the direct-upload + local-serve flow with a presigned-URL flow backed by an S3-compatible bucket (Railway). Two prefixes separate concerns:

- `raw/` — original uploads (configure auto-delete on this prefix in Railway)
- `hls/` — processed HLS output (permanent)

**Cost strategy** — bucket egress is free, service egress is not:

| Step | Traffic | Cost |
|---|---|---|
| Client → raw bucket | Presigned PUT, bypasses server | Free |
| Worker ← raw bucket | Server downloads to transcode | Service egress (unavoidable) |
| Worker → hls bucket | Server uploads HLS segments | Service egress (unavoidable) |
| Client ← hls bucket | Presigned GET URLs, bypasses server | Free |

---

## Storage Backends

The implementation uses two backends behind a shared `VideoProcessingService` interface:

- **`LocalStorage`** — keeps working as-is for local development. No presigned URLs; files live on disk; the server streams them directly.
- **`S3Storage`** — new production backend. Uses presigned PUTs for uploads and presigned GETs for playback; raw and HLS files are never stored on disk beyond transcoding.

Which backend is wired up is decided in `cmd/main.go` based on environment variables.

---

## Environment Variables

Add these to Railway (and locally for testing when using S3):

```
S3_ENDPOINT=
S3_ACCESS_KEY_ID=
S3_SECRET_ACCESS_KEY=
S3_BUCKET_NAME=
S3_REGION=auto
```

---

## Step 1 — Redesign the Storage Interface

Update `VideoProcessingService` in `internal/videoProcessing/service.go` so both backends can satisfy it. The key methods and how each backend implements them:

| Method | LocalStorage | S3Storage |
|---|---|---|
| `GenerateRawUploadURL(ctx, key) (string, error)` | Returns `("", nil)` — signals "no presigned upload" | Returns a presigned PUT URL for `raw/{key}` |
| `SaveRawFile(ctx, r io.Reader, key string) error` | Saves to `./files/raw/{key}` | Not needed — client uploads directly via presigned URL |
| `PrepareRawFile(ctx, key) (localPath string, cleanup func(), error)` | Returns the existing local path, no-op cleanup | Downloads `raw/{key}` to a temp file, cleanup deletes it |
| `HLSOutputPath(id string) (string, error)` | Returns `./files/hls/{id}/`, creates the dir | Returns a temp dir for ffmpeg to write into |
| `CommitHLSFiles(ctx, videoID, localDir string) error` | No-op — ffmpeg wrote directly to the final location | Walks `localDir`, uploads each file to `hls/{videoID}/`, then removes the temp dir |
| `GetFile(ctx, key) (io.ReadCloser, error)` | Opens the file from `./files/{key}` | Streams the object from S3 |
| `GeneratePresignedGetURL(ctx, key) (string, error)` | Returns `("", nil)` — signals "no presigned playback" | Returns a presigned GET URL for `key` |

Remove the old `RawFilePath` method — it leaks the local-filesystem assumption. `PrepareRawFile` replaces it.

---

## Step 2 — Add S3 Dependencies

```bash
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
go get github.com/aws/aws-sdk-go-v2/service/s3
go mod tidy
```

---

## Step 3 — Implement `S3Storage`

Create `internal/storage/s3.go`. The struct holds an `s3.Client`, a `s3.PresignClient`, and the bucket name. Implement each method from the table above. Key details:

- Presigned PUTs expire after 1 hour.
- `PrepareRawFile` downloads the object and writes it to `os.CreateTemp`; the returned cleanup func deletes that file.
- `HLSOutputPath` creates a temp dir with `os.MkdirTemp`; `CommitHLSFiles` uploads everything in it and then calls `os.RemoveAll`.
- Use `UsePathStyle = true` and a custom `BaseEndpoint` to target Railway's S3-compatible endpoint.

---

## Step 4 — Update `LocalStorage`

Add the three new interface methods as stubs on `LocalStorage`:

- `PrepareRawFile`: call `SaveRawFile` is no longer needed here — just return the existing local path (`./files/raw/{key}`) and a no-op cleanup func.
- `CommitHLSFiles`: no-op, return nil.
- `GenerateRawUploadURL` and `GeneratePresignedGetURL`: return `("", nil)`.
- Keep `SaveRawFile` — the handler still calls it in the local-dev path.

Remove `RawFilePath` from `LocalStorage` as well.

---

## Step 5 — Update the Transcoder

In `transcodeVideo` (`internal/videoTranscoder/service.go`):

1. Replace `RawFilePath(...)` with `PrepareRawFile(ctx, key)` and defer the returned cleanup.
2. Pass the returned local path to ffmpeg instead of the old filepath.
3. After `cmd.Wait()` succeeds, call `CommitHLSFiles(ctx, videoID, outputPath)`.

---

## Step 6 — Update the Handler

In `internal/videoProcessing/handler.go`:

### `UploadVideo`

Check whether the backend supports presigned uploads:

- If `GenerateRawUploadURL` returns a non-empty URL: parse `{"extension": "mp4"}` from JSON body, create the DB row, return `{"id": ..., "upload_url": ...}`. Do **not** start transcoding yet.
- If it returns `""` (LocalStorage): keep the existing multipart path — save the file via `SaveRawFile`, then call `h.transcoder.Submit(id)` immediately.

### Add `TriggerTranscode`

New endpoint `POST /videos/:id/transcode`. Look up the video, assert it's in `Pending` status, then call `h.transcoder.Submit(id)`. This endpoint is only called in the S3 flow (the client does: presigned PUT → trigger transcode).

### `GetVideoStream`

When serving `playlist.m3u8`:

- For `LocalStorage` (`GeneratePresignedGetURL` returns `""`): serve the file directly, no changes needed.
- For `S3Storage`: read the m3u8, scan line by line, rewrite each segment URI (including `EXT-X-MAP:URI`) to a presigned GET URL, and return the modified content.

For all other files (segments):

- `LocalStorage`: serve via `GetFile` as before.
- `S3Storage`: redirect (`307`) to a presigned GET URL.

---

## Step 7 — Wire It Up in `cmd/`

### `cmd/api.go`

- Add a `Storage VideoProcessingService` field to `Application`.
- Remove the hardcoded `NewLocalStorage` call inside `Mount()` — use `app.Storage` instead.
- Add the new route: `POST /videos/:id/transcode`.

### `cmd/main.go`

Select the backend based on env vars. If `S3_ENDPOINT` is set, build `S3Storage`; otherwise fall back to `LocalStorage`. Pass the chosen backend to both the transcoder and the handler via the `Application` struct.

---

## Upload Flow Summary (S3 path)

```
1. POST /upload-video  {"extension": "mp4"}
   ← {"id": "abc123", "upload_url": "https://...presigned..."}

2. PUT <upload_url>  <raw video bytes>          ← direct to S3, zero service egress

3. POST /videos/abc123/transcode
   ← {"message": "transcoding started"}

4. GET /video-stream/abc123/playlist.m3u8
   ← m3u8 with presigned segment URLs

5. HLS player fetches segments directly from S3  ← zero service egress
```

---

## Verification Checklist

- [ ] `go build ./...` compiles without errors
- [ ] LocalStorage path: multipart upload → immediate transcode → `GetVideoStream` still works
- [ ] S3 path: `POST /upload-video` returns `id` + `upload_url`
- [ ] PUT a test video to the presigned URL
- [ ] `POST /videos/:id/transcode` triggers transcoding
- [ ] Video status progresses to `finished`
- [ ] `GET /video-stream/:id/playlist.m3u8` returns a valid m3u8 with full presigned URLs for segments
- [ ] HLS player loads and plays the video
