# Implementation Plan: Podcast Summary UI

**Branch**: `001-podcast-summary-ui` | **Date**: 2025-11-29 | **Spec**: specs/001-podcast-summary-ui/spec.md
**Input**: Feature specification from `/specs/001-podcast-summary-ui/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command.

## Summary

Build a side-by-side podcast transcript and paragraph-level summary experience from a user-supplied podcast
URL. When transcripts are available, ingest and align them; when absent, generate transcripts from audio and
summaries for each paragraph. Provide export/copy, error handling, and status/progress for long-running jobs.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: TypeScript (Vite + React) for frontend; Go (latest stable) for backend (clean architecture).  
**Primary Dependencies**: Vite, React, TypeScript, Tailwind, shadcn/ui on frontend; Go stdlib + HTTP router, streaming client, summarization/transcription adapters; database/sql with pure-Go SQLite for storage.  
**Storage**: SQLite for persisted podcast source metadata, transcript paragraphs, summaries, job status, and single-process job locking.  
**Testing**: Frontend: vitest + @testing-library/react + playwright for integration; Backend: go test with table-driven unit tests, lightweight SQLite/default-mode integration, and contract tests for HTTP handlers.  
**Target Platform**: Web frontend; backend services on Linux container/runtime.  
**Project Type**: Web application with separate frontend and backend.  
**Performance Goals**: UI core interactions p95 <200ms (navigation, render, copy/export); transcript retrieval/summarization status visible, with 95% of existing-transcript flows <10s end-to-end; auto-transcription + summarization for ≤60m episodes complete within 5 minutes for 90% of cases; no >5% regression without exception.  
**Constraints**: Maintain paragraph alignment; avoid duplicate concurrent jobs for same URL in a single-process demo; graceful degradation when transcripts unavailable; accessibility for layout/text; structured logs for external calls and performance hotspots.  
**Scale/Scope**: Initial scope single-tenant app; anticipate tens of concurrent users, batch of transcripts per day; paragraphs per transcript up to low thousands.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Code Quality: frontend and backend separated; clean architecture on backend to isolate transport/core;
  lint/formatters (eslint/prettier, gofmt/golangci-lint) and dependency justification required.
- Testing Discipline: plan includes unit + integration/contract + regression tests; coverage ≥85% of touched
  code; playwright for UX validation; backend integration uses SQLite mode.
- UX Consistency: define standard UI layout (left transcript/right summary), predictable error states, copy/
  export affordances, and progress indicators; acceptance notes captured in specs/tasks.
- Performance & Efficiency: budgets set above (<200ms UI interactions; <10s with existing transcripts; <5m
  for transcription+summary), instrument external calls and job durations; use SQLite-backed locking for duplicate suppression.
- Definition of Done: logs/observability around external calls and job timings, documentation updates (UI
  usage/export), release verification notes on tests, UX validation, performance checks.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
backend/
├── src/
│   ├── api/              # HTTP handlers, request/response DTOs
│   ├── core/             # use cases, services (transcription, summarization)
│   ├── adapters/         # transcript fetchers, summarizer/transcriber clients
│   ├── repo/             # SQLite implementations
│   ├── infra/lock        # SQLite job locking
│   └── config/
└── tests/
    ├── unit/
    ├── integration/      # exercises SQLite/local runtime
    └── contract/         # HTTP contract tests

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   ├── hooks/
│   ├── services/         # API client
│   └── styles/
└── tests/
    ├── unit/
    ├── integration/
    └── e2e/              # playwright
```

**Structure Decision**: Web application with separate frontend and backend folders per stack above; tests
mirrored in each layer for unit/integration/contract/e2e coverage.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
