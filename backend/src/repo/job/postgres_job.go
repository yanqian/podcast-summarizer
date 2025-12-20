package jobrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"podcast-summarizer/src/core/domain"
)

type JobPgRepo struct {
	pool *pgxpool.Pool
}

func NewJobPgRepo(pool *pgxpool.Pool) *JobPgRepo {
	return &JobPgRepo{pool: pool}
}

func (r *JobPgRepo) Create(job domain.ProcessingJob) (string, error) {
	const q = `
INSERT INTO processing_job (podcast_id, type, status, started_at)
VALUES ($1,$2,$3,$4) RETURNING id`
	var id string
	if err := r.pool.QueryRow(context.Background(), q, job.PodcastID, job.Type, job.Status, time.Now()).Scan(&id); err != nil {
		return "", fmt.Errorf("create job: %w", err)
	}
	return id, nil
}

func (r *JobPgRepo) UpdateStatus(id, status string, durationMs *int, errorMessage *string) error {
	const q = `
UPDATE processing_job
SET status=$2, completed_at=CASE WHEN $2 IN ('succeeded','failed') THEN now() ELSE completed_at END,
    duration_ms=COALESCE($3,duration_ms),
    error_message=$4
WHERE id=$1`
	_, err := r.pool.Exec(context.Background(), q, id, status, durationMs, errorMessage)
	return err
}

func (r *JobPgRepo) Get(id string) (domain.ProcessingJob, error) {
	const q = `
SELECT id, podcast_id, type, status, duration_ms, error_message
FROM processing_job WHERE id=$1`
	var job domain.ProcessingJob
	err := r.pool.QueryRow(context.Background(), q, id).Scan(
		&job.ID, &job.PodcastID, &job.Type, &job.Status, &job.DurationMs, &job.Error)
	if err != nil {
		return domain.ProcessingJob{}, fmt.Errorf("get job: %w", err)
	}
	return job, nil
}
