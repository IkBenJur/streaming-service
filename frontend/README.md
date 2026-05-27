# Streaming Service — Frontend

React Router v7 (SSR) frontend for the streaming service.

## Requirements

- Node.js 26+
- Backend API running (see `VITE_API_URL` below)

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `VITE_API_URL` | Yes | Base URL of the backend API (e.g. `http://localhost:8080`) |

## Development

```bash
npm install
npm run dev
```

Available at `http://localhost:5173`.

## Production (local)

```bash
npm run build
VITE_API_URL=http://localhost:8080 npm run start
```

Available at `http://localhost:3000`.

## Docker

Build and run from the **project root** (`streaming-service/`).

`VITE_API_URL` must be provided at both build time (baked into the client bundle) and run time (used by the SSR server).

```bash
# Build
docker build -f build/Dockerfile.frontend \
  -t streaming-service-frontend \
  frontend/

# Run
docker run --rm -p 3000:3000 \
  -e VITE_API_URL=http://your-backend-host:port \
  streaming-service-frontend
```

If the backend is on localhost, use `host.docker.internal` on Linux:

```bash
docker build -f build/Dockerfile.frontend \
  -t streaming-service-frontend \
  frontend/

docker run --rm -p 3000:3000 \
  --add-host=host.docker.internal:host-gateway \
  -e VITE_API_URL=http://host.docker.internal:8080 \
  streaming-service-frontend
```

Available at `http://localhost:3000`.
