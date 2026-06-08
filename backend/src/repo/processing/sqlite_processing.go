package processingrepo

import (
	"context"
	"database/sql"
	"fmt"

	"podcast-summarizer/src/core/domain"
	"podcast-summarizer/src/internal/sqliteutil"
)

type ProcessingSQLiteRepo struct {
	db *sql.DB
}

func NewProcessingSQLiteRepo(db *sql.DB) *ProcessingSQLiteRepo {
	return &ProcessingSQLiteRepo{db: db}
}

func (r *ProcessingSQLiteRepo) SaveAudioChunks(episodeID string, chunks []domain.AudioChunk) ([]domain.AudioChunk, error) {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("sqlite begin audio chunks: %w", err)
	}
	defer rollback(tx)

	const q = `
INSERT INTO audio_chunk (
	id, episode_id, processing_job_id, order_index, file_path, start_seconds, end_seconds,
	duration_seconds, byte_size, checksum
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(episode_id, order_index) DO UPDATE SET
	processing_job_id = excluded.processing_job_id,
	file_path = excluded.file_path,
	start_seconds = excluded.start_seconds,
	end_seconds = excluded.end_seconds,
	duration_seconds = excluded.duration_seconds,
	byte_size = excluded.byte_size,
	checksum = excluded.checksum;`
	saved := make([]domain.AudioChunk, 0, len(chunks))
	for _, chunk := range chunks {
		id := chunk.ID
		if id == "" {
			id = existingIDByOrder(tx, "audio_chunk", episodeID, chunk.OrderIndex)
		}
		if id == "" {
			id = sqliteutil.NewID()
		}
		if _, err := tx.ExecContext(
			context.Background(),
			q,
			id,
			episodeID,
			sqliteutil.NullableString(chunk.ProcessingJobID),
			chunk.OrderIndex,
			chunk.FilePath,
			nullableFloat(chunk.StartSeconds),
			nullableFloat(chunk.EndSeconds),
			nullableFloat(chunk.DurationSeconds),
			nullableInt64(chunk.ByteSize),
			sqliteutil.NullableString(chunk.Checksum),
		); err != nil {
			return nil, fmt.Errorf("sqlite save audio chunk: %w", err)
		}
		chunk.ID = id
		chunk.EpisodeID = episodeID
		saved = append(saved, chunk)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("sqlite commit audio chunks: %w", err)
	}
	return saved, nil
}

func (r *ProcessingSQLiteRepo) ListAudioChunks(episodeID string) ([]domain.AudioChunk, error) {
	const q = `
SELECT id, episode_id, processing_job_id, order_index, file_path, start_seconds, end_seconds,
	duration_seconds, byte_size, checksum, created_at
FROM audio_chunk
WHERE episode_id = ?
ORDER BY order_index;`
	rows, err := r.db.QueryContext(context.Background(), q, episodeID)
	if err != nil {
		return nil, fmt.Errorf("sqlite list audio chunks: %w", err)
	}
	defer closeRows(rows)

	var result []domain.AudioChunk
	for rows.Next() {
		var item domain.AudioChunk
		var jobID sql.NullString
		var start, end, duration sql.NullFloat64
		var byteSize sql.NullInt64
		var checksum sql.NullString
		var createdAt string
		if err := rows.Scan(&item.ID, &item.EpisodeID, &jobID, &item.OrderIndex, &item.FilePath, &start, &end, &duration, &byteSize, &checksum, &createdAt); err != nil {
			return nil, fmt.Errorf("sqlite scan audio chunk: %w", err)
		}
		item.ProcessingJobID = stringPtr(jobID)
		item.StartSeconds = floatPtr(start)
		item.EndSeconds = floatPtr(end)
		item.DurationSeconds = floatPtr(duration)
		item.ByteSize = int64Ptr(byteSize)
		item.Checksum = stringPtr(checksum)
		item.CreatedAt = sqliteutil.ParseTime(createdAt)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *ProcessingSQLiteRepo) SaveTranscriptSegments(episodeID string, segments []domain.TranscriptSegment) ([]domain.TranscriptSegment, error) {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("sqlite begin transcript segments: %w", err)
	}
	defer rollback(tx)

	saved, err := saveTranscriptSegmentsTx(tx, episodeID, segments)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("sqlite commit transcript segments: %w", err)
	}
	return saved, nil
}

func (r *ProcessingSQLiteRepo) ListTranscriptSegments(episodeID string) ([]domain.TranscriptSegment, error) {
	const q = `
SELECT id, episode_id, audio_chunk_id, order_index, start_seconds, end_seconds, text, provider, model, created_at
FROM transcript_segment
WHERE episode_id = ?
ORDER BY order_index;`
	rows, err := r.db.QueryContext(context.Background(), q, episodeID)
	if err != nil {
		return nil, fmt.Errorf("sqlite list transcript segments: %w", err)
	}
	defer closeRows(rows)

	var result []domain.TranscriptSegment
	for rows.Next() {
		item, err := scanTranscriptSegment(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *ProcessingSQLiteRepo) SaveSummarySegments(episodeID string, segments []domain.SummarySegment) ([]domain.SummarySegment, error) {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("sqlite begin summary segments: %w", err)
	}
	defer rollback(tx)

	saved, err := saveSummarySegmentsTx(tx, episodeID, segments)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("sqlite commit summary segments: %w", err)
	}
	return saved, nil
}

