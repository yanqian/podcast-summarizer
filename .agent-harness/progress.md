# Progress

## Current System Status

Harness is installed and runnable. The project minspec for a local-first podcast summarizer rewrite has been accepted and recorded in `SPEC.md`.

F001 has been implemented and evaluator-accepted. Root `./init.sh` now runs harness verification, backend Go tests, frontend tests/build, and a deterministic local SQLite backend smoke check without live OpenAI credentials.

F002 has been implemented and evaluator-accepted. The implementation removes selectable Postgres, Valkey/Redis, R2, cloud-mode, and non-SQLite runtime paths from project-owned backend code and docs, keeps SQLite/local filesystem as the only runtime path, and adds a backend static test that prevents reintroducing required non-SQLite storage dependencies or env knobs in the default runtime.

F003 has been implemented and evaluator-accepted. The implementation adds SQLite persistence for episodes, processing jobs, audio chunks, transcript segments, summary segments, and transcript-to-summary mappings, with backend integration tests covering URL lookup, normal artifact writes, status transitions, and mapping retrieval.

F004 has been implemented and evaluator-accepted. The implementation adds pre-pipeline duplicate URL detection in backend ingestion, returns existing episode/job state for duplicate submissions, and adds backend tests proving duplicate submissions do not create another processing job.

The rewrite direction is:

- Local-first portfolio/demo application.
- Go backend and React/Vite frontend.
- SQLite-only persistence.
- Local filesystem audio and intermediate artifacts.
- OpenAI as the only real transcription and summarization provider for the initial rewrite.
- Deterministic fixture/stub verification for tests and smoke checks without live OpenAI calls.
- No remote deployment, Postgres, Valkey, Redis, cache services, or cloud mode.

## Last Completed Feature

F004 Add podcast URL ingestion and idempotency.

## Next Feature

F005 Download and chunk podcast audio locally.

## Known Issues

- Real OpenAI processing will require `OPENAI_API_KEY`, but routine tests and smoke checks must not require live credentials.
- Audio chunking may require selecting and documenting a local tool such as `ffmpeg` during F005.
- The Codex provider adapter produced completed coding work and an evaluator pass, but the orchestrator still recorded a non-zero coding-provider failure. Future unattended rounds may need provider adapter hardening if this repeats.
