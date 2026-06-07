package jobrepo

import (
	"context"
	"database/sql"
	"fmt"

	"podcast-summarizer/src/core/app/jobs"
	"podcast-summarizer/src/core/domain"
	"podcast-summarizer/src/internal/sqliteutil"
)

type JobSQLiteRepo struct {
	db *sql.DB
}

func NewJobSQLiteRepo(db *sql.DB) *JobSQLiteRepo {
	return &JobSQLiteRepo{db: db}
}

func (r *JobSQLiteRepo) Create(job domain.ProcessingJob) (string, error) {
	const q = `
INSERT INTO processing_job (id, podcast_id, type, status, started_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP);`
	id := sqliteutil.NewID()
	if _, err := r.db.ExecContext(context.Background(), q, id, job.PodcastID, job.Type, job.Status); err != nil {
		return "", fmt.Errorf("sqlite create job: %w", err)
	}
	return id, nil
}

func (r *JobSQLiteRepo) UpdateStatus(id, status string, durationMs *int, errorMessage *string) error {
	const q = `
UPDATE processing_job
SET status = ?,
	completed_at = CASE WHEN ? IN ('succeeded', 'failed') THEN CURRENT_TIMESTAMP ELSE completed_at END,
	duration_ms = COALESCE(?, duration_ms),
	error_message = ?
WHERE id = ?;`
	_, err := r.db.ExecContext(context.Background(), q, status, status, durationMs, sqliteutil.NullableString(errorMessage), id)
	return err
}

func (r *JobSQLiteRepo) Get(id string) (domain.ProcessingJob, error) {
	const q = `
SELECT id, podcast_id, type, status, duration_ms, error_message
FROM processing_job
WHERE id = ?;`
	var job domain.ProcessingJob
	var duration sql.NullInt64
	var errMsg sql.NullString
	if err := r.db.QueryRowContext(context.Background(), q, id).Scan(&job.ID, &job.PodcastID, &job.Type, &job.Status, &duration, &errMsg); err != nil {
		return domain.ProcessingJob{}, fmt.Errorf("sqlite get job: %w", err)
	}
	if duration.Valid {
		v := int(duration.Int64)
		job.DurationMs = &v
	}
	if errMsg.Valid {
		job.Error = &errMsg.String
	}
	return job, nil
}

var _ jobs.JobRepository = (*JobSQLiteRepo)(nil)
