---

description: "Task list for Podcast Summary UI feature"
---

# Tasks: Podcast Summary UI

**Input**: Design documents from `/Users/yanqiang/Ai/podcast-summarizer/specs/001-podcast-summary-ui/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are REQUIRED per constitution—every story needs unit plus integration/contract coverage, and
regressions must be captured with targeted tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions (absolute paths used here)

## Path Conventions

- **Frontend**: `/Users/yanqiang/Ai/podcast-summarizer/frontend/`
- **Backend**: `/Users/yanqiang/Ai/podcast-summarizer/backend/`
- **Tests**: co-located under each layer (unit/integration/contract/e2e)

<!-- 
  ============================================================================
  IMPORTANT: The tasks below are concrete for this feature and grouped by user
  story to allow independent implementation and testing.
  ============================================================================
-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create project structure per plan at /Users/yanqiang/Ai/podcast-summarizer/{backend,frontend}/ with src/, tests/, and config directories
- [X] T002 Initialize frontend (Vite + React + TypeScript + Tailwind + shadcn/ui) at /Users/yanqiang/Ai/podcast-summarizer/frontend/ with base layout and lint/format configs
- [X] T003 Initialize backend Go module with clean architecture scaffolding at /Users/yanqiang/Ai/podcast-summarizer/backend/ and add lint/format tooling (gofmt/golangci-lint)
- [X] T004 Add shared environment templates (.env.example) for SQLite, local storage, and API base URLs at /Users/yanqiang/Ai/podcast-summarizer/
- [X] T005 Configure shared gitignore and tooling configs for frontend/backed outputs at /Users/yanqiang/Ai/podcast-summarizer/.gitignore

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T006 Create current SQLite schema for episode, processing_job, audio_chunk, transcript_segment, summary_segment, transcript_summary_mapping, and job_lock tables at /Users/yanqiang/Ai/podcast-summarizer/backend/src/repo/migrations/001_init.sql
- [X] T007 Implement SQLite duplicate-job locking utilities at /Users/yanqiang/Ai/podcast-summarizer/backend/src/infra/lock/sqlite_lock.go
- [X] T008 Set up HTTP router/middleware (logging, request validation, error mapping) at /Users/yanqiang/Ai/podcast-summarizer/backend/src/api/router.go
- [X] T009 Define domain models and interfaces for repositories and transcription/summarization adapters at /Users/yanqiang/Ai/podcast-summarizer/backend/src/core/contracts.go
- [X] T010 Add observability hooks for external calls and job timing (structured logs/metrics) at /Users/yanqiang/Ai/podcast-summarizer/backend/src/core/observability.go
- [X] T011 Scaffold frontend API client with error handling and base types at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/services/client.ts
- [X] T012 Add job status polling utility with cancellation/backoff at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/services/jobPolling.ts
- [X] T013 Configure automated test runners (vitest/playwright, go test) and CI scripts at /Users/yanqiang/Ai/podcast-summarizer/.github/workflows/ci.yml

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Submit podcast URL and view summaries (Priority: P1) 🎯 MVP

**Goal**: Paste a podcast URL, ingest existing transcript if available, and display paragraph-aligned transcript and summaries side-by-side.

**Independent Test**: Submit a URL with available transcript; verify aligned transcript/summaries render with matching paragraph counts and scrolling preserves alignment.

### Tests for User Story 1 (Required) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation. Include performance checks where
> budgets apply and UX acceptance validation for outputs/errors.**

- [X] T014 [P] [US1] Contract test for /api/podcasts/ingest with existing transcript at /Users/yanqiang/Ai/podcast-summarizer/backend/tests/contract/ingest_existing_test.go
- [X] T015 [P] [US1] Contract test for /api/podcasts/{podcastId}/view alignment response at /Users/yanqiang/Ai/podcast-summarizer/backend/tests/contract/view_alignment_test.go
- [X] T016 [P] [US1] Frontend unit tests for URL input/validation and status messaging at /Users/yanqiang/Ai/podcast-summarizer/frontend/tests/unit/UrlInput.test.tsx
- [X] T017 [P] [US1] Playwright e2e for ingest + side-by-side rendering using mocked transcript at /Users/yanqiang/Ai/podcast-summarizer/frontend/tests/e2e/ingest_existing.spec.ts

### Implementation for User Story 1

- [X] T018 [US1] Implement transcript fetcher adapter (existing transcripts/captions) at /Users/yanqiang/Ai/podcast-summarizer/backend/src/adapters/transcript/fetcher.go
- [X] T019 [US1] Implement ingest service to persist podcast source and paragraphs with alignment at /Users/yanqiang/Ai/podcast-summarizer/backend/src/core/ingest_service.go
- [X] T020 [US1] Implement HTTP handlers for ingest and job acceptance at /Users/yanqiang/Ai/podcast-summarizer/backend/src/api/handlers/ingest.go
- [X] T021 [US1] Implement view endpoint returning transcript + summaries alignment at /Users/yanqiang/Ai/podcast-summarizer/backend/src/api/handlers/view.go
- [X] T022 [US1] Implement SQLite repositories for podcast and paragraph persistence at /Users/yanqiang/Ai/podcast-summarizer/backend/src/repo/
- [X] T023 [US1] Add duplicate request suppression with SQLite-backed locks at /Users/yanqiang/Ai/podcast-summarizer/backend/src/infra/lock/sqlite_lock.go
- [X] T024 [US1] Build side-by-side layout components and page for transcript/summary display at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/pages/PodcastView.tsx
- [X] T025 [US1] Implement frontend API client methods and polling for ingest/view at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/services/podcastClient.ts
- [X] T026 [US1] Add UX states for errors, loading, and alignment mismatch handling at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/components/StatusBanner.tsx
- [X] T027 [US1] Add logging/instrumentation for ingest/view durations at /Users/yanqiang/Ai/podcast-summarizer/backend/src/core/metrics_ingest_view.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Generate transcript when none exists (Priority: P2)

**Goal**: Automatically transcribe podcasts lacking published transcripts and produce aligned summaries with progress feedback.

**Independent Test**: Submit a URL without transcript; verify progress is shown, transcript and summaries appear within SLA, and alignment holds.

### Tests for User Story 2 (Required) ⚠️

- [X] T028 [P] [US2] Contract test for /api/podcasts/ingest triggering transcription path at /Users/yanqiang/Ai/podcast-summarizer/backend/tests/contract/ingest_transcription_test.go
- [X] T029 [P] [US2] Integration test for transcription + summary pipeline with fake adapter at /Users/yanqiang/Ai/podcast-summarizer/backend/tests/integration/transcription_pipeline_test.go
- [X] T030 [P] [US2] Frontend e2e for progress polling and final render when transcription runs at /Users/yanqiang/Ai/podcast-summarizer/frontend/tests/e2e/transcription_progress.spec.ts

### Implementation for User Story 2

- [X] T031 [US2] Implement transcription adapter client and interface at /Users/yanqiang/Ai/podcast-summarizer/backend/src/adapters/transcription/client.go
- [X] T032 [US2] Implement summarize-after-transcribe job workflow with SQLite lock at /Users/yanqiang/Ai/podcast-summarizer/backend/src/core/jobs/transcribe_worker.go
- [X] T033 [US2] Extend job status API with durations/error fields for transcription path at /Users/yanqiang/Ai/podcast-summarizer/backend/src/api/handlers/job_status.go
- [X] T034 [US2] Update frontend polling and progress UI for transcription jobs at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/services/jobPolling.ts
- [X] T035 [US2] Add frontend components for progress indicators and retry affordances at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/components/ProgressIndicator.tsx
- [X] T036 [US2] Add backend logging/metrics for transcription and summarization durations at /Users/yanqiang/Ai/podcast-summarizer/backend/src/core/metrics_transcription.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Export and share outputs (Priority: P3)

**Goal**: Allow users to copy/download transcript and summaries with paragraph order preserved.

**Independent Test**: From a completed view, export/copy both texts and confirm content/ordering matches on download/clipboard.

### Tests for User Story 3 (Required) ⚠️

- [X] T037 [P] [US3] Contract test for /api/podcasts/{podcastId}/export payload at /Users/yanqiang/Ai/podcast-summarizer/backend/tests/contract/export_test.go
- [X] T038 [P] [US3] Frontend unit/e2e tests for copy/download controls at /Users/yanqiang/Ai/podcast-summarizer/frontend/tests/e2e/export_copy.spec.ts

### Implementation for User Story 3

- [X] T039 [US3] Implement export service and handler producing transcript/summary text bundles at /Users/yanqiang/Ai/podcast-summarizer/backend/src/api/handlers/export.go
- [X] T040 [US3] Implement formatter for numbered paragraph outputs at /Users/yanqiang/Ai/podcast-summarizer/backend/src/core/exporter.go
- [X] T041 [US3] Add frontend export/copy UI controls and wiring to API at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/components/ExportControls.tsx
- [X] T042 [US3] Add instrumentation for export latency and errors at /Users/yanqiang/Ai/podcast-summarizer/backend/src/core/metrics_export.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T043 [P] Documentation updates in specs/001-podcast-summary-ui/quickstart.md and README snippets for usage/export
- [ ] T044 Code cleanup and refactoring across frontend/backend with lint/format passes
- [ ] T045 Performance validation against budgets (<200ms UI, <10s existing transcript, <5m transcription) and remediate regressions (placeholder: measure ingest job durations and streaming p95 once real services wired)
- [ ] T046 [P] Additional regression tests for discovered defects in /Users/yanqiang/Ai/podcast-summarizer/{frontend,backend}/tests/
- [X] T047 Security and error-hardening review (input validation, safe error messages) across /Users/yanqiang/Ai/podcast-summarizer/backend/src/
- [X] T048 Run quickstart validation end-to-end and record release verification note in specs/001-podcast-summary-ui/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories proceed in priority order: US1 (P1) → US2 (P2) → US3 (P3)
- **Polish (Final Phase)**: Depends on desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational; no dependencies on other stories
- **User Story 2 (P2)**: Depends on Foundational and US1 data structures/endpoints
- **User Story 3 (P3)**: Depends on Foundational and availability of transcript/summary data from US1/US2

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Performance/UX acceptance tasks accompany implementation and must be validated before closure
- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Dependency Graph

- Phase2 → US1 → US2 → US3 → Polish

### Parallel Opportunities

- Setup tasks T001–T005 can run in parallel where files do not overlap.
- Foundational tasks T006–T013 mostly parallelizable except migrations before repo wiring.
- US1 tests (T014–T017) parallel; frontend layout (T024–T026) can proceed with backend in progress using mocks.
- US2 tests T028–T030 parallel; backend adapter T031 and worker T032 can proceed alongside frontend progress UI T034–T035.
- US3 tests T037–T038 parallel; export handler T039 and formatter T040 can proceed alongside frontend controls T041.

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Contract test for /api/podcasts/ingest with existing transcript at /Users/yanqiang/Ai/podcast-summarizer/backend/tests/contract/ingest_existing_test.go"
Task: "Contract test for /api/podcasts/{podcastId}/view alignment response at /Users/yanqiang/Ai/podcast-summarizer/backend/tests/contract/view_alignment_test.go"
Task: "Frontend unit tests for URL input/validation and status messaging at /Users/yanqiang/Ai/podcast-summarizer/frontend/tests/unit/UrlInput.test.tsx"
Task: "Playwright e2e for ingest + side-by-side rendering using mocked transcript at /Users/yanqiang/Ai/podcast-summarizer/frontend/tests/e2e/ingest_existing.spec.ts"

# Launch all models/services for User Story 1 together where independent:
Task: "Implement transcript fetcher adapter (existing transcripts/captions) at /Users/yanqiang/Ai/podcast-summarizer/backend/src/adapters/transcript/fetcher.go"
Task: "Implement SQLite repositories for podcast and paragraph persistence at /Users/yanqiang/Ai/podcast-summarizer/backend/src/repo/"
Task: "Build side-by-side layout components and page for transcript/summary display at /Users/yanqiang/Ai/podcast-summarizer/frontend/src/pages/PodcastView.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Demo (MVP!)
3. Add User Story 2 → Test independently → Demo
4. Add User Story 3 → Test independently → Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
