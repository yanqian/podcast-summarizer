# Running and Demoing

This project is optimized as a local-first portfolio demo. The setup uses SQLite and local file storage, so a reviewer can run the full app on one machine.

## Recommended Demo Path

Use this when showing the project on a resume, in an interview, or in a short screen recording.

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

If these are unset, the app uses local/stub adapters where available, which keeps the demo self-contained.

## Docker Demo

Docker is useful when you want a packaged demo, but it is not required for normal development.

Build the backend image:

```bash
docker build backend -t podcast-api:local
```

Run it with a mounted data directory:

```bash
mkdir -p backend/data
docker run --rm \
	-p 8080:8080 \
	-v "$PWD/backend/data:/app/data" \
	--env SQLITE_PATH=data/podcast.db \
	--env LOCAL_STORAGE_PATH=data/storage \
	podcast-api:local
```

Build the frontend image. `VITE_API_BASE_URL` is baked into the static bundle at build time:

```bash
docker build frontend \
  --build-arg VITE_API_BASE_URL=http://localhost:8080 \
  -t podcast-web:local
```

Run the frontend:

```bash
docker run --rm -p 8081:8080 podcast-web:local
```

Open `http://localhost:8081`.

## Portfolio Notes

For a resume or project page, emphasize the engineering choices rather than infrastructure plumbing:

- Clean architecture in Go with repository and adapter boundaries.
- SQLite as the runtime database because the project is easy to run and inspect.
- SSE streaming from backend jobs to the React UI.
- Media pipeline integration around download, `ffmpeg` chunking, transcription, summarization, and export.

Suggested demo assets:

- A short screen recording of ingest -> streaming transcript -> summary view.
- A screenshot of the podcast list after restarting the backend to show persistence.
- A small architecture diagram or the Mermaid diagram from `README.md`.

## Checks

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
