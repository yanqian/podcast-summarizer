# Podcast Summarizer

Monorepo for a podcast ingestion and summarization app. Backend is Go (clean architecture) with HTTP API and SSE streaming. It defaults to local SQLite for demo/self-hosted use, with optional Postgres + Valkey for cloud deployments. Frontend is Vite + React + TypeScript + Tailwind/shadcn/ui.

## Project layout
- `backend/`: Go services, adapters, infra, and jobs; entrypoint at `cmd/server`.
- `frontend/`: Vite SPA, pages/components/hooks/services; talks to the backend API.
- `DEPLOYMENT.md`: Cloud Run + Cloud Build cheatsheet.
- `specs/`: planning artifacts.

## Prerequisites
- Go 1.25+
- Node 20+ and npm
- ffmpeg on PATH (set `FFMPEG_PATH` if not simply `ffmpeg`)

## Backend (Go)
1) Copy `backend/.env.example` to `backend/.env` and fill any optional values (never commit secrets).
2) From `backend/`:
   - Run: `make run` (loads `.env` then `go run ./cmd/server`, defaults to `:8080`)
   - Test: `make test` (or `go test ./...`)
   - Lint: `make lint` (requires golangci-lint installed)

Key env vars (`backend/src/config/config.go`):
- `STORAGE_DRIVER`: defaults to `sqlite`; set `postgres` to use Postgres.
- `SQLITE_PATH`: defaults to `data/podcast.db`.
- `POSTGRES_URL`, `VALKEY_URL`: required together for optional Postgres cloud mode. SQLite mode uses the local database for job locking.
- `API_BASE_URL` (public base for building links)
- `FFMPEG_PATH`
- Transcription adapters: `TRANSCRIBE_URL`, `TRANSCRIBE_KEY`
- Summarization adapters: `SUMMARIZE_URL`, `SUMMARIZE_KEY`
- OpenAI: `OPENAI_API_KEY`, `OPENAI_TRANSCRIBE_MODEL`, `OPENAI_SUMMARIZE_MODEL`
- Object storage: `OBJECT_STORAGE_DRIVER` defaults to `local`, storing generated audio/chunk artifacts under `LOCAL_STORAGE_PATH` (`data/storage`). Set `OBJECT_STORAGE_DRIVER=r2` plus `R2_ENDPOINT`, `R2_BUCKET`, `R2_ACCESS_KEY`, `R2_SECRET_KEY`, `R2_PUBLIC_BASE_URL` to use Cloudflare R2.

## Frontend (Vite + React)
From `frontend/`:
- Install deps: `npm ci`
- Dev server: `npm run dev -- --host`
- Type-check + build: `npm run build`
- Lint: `npm run lint`
- Test: `npm test` (Vitest + jsdom)

## Deployment
- Build two images: `backend/Dockerfile` (Go → distroless), `frontend/Dockerfile` (Vite build → nginx).
- Push via Cloud Build: `gcloud builds submit --tag gcr.io/$PROJECT_ID/podcast-api ./backend` and `.../podcast-web ./frontend`.
- Deploy to Cloud Run: see commands and CI/CD example in `DEPLOYMENT.md`. SQLite is best for local/demo runs; Cloud Run's filesystem is ephemeral, so use Postgres mode or attach durable storage before storing data you care about.

## Contributing / workflow
- Keep backend/frontend changes in sync via the monorepo.
- Favor feature branches, PRs, and CI that runs `npm test && npm run lint` (frontend) plus `go test ./...` (backend).
- Do not commit `.env` files or secrets. 
- Do not commit local SQLite database files or generated local storage artifacts; they are ignored via `.gitignore`.