func (r *ProcessingSQLiteRepo) ListSummarySegments(episodeID string) ([]domain.SummarySegment, error) {
	const q = `
SELECT id, episode_id, order_index, text, provider, model, created_at
FROM summary_segment
WHERE episode_id = ?
ORDER BY order_index;`
	rows, err := r.db.QueryContext(context.Background(), q, episodeID)
	if err != nil {
		return nil, fmt.Errorf("sqlite list summary segments: %w", err)
	}
	defer closeRows(rows)

	var result []domain.SummarySegment
	for rows.Next() {
		item, err := scanSummarySegment(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *ProcessingSQLiteRepo) SaveTranscriptSummaryMappings(episodeID string, mappings []domain.TranscriptSummaryMapping) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("sqlite begin mappings: %w", err)
	}
	defer rollback(tx)

	if err := replaceTranscriptSummaryMappingsTx(tx, episodeID, mappings); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite commit mappings: %w", err)
	}
	return nil
}

func (r *ProcessingSQLiteRepo) GetSummarySourceMappings(episodeID string) ([]domain.SummarySourceMapping, error) {
	const q = `
SELECT
	ss.id, ss.episode_id, ss.order_index, ss.text, ss.provider, ss.model, ss.created_at,
	ts.id, ts.episode_id, ts.audio_chunk_id, ts.order_index, ts.start_seconds, ts.end_seconds,
	ts.text, ts.provider, ts.model, ts.created_at
FROM summary_segment ss
LEFT JOIN transcript_summary_mapping m ON m.summary_segment_id = ss.id
LEFT JOIN transcript_segment ts ON ts.id = m.transcript_segment_id
WHERE ss.episode_id = ?
ORDER BY ss.order_index, m.source_order, ts.order_index;`
	rows, err := r.db.QueryContext(context.Background(), q, episodeID)
	if err != nil {
		return nil, fmt.Errorf("sqlite get summary mappings: %w", err)
	}
	defer closeRows(rows)

	var result []domain.SummarySourceMapping
	summaryIndex := map[string]int{}
	for rows.Next() {
		var summary domain.SummarySegment
		var provider, model sql.NullString
		var summaryCreatedAt string
		var transcriptID sql.NullString
		var transcriptEpisodeID sql.NullString
		var chunkID sql.NullString
		var transcriptOrder sql.NullInt64
		var start, end sql.NullFloat64
		var transcriptText sql.NullString
		var transcriptProvider, transcriptModel sql.NullString
		var transcriptCreatedAt sql.NullString
		if err := rows.Scan(
			&summary.ID,
			&summary.EpisodeID,
			&summary.OrderIndex,
			&summary.Text,
			&provider,
			&model,
			&summaryCreatedAt,
			&transcriptID,
			&transcriptEpisodeID,
			&chunkID,
			&transcriptOrder,
			&start,
			&end,
			&transcriptText,
			&transcriptProvider,
			&transcriptModel,
			&transcriptCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("sqlite scan summary mapping: %w", err)
		}
		summary.Provider = stringPtr(provider)
		summary.Model = stringPtr(model)
		summary.CreatedAt = sqliteutil.ParseTime(summaryCreatedAt)

		idx, ok := summaryIndex[summary.ID]
		if !ok {
			result = append(result, domain.SummarySourceMapping{Summary: summary})
			idx = len(result) - 1
			summaryIndex[summary.ID] = idx
		}
		if transcriptID.Valid {
			segment := domain.TranscriptSegment{
				ID:           transcriptID.String,
				EpisodeID:    transcriptEpisodeID.String,
				AudioChunkID: stringPtr(chunkID),
				OrderIndex:   int(transcriptOrder.Int64),
				StartSeconds: floatPtr(start),
				EndSeconds:   floatPtr(end),
				Text:         transcriptText.String,
				Provider:     stringPtr(transcriptProvider),
				Model:        stringPtr(transcriptModel),
				CreatedAt:    sqliteutil.ParseTime(transcriptCreatedAt.String),
			}
			result[idx].TranscriptSegments = append(result[idx].TranscriptSegments, segment)
		}
	}
	return result, rows.Err()
}

