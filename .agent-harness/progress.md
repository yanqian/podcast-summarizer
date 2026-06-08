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

The rewrite direction is:

- Local-first portfolio/demo application.
- Go backend and React/Vite frontend.
- SQLite-only persistence.
- Local filesystem audio and intermediate artifacts.
- OpenAI as the only real transcription and summarization provider for the initial rewrite.
- Deterministic fixture/stub verification for tests and smoke checks without live OpenAI calls.
- No remote deployment, Postgres, Valkey, Redis, cache services, or cloud mode.

## Last Completed Feature

F007 Add OpenAI summarization and transcript mapping.

## Next Feature

F008 Orchestrate podcast processing pipeline.

## Known Issues

- Real OpenAI processing will require `OPENAI_API_KEY`, but routine tests and smoke checks must not require live credentials.
- F005 now documents `ffmpeg` as the local chunking tool; routine tests use fixtures and do not require a large podcast file.
- The Codex provider adapter produced completed coding work and an evaluator pass, but the orchestrator still recorded a non-zero coding-provider failure. Future unattended rounds may need provider adapter hardening if this repeats.
- Manual F006 coding encountered a sandbox listener limitation when using `httptest`; adapter verification now uses a fake `http.RoundTripper` fixture, so no socket binding or live OpenAI call is required.
