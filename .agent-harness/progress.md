# Progress

## Current System Status

Harness is installed and runnable. The project minspec for a local-first podcast summarizer rewrite has been accepted and recorded in `SPEC.md`.

F001 has been implemented and evaluator-accepted. Root `./init.sh` now runs harness verification, backend Go tests, frontend tests/build, and a deterministic local SQLite backend smoke check without live OpenAI credentials.

F002 has been implemented and evaluator-accepted. The implementation removes selectable Postgres, Valkey/Redis, R2, cloud-mode, and non-SQLite runtime paths from project-owned backend code and docs, keeps SQLite/local filesystem as the only runtime path, and adds a backend static test that prevents reintroducing required non-SQLite storage dependencies or env knobs in the default runtime.

F003 has been implemented and evaluator-accepted. The implementation adds SQLite persistence for episodes, processing jobs, audio chunks, transcript segments, summary segments, and transcript-to-summary mappings, with backend integration tests covering URL lookup, normal artifact writes, status transitions, and mapping retrieval.

F004 has been implemented and evaluator-accepted. The implementation adds pre-pipeline duplicate URL detection in backend ingestion, returns existing episode/job state for duplicate submissions, and adds backend tests proving duplicate submissions do not create another processing job.

F005 has been implemented and evaluator-accepted. The implementation stores downloaded original audio and ordered chunks under local storage, persists `audio_chunk` metadata with byte size/checksum and stable local file references, propagates download/chunk/storage failures into failed job status, and adds fixture-backed backend tests that do not require a large real podcast file.

F006 has been implemented and evaluator-accepted. The implementation adds a testable OpenAI audio transcription adapter, selects OpenAI by configuration when `OPENAI_API_KEY` is set, preserves no-live-key fixture tests, fails jobs on transcription errors, and persists generated transcript segments with provider/model metadata and local audio chunk association.

F007 has been implemented and evaluator-accepted. The implementation adds grouped transcript-segment summarization through the OpenAI summarizer adapter, deterministic no-live grouping through the configured summary client, persistence of grouped `summary_segment` rows, and source mappings in `transcript_summary_mapping`.

F008 has been implemented and evaluator-accepted. The implementation persists pipeline progress through download, chunking, transcription, summarization, and completion job statuses; fails summarization errors instead of fabricating fallback summaries; and adds a deterministic SQLite-backed backend integration test for fixture-driven end-to-end processing plus idempotent artifact rewrites.

F009 has been implemented and evaluator-accepted. The implementation exposes local backend API detail/status reads for episodes, preserves the existing ingest/list/job endpoints, returns transcript-summary segment mappings for the viewer, includes latest job status and error metadata in episode responses, and adds HTTP contract tests for success, duplicate submission, not found, and failed processing responses.

F010 has been implemented and evaluator-accepted. The implementation adds a frontend admin submission/status page wired to ingest and episode status APIs, distinguishes duplicate existing episodes from new processing jobs, displays queued/completed/failed states and processing errors, refreshes selected episode status, and adds frontend interaction tests for submit, duplicate, loading, success, and error states.

F011 has been implemented and evaluator-accepted. The implementation adds a selected-episode transcript and summary viewer wired to the backend episode detail API, groups transcript segments by summary source mappings, displays unmapped transcript segments, handles loading/empty/failed/completed states, and adds frontend interaction tests with real-shaped fixture API data.

F012 has local Docker demo and documentation implementation artifacts in place but is blocked pending Docker runtime verification. The implementation adds root Docker Compose for backend/frontend, container health checks, a `scripts/verify-docker-demo.sh` verifier, README/runbook updates for deterministic no-key demo mode and portfolio scope, and a static regression test for the local-only Docker demo. Docker CLI/Compose are installed in the current environment, but Docker API access is denied at the active Colima socket, so the required container startup verification could not run.

The rewrite direction is:

- Local-first portfolio/demo application.
- Go backend and React/Vite frontend.
- SQLite-only persistence.
- Local filesystem audio and intermediate artifacts.
- OpenAI as the only real transcription and summarization provider for the initial rewrite.
- Deterministic fixture/stub verification for tests and smoke checks without live OpenAI calls.
- No remote deployment, Postgres, Valkey, Redis, cache services, or cloud mode.

## Last Completed Feature

F011 Build transcript and summary viewer.

## Next Feature

F012 Polish local Docker demo and documentation.

## Known Issues

- Real OpenAI processing will require `OPENAI_API_KEY`, but routine tests and smoke checks must not require live credentials.
- F005 now documents `ffmpeg` as the local chunking tool; routine tests use fixtures and do not require a large podcast file.
- The Codex provider adapter produced completed coding work and an evaluator pass, but the orchestrator still recorded a non-zero coding-provider failure. Future unattended rounds may need provider adapter hardening if this repeats.
- Manual F006 coding encountered a sandbox listener limitation when using `httptest`; adapter verification now uses a fake `http.RoundTripper` fixture, so no socket binding or live OpenAI call is required.
- Manual F010 coding encountered the same sandbox listener limitation for the frontend dev server (`listen EPERM` on `127.0.0.1:5173`) and the in-app Browser surface was unavailable (`iab` not available). Frontend behavior is verified by Vitest interaction tests, lint, TypeScript build, and Vite production build; browser visual inspection remains an environment limitation for this run.
- Manual F011 coding encountered the same sandbox listener limitation for the frontend dev server (`listen EPERM` on `127.0.0.1:5173`). Frontend behavior is verified by Vitest interaction tests, lint, TypeScript build, Vite production build, feature validation, and final root `./init.sh`; browser visual inspection remains an environment limitation for this run.
- Manual F012 coding added the Compose demo workflow, but Docker runtime verification is blocked because the active Colima Docker socket denies access in this environment (`./scripts/verify-docker-demo.sh` exits 1 with Docker API permission denied). Non-Docker verification passes; F012 must remain incomplete until `./scripts/verify-docker-demo.sh` runs successfully in an environment with Docker Compose and Docker API access.
- Manual F012 retry on 2026-06-08 confirmed the same Docker API capability gap: Docker CLI/Compose are installed, root `./init.sh` passes, but `./scripts/verify-docker-demo.sh` still exits 1 before service startup because the active Colima socket denies API access.
