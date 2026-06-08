# Progress

## Current System Status

Harness is installed and runnable. The project minspec for a local-first podcast summarizer rewrite has been accepted and recorded in `SPEC.md`.

F001 has been implemented and evaluator-accepted. Root `./init.sh` now runs harness verification, backend Go tests, frontend tests/build, and a deterministic local SQLite backend smoke check without live OpenAI credentials.

The rewrite direction is:

- Local-first portfolio/demo application.
- Go backend and React/Vite frontend.
- SQLite-only persistence.
- Local filesystem audio and intermediate artifacts.
- OpenAI as the only real transcription and summarization provider for the initial rewrite.
- Deterministic fixture/stub verification for tests and smoke checks without live OpenAI calls.
- No remote deployment, Postgres, Valkey, Redis, cache services, or cloud mode.

## Last Completed Feature

F001 Create local-first runnable skeleton.

## Next Feature

F002 Remove cloud and non-SQLite runtime paths.

## Known Issues

- The existing app implementation and docs may still contain obsolete cloud, Postgres, Valkey, cache, or deployment assumptions until F002 is completed.
- Real OpenAI processing will require `OPENAI_API_KEY`, but routine tests and smoke checks must not require live credentials.
- Audio chunking may require selecting and documenting a local tool such as `ffmpeg` during F005.
- The Codex provider adapter produced completed coding work and an evaluator pass, but the orchestrator still recorded a non-zero coding-provider failure. Future unattended rounds may need provider adapter hardening if this repeats.
