package paragraphrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"podcast-summarizer/src/core/domain"
)

type ParagraphPgRepo struct {
	pool *pgxpool.Pool
}

func NewParagraphPgRepo(pool *pgxpool.Pool) *ParagraphPgRepo {
	return &ParagraphPgRepo{pool: pool}
}

func (r *ParagraphPgRepo) SaveTranscript(podcastID string, paragraphs []domain.Paragraph) error {
	const q = `
INSERT INTO transcript_paragraph (podcast_id, order_index, text, source)
VALUES ($1,$2,$3,'generated')
ON CONFLICT (podcast_id, order_index) DO UPDATE SET text=EXCLUDED.text`
	batch := &pgx.Batch{}
	for _, p := range paragraphs {
		batch.Queue(q, podcastID, p.OrderIndex, p.Text)
	}
	br := r.pool.SendBatch(context.Background(), batch)
	defer br.Close()
	for i := 0; i < len(paragraphs); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("save transcript: %w", err)
		}
	}
	return nil
}

func (r *ParagraphPgRepo) SaveSummaries(podcastID string, summaries []domain.Summary) error {
	const q = `
INSERT INTO summary_paragraph (transcript_paragraph_id, summary_text)
SELECT tp.id, $3
FROM transcript_paragraph tp
WHERE tp.podcast_id = $1 AND tp.order_index = $2
ON CONFLICT (transcript_paragraph_id) DO UPDATE SET summary_text=EXCLUDED.summary_text`
	batch := &pgx.Batch{}
	for _, s := range summaries {
		batch.Queue(q, podcastID, s.OrderIndex, s.Text)
	}
	br := r.pool.SendBatch(context.Background(), batch)
	defer br.Close()
	for i := 0; i < len(summaries); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("save summaries: %w", err)
		}
	}
	return nil
}

func (r *ParagraphPgRepo) GetAligned(podcastID string) ([]domain.ParagraphWithSummary, error) {
	const q = `
SELECT tp.order_index, tp.text, COALESCE(sp.summary_text,'')
FROM transcript_paragraph tp
LEFT JOIN summary_paragraph sp ON sp.transcript_paragraph_id = tp.id
WHERE tp.podcast_id=$1
ORDER BY tp.order_index`
	rows, err := r.pool.Query(context.Background(), q, podcastID)
	if err != nil {
		return nil, fmt.Errorf("get aligned: %w", err)
	}
	defer rows.Close()
	var result []domain.ParagraphWithSummary
	for rows.Next() {
		var p domain.ParagraphWithSummary
		if err := rows.Scan(&p.OrderIndex, &p.Text, &p.Summary); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}
