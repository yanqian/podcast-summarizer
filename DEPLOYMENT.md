# Running and Demoing

This project is optimized as a local-first portfolio demo. The setup uses SQLite and local file storage, so a reviewer can run the full app on one machine without remote services.

The project is not packaged as a hosted service. The intended review paths are:

- `./init.sh` for deterministic recovery verification.
- Local backend/frontend dev servers for interactive development.
- Docker Compose for a two-container demo on one machine.

## Recommended Demo Path

Use this when showing the project on a resume, in an interview, or in a short screen recording.

0. Verify the checkout first:

   ```bash
   ./init.sh
   ```

   This runs harness checks, backend tests, frontend tests/build, and a local backend smoke check without live OpenAI credentials.

1. Start the backend:

   ```bash
   cd backend
   cp .env.example .env
   make run
   ```

2. Start the frontend:

   ```bash
   cd frontend
   npm ci
   npm run dev -- --host
   ```

3. Open the Vite dev URL and submit a podcast URL.

4. Show these behaviors:
   - Ingest accepts a podcast URL and creates a processing job.
   - Transcript progress streams into the UI through SSE.
   - Summary paragraphs align with transcript paragraphs.
   - The podcast list and detail view persist across restarts because data is stored in SQLite.
   - Export endpoints return reusable transcript/summary output.

Local data lives under `backend/data/` and is ignored by git.
Generated audio, chunks, and the SQLite database remain local to that directory unless overridden by environment variables.

## Local Configuration

The backend defaults are enough for a demo:

```text
SQLITE_PATH=data/podcast.db
LOCAL_STORAGE_PATH=data/storage
FFMPEG_PATH=ffmpeg
```

Optional model adapters:

```text
OPENAI_API_KEY=
OPENAI_TRANSCRIBE_MODEL=whisper-1
OPENAI_SUMMARIZE_MODEL=gpt-4o-mini
TRANSCRIBE_URL=
TRANSCRIBE_KEY=
SUMMARIZE_URL=
SUMMARIZE_KEY=
```

If `OPENAI_API_KEY` is unset, routine verification stays deterministic and avoids live OpenAI calls. Backend tests and `./init.sh` use local fixtures or stub adapters; submitted real podcast URLs can still exercise ingestion and status handling, but real transcription and summarization require a configured OpenAI key.

## Docker Demo

Docker is useful when you want a packaged demo, but it is not required for normal development. The Compose file starts only the local backend and frontend.

Run the full Docker demo verification:

```bash
./scripts/verify-docker-demo.sh --keep-running
```

Open:

- Frontend: `http://localhost:8081`
- Backend health: `http://localhost:8080/health`

Stop the demo when finished:

```bash
docker compose down
```

The script verifies:

- backend health returns `{"status":"ok"}`;
- backend podcast list returns a JSON `items` array;
- frontend health returns `ok`;
- the containers are started from the checked-in Compose file.

The backend service mounts `./backend/data:/app/data`, so `podcast.db` and generated media artifacts persist across container restarts. `VITE_API_BASE_URL` is baked into the frontend image as `http://localhost:8080`, which is the backend URL reachable from the browser during the local demo.

To run without the verifier:

```bash
docker compose up --build
```

## Portfolio Notes

For a resume or project page, emphasize the engineering choices rather than infrastructure plumbing:

- Clean architecture in Go with repository and adapter boundaries.
- SQLite as the runtime database because the project is easy to run and inspect.
- SSE streaming from backend jobs to the React UI.
- Media pipeline integration around download, `ffmpeg` chunking, transcription, summarization, and export.
- Non-production scope: this is a single-machine portfolio app, not a multi-user hosted service.

Suggested demo assets:

- A short screen recording of ingest -> streaming transcript -> summary view.
- A screenshot of the podcast list after restarting the backend to show persistence.
- A small architecture diagram or the Mermaid diagram from `README.md`.

## Checks

Full deterministic recovery check:

```bash
./init.sh
```

Docker demo check:

```bash
./scripts/verify-docker-demo.sh
```

Backend:

```bash
cd backend
GOCACHE=$(pwd)/.gocache go test ./...
```

Frontend:

```bash
cd frontend
npm run build
npm test
npm run lint
```
