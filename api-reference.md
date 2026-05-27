# API Reference

This document describes all HTTP endpoints exposed by the Go backend. Intended for use by frontend code or other Claude instances working on the frontend.

## Base URL

Configured via environment. All routes are relative to the server root.

## Global Behavior

All responses are JSON. CORS headers are set on every response, allowing any origin with `GET`, `POST`, and `PUT` methods. Preflight `OPTIONS` requests are handled automatically.

---

## Endpoints

### Health Check

```
GET /health
```

No request body. Returns a simple OK confirmation.

**Response `200 OK`:**
```json
{ "message": "OK" }
```

---

### Create Video and Get Upload URL

```
POST /videos/create-and-get-upload-url
```

Creates a video record in the database with `pending` status and returns a presigned URL the client can use to upload the raw video file directly to storage.

**Request body:**
```json
{
  "title": "string (required)",
  "file_name": "string (required)"
}
```

`file_name` must end in `.mp4` or `.webm`. Any other extension (or no extension) returns a 400.

**Response `201 Created`:**
```json
{
  "id": "<uuid>",
  "upload-url": "<presigned upload URL, valid 1 hour>"
}
```

**Error responses:**
- `400 Bad Request` — missing fields or unsupported file extension

---

### Submit Video Processing Job

```
POST /videos/:id/process
```

Triggers the transcoding pipeline for a video that has already been uploaded. The video must be in `pending` status and the raw file must exist in storage.

**URL params:**
- `id` — UUID of the video

**Request body:** none

**Response `200 OK`:**
```json
{ "message": "processing job submitted" }
```

**Error responses:**
- `400 Bad Request` — invalid UUID format, or video is not in a valid state to start processing
- `404 Not Found` — video not found, or uploaded file not found in storage
- `409 Conflict` — video is not in `pending` status

---

### Get Signed URL for HLS Segment

```
GET /videos/:id/stream/:file/signed-url
```

Returns a short-lived presigned GET URL for a specific HLS segment or playlist file. The video must be in `finished` status.

**URL params:**
- `id` — UUID of the video
- `file` — filename of the HLS asset (e.g. `index.m3u8`, `segment0.ts`). Path separators (`/`, `\`) are rejected.

**Request body:** none

**Response `200 OK`:**
```json
{ "signed_url": "<presigned GET URL, valid 1 hour>" }
```

**Error responses:**
- `400 Bad Request` — invalid UUID, video not finished, or invalid file path
- `404 Not Found` — video not found

---

## Local Storage Only

The following two endpoints are only registered when the server is started with `RUN_LOCAL_STORAGE=true`. They are not available in production S3 mode.

### Upload Raw File (local dev)

```
PUT /videos/upload-raw/:filename
```

Accepts a raw video file upload directly to the server's local storage. Used in place of the presigned S3 upload flow during local development.

**URL params:**
- `filename` — name of the file to save

**Request body:** binary file data (multipart form, max 8 MB)

**Response `200 OK`:** no body

**Error responses:**
- `500 Internal Server Error` — file could not be saved

---

### Serve HLS File (local dev)

```
GET /videos/hls/:id/:file
```

Serves a processed HLS file directly from local storage. Used in place of signed URL redirects during local development.

**URL params:**
- `id` — video ID (path traversal characters rejected)
- `file` — HLS asset filename (path traversal characters rejected)

**Response `200 OK`:** binary file content
- `Content-Type: application/vnd.apple.mpegurl` for `.m3u8` files
- `Content-Type: video/mp4` for all other files

**Error responses:**
- `400 Bad Request` — path traversal attempt in `id` or `file`
- `404 Not Found` — file not found

---

## Error Response Shape

All error responses follow this structure:

```json
{ "error": "human-readable message" }
```

---

## Video Status Values

A video moves through these statuses during its lifecycle:

| Status | Meaning |
|--------|---------|
| `pending` | Created, awaiting raw file upload and processing trigger |
| `processing` | Transcoding job is running |
| `finished` | HLS output is ready; streaming endpoints are available |

Only `finished` videos can have their HLS segments fetched via the signed-URL endpoint.

---

## Typical Frontend Flow

1. `POST /videos/create-and-get-upload-url` — get `id` and `upload-url`
2. `PUT <upload-url>` — upload the raw file directly to storage using the presigned URL
3. `POST /videos/:id/process` — trigger transcoding
4. Poll or wait until video status is `finished`
5. `GET /videos/:id/stream/index.m3u8/signed-url` — get a signed URL for the HLS playlist
6. Feed the signed URL to an HLS-capable video player; player requests individual segments via subsequent calls to `GET /videos/:id/stream/:file/signed-url`
