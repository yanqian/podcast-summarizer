# Podcast Summarizer

Monorepo for a podcast ingestion and summarization app. Backend is Go (clean architecture) with HTTP API, SSE streaming, Postgres, and Valkey; frontend is Vite + React + TypeScript + Tailwind/shadcn/ui.

## Project layout
- `backend/`: Go services, adapters, infra, and jobs; entrypoint at `cmd/server`.
- `frontend/`: Vite SPA, pages/components/hooks/services; talks to the backend API.
- `DEPLOYMENT.md`: Cloud Run + Cloud Build cheatsheet.
- `specs/`: planning artifacts.

## Prerequisites
- Go 1.23+
- Node 20+ and npm
- ffmpeg on PATH (set `FFMPEG_PATH` if not simply `ffmpeg`)

## Backend (Go)
1) Copy `backend/.env.example` to `backend/.env` and fill values (never commit secrets).
2) From `backend/`:
   - Run: `make run` (loads `.env` then `go run ./cmd/server`, defaults to `:8080`)
   - Test: `make test` (or `go test ./...`)
   - Lint: `make lint` (requires golangci-lint installed)

Key env vars (`backend/src/config/config.go`):
- `POSTGRES_URL`, `VALKEY_URL`
- `API_BASE_URL` (public base for building links)
- `FFMPEG_PATH`
- Transcription adapters: `TRANSCRIBE_URL`, `TRANSCRIBE_KEY`
- Summarization adapters: `SUMMARIZE_URL`, `SUMMARIZE_KEY`
- OpenAI: `OPENAI_API_KEY`, `OPENAI_TRANSCRIBE_MODEL`, `OPENAI_SUMMARIZE_MODEL`
- R2/object storage: `R2_ENDPOINT`, `R2_BUCKET`, `R2_ACCESS_KEY`, `R2_SECRET_KEY`, `R2_PUBLIC_BASE_URL`

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
- Deploy to Cloud Run: see commands and CI/CD example in `DEPLOYMENT.md`. Use Secret Manager + `--set-secrets` for keys, and VPC connector if hitting private Postgres/Valkey.

## Contributing / workflow
- Keep backend/frontend changes in sync via the monorepo.
- Favor feature branches, PRs, and CI that runs `npm test && npm run lint` (frontend) plus `go test ./...` (backend).
- Do not commit `.env` files or secrets. 
