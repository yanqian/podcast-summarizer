# Data Model: Podcast Summary UI

Date: 2025-11-29  
Branch: 001-podcast-summary-ui

## Entities

### Episode
- **Fields**: id (uuid), url (string, unique), title (string), description (string), duration_seconds (int),
  audio_url (string?), transcript_url (string?), has_transcript (bool), created_at (timestamp), updated_at (timestamp)
- **Rules**: url must be reachable and normalized; duration used for SLA expectations.
- **Relationships**: Has many TranscriptSegment, SummarySegment, AudioChunk, and ProcessingJob records.

### TranscriptSegment
- **Fields**: id (uuid), episode_id (uuid FK), audio_chunk_id (uuid FK?), order_index (int), text (string),
  start_seconds (float?), end_seconds (float?), provider (string?), model (string?), created_at (timestamp)
- **Rules**: order_index unique per episode; text non-empty.
- **Relationships**: May be linked to one AudioChunk; may map to one or more SummarySegment records.

### SummarySegment
- **Fields**: id (uuid), episode_id (uuid FK), order_index (int), text (string), provider (string?),
  model (string?), created_at (timestamp)
- **Rules**: order_index unique per episode; text non-empty.
- **Relationships**: Maps to source TranscriptSegment records through TranscriptSummaryMapping.

### TranscriptSummaryMapping
- **Fields**: id (uuid), episode_id (uuid FK), summary_segment_id (uuid FK), transcript_segment_id (uuid FK),
  source_order (int), created_at (timestamp)
- **Rules**: source_order preserves source transcript ordering inside a summary group.

### ProcessingJob
- **Fields**: id (uuid), podcast_id (uuid FK), type (enum: ingest|transcribe|summarize), status (enum:
  queued|running|succeeded|failed), started_at (timestamp), completed_at (timestamp?), error_message (string?),
  duration_ms (int?), created_at (timestamp)
- **Rules**: one active job per podcast/type enforced via SQLite lock; status transitions follow queued → running
  → succeeded/failed.
- **Relationships**: Belongs to Episode.

## Validation & Alignment
- Transcript-summary alignment is enforced by transcript_summary_mapping source links and stable order_index values.
- Jobs must not start summarize before transcript segments exist; re-run summarization allowed without
  re-fetching transcript if already stored.
