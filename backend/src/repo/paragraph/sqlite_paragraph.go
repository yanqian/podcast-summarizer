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
	for _, p := range paragraphs {
		if _, err := tx.ExecContext(context.Background(), q, sqliteutil.NewID(), podcastID, p.OrderIndex, p.Text); err != nil {
			return fmt.Errorf("sqlite save transcript: %w", err)
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
	for _, summary := range summaries {
		if _, err := tx.ExecContext(context.Background(), q, sqliteutil.NewID(), summary.Text, podcastID, summary.OrderIndex); err != nil {
			return fmt.Errorf("sqlite save summaries: %w", err)
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
	defer rows.Close()

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

var _ domain.ParagraphRepository = (*ParagraphSQLiteRepo)(nil)
