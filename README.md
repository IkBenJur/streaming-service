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

The app reads `GOOSE_DBSTRING` at startup. The other `GOOSE_*` variables are only needed when using the goose CLI directly (e.g. for rollbacks).

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
