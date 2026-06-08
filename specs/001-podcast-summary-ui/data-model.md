# Data Model: Podcast Summary UI

Date: 2025-11-29  
Branch: 001-podcast-summary-ui

## Entities

### PodcastSource
- **Fields**: id (uuid), url (string, unique), title (string), description (string), duration_seconds (int),
  has_transcript (bool), created_at (timestamp), updated_at (timestamp)
- **Rules**: url must be reachable and normalized; duration used for SLA expectations.
- **Relationships**: Has many TranscriptParagraph; has many ProcessingJob.

### TranscriptParagraph
- **Fields**: id (uuid), podcast_id (uuid FK), order_index (int), text (string), timestamp_seconds (int?),
  source (enum: provided|generated), created_at (timestamp)
- **Rules**: order_index unique per podcast; text non-empty.
- **Relationships**: Has one SummaryParagraph.

### SummaryParagraph
- **Fields**: id (uuid), transcript_paragraph_id (uuid FK unique), summary_text (string), created_at (timestamp)
- **Rules**: one-to-one with TranscriptParagraph; summary_text non-empty.

### ProcessingJob
- **Fields**: id (uuid), podcast_id (uuid FK), type (enum: ingest|transcribe|summarize), status (enum:
  queued|running|succeeded|failed), started_at (timestamp), completed_at (timestamp?), error_message (string?),
  duration_ms (int?), created_at (timestamp)
- **Rules**: one active job per podcast/type enforced via SQLite lock; status transitions follow queued → running
  → succeeded/failed.
- **Relationships**: Belongs to PodcastSource.

## Validation & Alignment
- Paragraph alignment enforced by shared order_index and transcript_paragraph_id between transcript and summary.
- Jobs must not start summarize before transcript paragraphs exist; re-run summarization allowed without
  re-fetching transcript if already stored.