func saveTranscriptSegmentsTx(tx *sql.Tx, episodeID string, segments []domain.TranscriptSegment) ([]domain.TranscriptSegment, error) {
	const q = `
INSERT INTO transcript_segment (
	id, episode_id, audio_chunk_id, order_index, start_seconds, end_seconds, text, provider, model
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(episode_id, order_index) DO UPDATE SET
	audio_chunk_id = excluded.audio_chunk_id,
	start_seconds = excluded.start_seconds,
	end_seconds = excluded.end_seconds,
	text = excluded.text,
	provider = excluded.provider,
	model = excluded.model;`
	saved := make([]domain.TranscriptSegment, 0, len(segments))
	for _, segment := range segments {
		id := segment.ID
		if id == "" {
			id = existingIDByOrder(tx, "transcript_segment", episodeID, segment.OrderIndex)
		}
		if id == "" {
			id = sqliteutil.NewID()
		}
		if _, err := tx.ExecContext(
			context.Background(),
			q,
			id,
			episodeID,
			sqliteutil.NullableString(segment.AudioChunkID),
			segment.OrderIndex,
			nullableFloat(segment.StartSeconds),
			nullableFloat(segment.EndSeconds),
			segment.Text,
			sqliteutil.NullableString(segment.Provider),
			sqliteutil.NullableString(segment.Model),
		); err != nil {
			return nil, fmt.Errorf("sqlite save transcript segment: %w", err)
		}
		segment.ID = id
		segment.EpisodeID = episodeID
		saved = append(saved, segment)
	}
	return saved, nil
}

func saveSummarySegmentsTx(tx *sql.Tx, episodeID string, segments []domain.SummarySegment) ([]domain.SummarySegment, error) {
	const q = `
INSERT INTO summary_segment (id, episode_id, order_index, text, provider, model)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(episode_id, order_index) DO UPDATE SET
	text = excluded.text,
	provider = excluded.provider,
	model = excluded.model;`
	saved := make([]domain.SummarySegment, 0, len(segments))
	for _, segment := range segments {
		id := segment.ID
		if id == "" {
			id = existingIDByOrder(tx, "summary_segment", episodeID, segment.OrderIndex)
		}
		if id == "" {
			id = sqliteutil.NewID()
		}
		if _, err := tx.ExecContext(
			context.Background(),
			q,
			id,
			episodeID,
			segment.OrderIndex,
			segment.Text,
			sqliteutil.NullableString(segment.Provider),
			sqliteutil.NullableString(segment.Model),
		); err != nil {
			return nil, fmt.Errorf("sqlite save summary segment: %w", err)
		}
		segment.ID = id
		segment.EpisodeID = episodeID
		saved = append(saved, segment)
	}
	return saved, nil
}

func replaceTranscriptSummaryMappingsTx(tx *sql.Tx, episodeID string, mappings []domain.TranscriptSummaryMapping) error {
	if _, err := tx.ExecContext(context.Background(), `DELETE FROM transcript_summary_mapping WHERE episode_id = ?`, episodeID); err != nil {
		return fmt.Errorf("sqlite delete mappings: %w", err)
	}
	const q = `
INSERT INTO transcript_summary_mapping (id, episode_id, summary_segment_id, transcript_segment_id, source_order)
VALUES (?, ?, ?, ?, ?);`
	for _, mapping := range mappings {
		id := mapping.ID
		if id == "" {
			id = sqliteutil.NewID()
		}
		if _, err := tx.ExecContext(context.Background(), q, id, episodeID, mapping.SummarySegmentID, mapping.TranscriptSegmentID, mapping.SourceOrder); err != nil {
			return fmt.Errorf("sqlite save mapping: %w", err)
		}
	}
	return nil
}

func existingIDByOrder(tx *sql.Tx, table, episodeID string, orderIndex int) string {
	var id string
	q := fmt.Sprintf("SELECT id FROM %s WHERE episode_id = ? AND order_index = ?", table)
	if err := tx.QueryRowContext(context.Background(), q, episodeID, orderIndex).Scan(&id); err == nil {
		return id
	}
	return ""
}

func scanTranscriptSegment(rows *sql.Rows) (domain.TranscriptSegment, error) {
	var item domain.TranscriptSegment
	var chunkID, provider, model sql.NullString
	var start, end sql.NullFloat64
	var createdAt string
	if err := rows.Scan(&item.ID, &item.EpisodeID, &chunkID, &item.OrderIndex, &start, &end, &item.Text, &provider, &model, &createdAt); err != nil {
		return domain.TranscriptSegment{}, fmt.Errorf("sqlite scan transcript segment: %w", err)
	}
	item.AudioChunkID = stringPtr(chunkID)
	item.StartSeconds = floatPtr(start)
	item.EndSeconds = floatPtr(end)
	item.Provider = stringPtr(provider)
	item.Model = stringPtr(model)
	item.CreatedAt = sqliteutil.ParseTime(createdAt)
	return item, nil
}

func scanSummarySegment(rows *sql.Rows) (domain.SummarySegment, error) {
	var item domain.SummarySegment
	var provider, model sql.NullString
	var createdAt string
	if err := rows.Scan(&item.ID, &item.EpisodeID, &item.OrderIndex, &item.Text, &provider, &model, &createdAt); err != nil {
		return domain.SummarySegment{}, fmt.Errorf("sqlite scan summary segment: %w", err)
	}
	item.Provider = stringPtr(provider)
	item.Model = stringPtr(model)
	item.CreatedAt = sqliteutil.ParseTime(createdAt)
	return item, nil
}

func nullableFloat(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func stringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func floatPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	return &v.Float64
}

func int64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func closeRows(rows *sql.Rows) {
	_ = rows.Close()
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}
