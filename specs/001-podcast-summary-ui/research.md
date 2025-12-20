# Research: Podcast Summary UI

Date: 2025-11-29  
Branch: 001-podcast-summary-ui  
Spec: specs/001-podcast-summary-ui/spec.md

## Decisions

### Transcript acquisition fallback order
- **Decision**: Attempt published transcripts/captions first; if unavailable, fetch audio/video stream and run
  transcription; if stream is inaccessible, fail with actionable error.
- **Rationale**: Preserves fastest path (<10s) when text exists and provides clear failure when media is not
  reachable.
- **Alternatives considered**: (1) Always transcribe even when transcripts exist (slower, redundant); (2) Only
  support transcripts (excludes many podcasts).

### Paragraph alignment strategy
- **Decision**: Normalize transcript into paragraph array with stable IDs and optional timestamps; summaries
  link 1:1 by paragraph ID; store both in Postgres with ordering; cache assembled pairs in Valkey.
- **Rationale**: Guarantees consistent left/right alignment and enables re-run of summaries without refetching
  text.
- **Alternatives considered**: (1) Sentence-level alignment (more granular but harder to display cleanly);
  (2) page-based chunking (breaks paragraph semantics).

### Long-running job handling
- **Decision**: Submit ingestion as a job with statuses (queued, running, succeeded, failed); expose status API
  polled by frontend; prevent duplicate active jobs per URL using Valkey locks.
- **Rationale**: Meets UX requirement for progress, avoids duplicate processing, supports 5-minute SLA for
  transcription + summarization.
- **Alternatives considered**: (1) Synchronous request (would timeout for transcription); (2) Client-side only
  polling without server state (loses durability).

### Export format
- **Decision**: Provide copy-to-clipboard and downloadable UTF-8 text files for transcript and summaries (two
  files) with paragraph numbering.
- **Rationale**: Keeps implementation simple, portable, and aligned with paragraph structure.
- **Alternatives considered**: (1) PDF export (more complex layout work); (2) CSV (harder to read multiline
  paragraphs).

### Performance instrumentation
- **Decision**: Log timings for transcript fetch/transcription/summarization and cache hits; expose duration
  fields in job status; collect frontend render timing (e.g., interaction timing) for p95 checks.
- **Rationale**: Enables validation of constitution performance budgets and quick regression detection.
- **Alternatives considered**: (1) No explicit timings (violates performance principle); (2) sampling only
  after release (delays detection).

## Resolved Clarifications
- No open clarifications; defaults above cover processing and export behavior.
