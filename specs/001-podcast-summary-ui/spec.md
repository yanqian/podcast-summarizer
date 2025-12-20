# Feature Specification: Podcast Summary UI

**Feature Branch**: `001-podcast-summary-ui`  
**Created**: 2025-11-29  
**Status**: Draft  
**Input**: User description: "I would like to create an application that can read from user input of podcast url and then analyse the video or the text script if there is text script, and generate the text script, and summarise the texts paragraphe by paragraphe. The display shoule be like the left side with original text script, and the right side with the summarized text."

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Submit podcast URL and view summaries (Priority: P1)

As a listener, I want to paste a podcast URL, get the transcript (using an existing script when available), and
see paragraph-by-paragraph summaries beside the original text so I can skim quickly.

**Why this priority**: This is the core value—turn a URL into a readable transcript with aligned summaries.

**Independent Test**: Paste a URL with an accessible transcript and verify the side-by-side transcript and
summaries render with matching paragraph counts and headers.

**Acceptance Scenarios**:

1. **Given** a valid podcast URL with an available transcript, **When** I submit it, **Then** the app displays
   the original transcript on the left and a summary for each paragraph on the right.
2. **Given** a valid podcast URL with transcript, **When** paragraphs exceed the viewport, **Then** the view
   keeps the paragraph alignment (counts stay matched) and supports scrolling without losing context.

---

### User Story 2 - Generate transcript when none exists (Priority: P2)

As a listener, I want the app to create a transcript from the podcast audio/video when no script exists so the
side-by-side summary still works.

**Why this priority**: Many podcasts lack published transcripts; auto-transcription preserves core value.

**Independent Test**: Use a URL without a transcript, wait for processing, and verify the generated transcript
and summaries appear with progress feedback.

**Acceptance Scenarios**:

1. **Given** a valid podcast URL without a transcript, **When** I submit it, **Then** I see processing status
   and receive a generated transcript with aligned summaries once complete.

---

### User Story 3 - Export and share outputs (Priority: P3)

As a listener, I want to copy or download the transcript and summaries so I can use them elsewhere.

**Why this priority**: Sharing/export increases reuse and keeps the experience useful beyond the UI.

**Independent Test**: From a completed summary view, export or copy both texts and confirm the content matches
what is displayed.

**Acceptance Scenarios**:

1. **Given** a completed transcript and summaries, **When** I choose export/copy, **Then** I receive both the
   original text and summaries with paragraph alignment preserved.

---

### Edge Cases

- Invalid or unreachable podcast URL (bad format, 404, geo-blocked content).
- Podcast provides neither transcript nor accessible audio/video stream.
- Extremely long episodes or malformed transcripts causing paragraph detection failures.
- Mixed-language transcripts or missing punctuation leading to poor paragraph segmentation.
- User navigates away or resubmits while transcription/summarization is in progress.

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: System MUST accept a podcast URL input and validate it as reachable content before processing.
- **FR-002**: System MUST detect and ingest an existing transcript or captions from the provided URL when
  available.
- **FR-003**: System MUST generate a transcript from the podcast audio/video when no transcript is available,
  with progress feedback during processing.
- **FR-004**: System MUST segment transcripts into paragraphs with stable identifiers to support alignment.
- **FR-005**: System MUST generate a concise summary for each paragraph that preserves key facts and intent.
- **FR-006**: System MUST render a side-by-side view with original transcript on the left and paragraph-aligned
  summaries on the right, maintaining consistent paragraph counts and ordering.
- **FR-007**: System MUST surface actionable errors for invalid URLs, unsupported media, or processing failures,
  without exposing internal details.
- **FR-008**: System MUST let users copy or download both the transcript and the summaries in a portable text
  format while keeping paragraph order.
- **FR-009**: System MUST allow users to re-run summarization after transcript generation without re-entering
  the URL.
- **FR-010**: System MUST provide status indicators for long-running work (e.g., transcription, summarization)
  and prevent duplicate submissions while a job is active.

### Quality & Engineering Constraints

- **Q-001**: Tests MUST be defined for each story (unit + integration/contract) with a path to ≥85% coverage
  of touched code; note any regression tests added.
- **Q-002**: UX expectations MUST be explicit: CLI/API flags, output schema, error formats, and accessibility
  notes to keep interactions consistent.
- **Q-003**: Performance budgets MUST be measurable (default p95 <200ms for core operations; specify targets
  for long-running/background work) and include how they will be measured.
- **Q-004**: Observability MUST cover structured logs or metrics around external calls and performance-
  critical paths to validate budgets and diagnose issues.

### Key Entities *(include if feature involves data)*

- **Podcast Source**: Input URL plus fetched metadata (title, duration, availability of transcripts/captions).
- **Transcript Paragraph**: Ordered paragraph content with stable identifiers and optional timestamps.
- **Summary Paragraph**: Condensed text linked 1:1 to a Transcript Paragraph.
- **Processing Job**: Status for transcription/summarization tasks with progress and error states.

### Dependencies & Assumptions

- Podcast hosts provide either an accessible transcript, captions, or streamable audio/video for transcription.
- Users have permission to process the provided podcast content for personal use.
- Long-running jobs may rely on background processing but must surface progress to the user.

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: For podcasts with an existing transcript, 95% of requests display side-by-side transcript and
  summaries within 10 seconds of submission.
- **SC-002**: For podcasts without transcripts up to 60 minutes, 90% of requests finish transcription +
  paragraph summaries within 5 minutes with clear progress feedback.
- **SC-003**: At least 95% of transcript paragraphs have corresponding summaries with matched ordering and no
  missing entries on first render.
- **SC-004**: 90% of users in testing report that summaries reduce reading time by at least 50% while still
  understanding key points (qualitative survey or task completion study).
