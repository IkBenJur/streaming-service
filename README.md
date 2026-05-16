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
| Go 1.26+ | https://go.dev/dl/ |
| Docker + Docker Compose | https://docs.docker.com/get-docker/ |
| sqlc | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| goose | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| ffmpeg | `sudo pacman -S ffmpeg` or `sudo apt install ffmpeg` |

## Dev Setup

All commands run from the `go/` directory.

**1. Start the database**

```sh
docker compose up -d
```

**2. Run migrations**

```sh
source .env && goose up
```

**3. Install Go dependencies**

```sh
go mod download
```

**4. Start the server**

```sh
go run ./cmd
```

The server starts on `http://localhost:8080`.

## Common Tasks

**Run a database migration**

```sh
source .env && goose up
```

**Roll back the last migration**

```sh
source .env && goose down
```

**Regenerate sqlc queries** (after editing `internal/postgres/sqlc/queries.sql`)

```sh
sqlc generate
```

> Do not manually edit sqlc-generated files (`db.go`, `models.go`, `querier.go`, `queries.sql.go`).

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/upload-video` | Upload a video file (multipart form) |
| GET | `/stream-video/:fileName` | Stream a transcoded video |

## Project Structure

```
go/
├── cmd/               # Entry point and HTTP routing
├── internal/
│   ├── postgres/
│   │   ├── migrations/    # goose migration files
│   │   └── sqlc/          # sqlc-generated code + queries
│   ├── videoProcessing/   # Upload and streaming handlers
│   └── videoTranscoder/   # Async transcoding worker
├── docker-compose.yaml
├── sqlc.yml
└── .env               # goose connection config (not for app secrets)
```
