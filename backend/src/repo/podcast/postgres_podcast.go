package podcastrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"podcast-summarizer/src/core/domain"
)

type PodcastPgRepo struct {
	pool *pgxpool.Pool
}

func NewPodcastPgRepo(pool *pgxpool.Pool) *PodcastPgRepo {
	return &PodcastPgRepo{pool: pool}
}

func (r *PodcastPgRepo) UpsertSource(url, title, description string, durationSeconds *int, audioURL *string, transcriptURL *string) (string, error) {
	const q = `
INSERT INTO podcast_source (url, title, description, duration_seconds, has_transcript)
VALUES ($1,$2,$3,$4, COALESCE($5,false))
ON CONFLICT (url) DO UPDATE SET
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  duration_seconds = EXCLUDED.duration_seconds,
  updated_at = now()
RETURNING id;
`
	ctx := context.Background()
	var id string
	hasTranscript := transcriptURL != nil
	// audioURL is unused for now; keep signature aligned with interface.
	if err := r.pool.QueryRow(ctx, q, url, title, description, durationSeconds, hasTranscript).Scan(&id); err != nil {
		return "", fmt.Errorf("upsert source: %w", err)
	}
	return id, nil
}

func (r *PodcastPgRepo) MarkHasTranscript(id string) error {
	const q = `UPDATE podcast_source SET has_transcript=true, updated_at=now() WHERE id=$1`
	_, err := r.pool.Exec(context.Background(), q, id)
	return err
}

func (r *PodcastPgRepo) ListSources(limit int) ([]domain.PodcastSource, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
SELECT ps.id, ps.url, ps.title, ps.has_transcript, ps.created_at,
       pj.id as job_id, pj.status as job_status
FROM podcast_source ps
LEFT JOIN LATERAL (
    SELECT id, status
    FROM processing_job
    WHERE podcast_id = ps.id
    ORDER BY created_at DESC
    LIMIT 1
) pj ON true
ORDER BY ps.created_at DESC
LIMIT $1;
`
	rows, err := r.pool.Query(context.Background(), q, limit)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()

	var items []domain.PodcastSource
	for rows.Next() {
		var ps domain.PodcastSource
		if err := rows.Scan(&ps.ID, &ps.URL, &ps.Title, &ps.HasTranscript, &ps.CreatedAt, &ps.LatestJobID, &ps.LatestStatus); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		items = append(items, ps)
	}
	return items, nil
}
