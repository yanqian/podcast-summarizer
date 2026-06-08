package paragraphrepo

import (
	"context"
	"database/sql"
	"fmt"

	"podcast-summarizer/src/core/domain"
	"podcast-summarizer/src/internal/sqliteutil"
)

type ParagraphSQLiteRepo struct {
	db *sql.DB
}

func NewParagraphSQLiteRepo(db *sql.DB) *ParagraphSQLiteRepo {
	return &ParagraphSQLiteRepo{db: db}
}

func (r *ParagraphSQLiteRepo) SaveTranscript(podcastID string, paragraphs []domain.Paragraph) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("sqlite begin transcript: %w", err)
	}
	defer rollback(tx)

	const q = `
INSERT INTO transcript_paragraph (id, podcast_id, order_index, text, source)
VALUES (?, ?, ?, ?, 'generated')
ON CONFLICT(podcast_id, order_index) DO UPDATE SET text = excluded.text;`
	const segmentQ = `
INSERT INTO transcript_segment (id, episode_id, order_index, text)
VALUES (?, ?, ?, ?)
ON CONFLICT(episode_id, order_index) DO UPDATE SET text = excluded.text;`
	for _, p := range paragraphs {
		if _, err := tx.ExecContext(context.Background(), q, sqliteutil.NewID(), podcastID, p.OrderIndex, p.Text); err != nil {
			return fmt.Errorf("sqlite save transcript: %w", err)
		}
		segmentID := existingSegmentID(tx, "transcript_segment", podcastID, p.OrderIndex)
		if segmentID == "" {
			segmentID = sqliteutil.NewID()
		}
		if _, err := tx.ExecContext(context.Background(), segmentQ, segmentID, podcastID, p.OrderIndex, p.Text); err != nil {
			return fmt.Errorf("sqlite save transcript segment: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite commit transcript: %w", err)
	}
	return nil
}

func (r *ParagraphSQLiteRepo) SaveSummaries(podcastID string, summaries []domain.Summary) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("sqlite begin summaries: %w", err)
	}
	defer rollback(tx)

	const q = `
INSERT INTO summary_paragraph (id, transcript_paragraph_id, summary_text)
SELECT ?, tp.id, ?
FROM transcript_paragraph tp
WHERE tp.podcast_id = ? AND tp.order_index = ?
ON CONFLICT(transcript_paragraph_id) DO UPDATE SET summary_text = excluded.summary_text;`
	const summaryQ = `
INSERT INTO summary_segment (id, episode_id, order_index, text)
VALUES (?, ?, ?, ?)
ON CONFLICT(episode_id, order_index) DO UPDATE SET text = excluded.text;`
	const mappingQ = `
INSERT INTO transcript_summary_mapping (id, episode_id, summary_segment_id, transcript_segment_id, source_order)
VALUES (?, ?, ?, ?, 1)
ON CONFLICT(summary_segment_id, transcript_segment_id) DO UPDATE SET source_order = excluded.source_order;`
	for _, summary := range summaries {
		if _, err := tx.ExecContext(context.Background(), q, sqliteutil.NewID(), summary.Text, podcastID, summary.OrderIndex); err != nil {
			return fmt.Errorf("sqlite save summaries: %w", err)
		}
		summaryID := existingSegmentID(tx, "summary_segment", podcastID, summary.OrderIndex)
		if summaryID == "" {
			summaryID = sqliteutil.NewID()
		}
		if _, err := tx.ExecContext(context.Background(), summaryQ, summaryID, podcastID, summary.OrderIndex, summary.Text); err != nil {
			return fmt.Errorf("sqlite save summary segment: %w", err)
		}
		transcriptID := existingSegmentID(tx, "transcript_segment", podcastID, summary.OrderIndex)
		if transcriptID != "" {
			mappingID := existingMappingID(tx, summaryID, transcriptID)
			if mappingID == "" {
				mappingID = sqliteutil.NewID()
			}
			if _, err := tx.ExecContext(context.Background(), mappingQ, mappingID, podcastID, summaryID, transcriptID); err != nil {
				return fmt.Errorf("sqlite save summary mapping: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite commit summaries: %w", err)
	}
	return nil
}

func (r *ParagraphSQLiteRepo) GetAligned(podcastID string) ([]domain.ParagraphWithSummary, error) {
	const q = `
SELECT tp.order_index, tp.text, COALESCE(sp.summary_text, '')
FROM transcript_paragraph tp
LEFT JOIN summary_paragraph sp ON sp.transcript_paragraph_id = tp.id
WHERE tp.podcast_id = ?
ORDER BY tp.order_index;`

	rows, err := r.db.QueryContext(context.Background(), q, podcastID)
	if err != nil {
		return nil, fmt.Errorf("sqlite get aligned: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var result []domain.ParagraphWithSummary
	for rows.Next() {
		var item domain.ParagraphWithSummary
		if err := rows.Scan(&item.OrderIndex, &item.Text, &item.Summary); err != nil {
			return nil, fmt.Errorf("sqlite scan aligned: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

func existingSegmentID(tx *sql.Tx, table, podcastID string, orderIndex int) string {
	var id string
	q := fmt.Sprintf("SELECT id FROM %s WHERE episode_id = ? AND order_index = ?", table)
	if err := tx.QueryRowContext(context.Background(), q, podcastID, orderIndex).Scan(&id); err == nil {
		return id
	}
	return ""
}

func existingMappingID(tx *sql.Tx, summaryID, transcriptID string) string {
	var id string
	const q = `
SELECT id
FROM transcript_summary_mapping
WHERE summary_segment_id = ? AND transcript_segment_id = ?;`
	if err := tx.QueryRowContext(context.Background(), q, summaryID, transcriptID).Scan(&id); err == nil {
		return id
	}
	return ""
}

var _ domain.ParagraphRepository = (*ParagraphSQLiteRepo)(nil)
