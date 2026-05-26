# Streaming Service

A video upload and streaming API built with Go. Uploaded videos are transcoded asynchronously and served as streams.

## Tech Stack

- **Go** — application server
- **Gin** — HTTP router
- **PostgreSQL** — metadata storage
- **sqlc** — type-safe SQL query generation
- **goose** — database migrations
- **ffmpeg** — video transcoding

## Prerequisites

Install the following tools before starting:

| Tool | Install |
|------|---------|
| Go 1.26.2+ | https://go.dev/dl/ |
| Docker + Docker Compose | https://docs.docker.com/get-docker/ |
| sqlc | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| goose | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| ffmpeg | `sudo pacman -S ffmpeg` or `sudo apt install ffmpeg` |

## Environment Variables

All commands run from the `go/` directory. Copy the defaults into a local `.env`:

```sh
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `GOOSE_DBSTRING` | `host=localhost user=postgres password=postgres dbname=video-stream sslmode=disable` | PostgreSQL connection string — used by both the app and the goose CLI |
| `GOOSE_DRIVER` | `postgres` | goose driver — only needed for CLI usage |
| `GOOSE_MIGRATION_DIR` | `./internal/postgres/migrations` | Migration directory — only needed for CLI usage |
| `PORT` | `8080` | Port the HTTP server listens on |
| `TRANCODE_JOB_NUMBER_OF_WORKERS` | `2` | Number of concurrent transcoding workers |
| `RUN_LOCAL_STORAGE` | `false` | Set to `true` to skip S3 and use local file storage instead |
| `AWS_ENDPOINT_URL` | — | S3-compatible endpoint (e.g. `https://t3.storageapi.dev` for Railway) |
| `AWS_S3_BUCKET_NAME` | — | Name of the S3 bucket |
| `AWS_DEFAULT_REGION` | — | Region — use `auto` for Railway/Tigris |
| `AWS_ACCESS_KEY_ID` | — | S3 access key |
| `AWS_SECRET_ACCESS_KEY` | — | S3 secret key |

The app reads `GOOSE_DBSTRING` at startup. The other `GOOSE_*` variables are only needed when using the goose CLI directly (e.g. for rollbacks). S3 variables are required when `RUN_LOCAL_STORAGE=false`.

## Dev Setup

All commands run from the `go/` directory.

**1. Start the database**

```sh
docker compose up -d
```

**2. Start the server**

Migrations run automatically on startup.

```sh
source .env && go run ./cmd
```

The server starts on `http://localhost:8080`.

## Docker

All commands run from the repo root (`streaming-service/`).

### Backend

**Build**

```sh
docker build -f build/Dockerfile.backend -t streaming-service-backend .
```

**Run**

```sh
docker run --rm \
  -p 8080:8080 \
  --network go_default \
  -e GOOSE_DBSTRING="host=video-stream-postgres user=postgres password=postgres dbname=video-stream sslmode=disable" \
  -v $(pwd)/files:/app/files \
  streaming-service-backend
```

The backend container must be on the same Docker network as Postgres. `go_default` is the network created by `docker compose up` when run from the `go/` directory — verify with `docker network ls` if the name differs.

### Frontend

**Build**

```sh
docker build -f build/Dockerfile.frontend -t streaming-service-frontend .
```

**Run**

```sh
docker run --rm \
  -p 80:80 \
  -e BACKEND_URL="http://localhost:8080" \
  streaming-service-frontend
```

Set `BACKEND_URL` to wherever the backend is reachable from the browser.

## Common Tasks

**Roll back the last migration**

```sh
source .env && goose down
```

**Regenerate sqlc queries** (after editing `internal/postgres/sqlc/queries.sql`)

```sh
sqlc generate
```

> Do not manually edit sqlc-generated files (`db.go`, `models.go`, `querier.go`, `queries.sql.go`).

## Object Storage (Railway / Tigris)

Videos are stored in an S3-compatible bucket. Railway uses [Tigris](https://www.tigrisdata.com/) under the hood — credentials are available in the Railway dashboard under the bucket's **Settings** tab.

Railway has no built-in UI for browsing bucket contents. Use the AWS CLI instead:

```sh
# Install (Arch)
sudo pacman -S aws-cli-v2

# List all objects
source go/.env && aws s3 ls s3://$AWS_S3_BUCKET_NAME/ --endpoint-url $AWS_ENDPOINT_URL --recursive
```

To upload a file directly via a presigned URL (useful for testing):

```sh
./upload_file_to_url.sh "<presigned-url>"
```

Get a presigned URL from the API:

```sh
curl -s -X POST http://localhost:8080/videos/create-and-get-upload-url \
  -H "Content-Type: application/json" \
  -d '{"title": "my video", "file_name": "video.mp4"}' \
  | jq -r '."upload-url"'
```

## API

### Core endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/videos/create-and-get-upload-url` | Create a video entry and return a presigned S3 upload URL |
| POST | `/videos/:id/process` | Trigger transcoding for a video after it has been uploaded |
| GET | `/videos/:id/stream/:file/signed-url` | Get a short-lived presigned URL for a HLS segment or playlist file |

### Local storage only (`RUN_LOCAL_STORAGE=true`)

| Method | Path | Description |
|--------|------|-------------|
| PUT | `/videos/upload-raw/:filename` | Upload a raw video file directly to the server |
| GET | `/videos/hls/:id/:file` | Serve a HLS segment or playlist file from local disk |

### Upload flow

1. `POST /videos/create-and-get-upload-url` — create a DB entry and get a presigned PUT URL.  Body: `{ "title": "...", "file_name": "video.mp4" }`.  Response includes `id` and `upload-url`.
2. Upload the file to the presigned URL (e.g. with `curl -X PUT`).
3. `POST /videos/:id/process` — submit the transcoding job once the upload is complete.

## Project Structure

```
go/
├── cmd/               # Entry point and HTTP routing
├── internal/
│   ├── env/               # Environment variable loading
│   ├── json/              # JSON response helpers
│   ├── postgres/
│   │   ├── migrations/    # goose migration files
│   │   └── sqlc/          # sqlc-generated code + queries
│   ├── storage/           # StorageClient interface, S3 + local implementations, HTTP handlers
│   └── videoTranscoder/   # Async transcoding worker
├── docker-compose.yaml
├── sqlc.yml
└── .env               # goose connection config (not for app secrets)
```
