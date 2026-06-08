package jobrepo

import (
	"context"
	"database/sql"
	"errors"
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
SELECT id, podcast_id, type, status, started_at, completed_at, duration_ms, error_message
FROM processing_job
WHERE id = ?;`
	return r.scanJob(q, id)
}

func (r *JobSQLiteRepo) GetLatestByPodcastID(podcastID string) (domain.ProcessingJob, error) {
	const q = `
SELECT id, podcast_id, type, status, started_at, completed_at, duration_ms, error_message
FROM processing_job
WHERE podcast_id = ?
ORDER BY created_at DESC
LIMIT 1;`
	return r.scanJob(q, podcastID)
}

func (r *JobSQLiteRepo) scanJob(query string, arg string) (domain.ProcessingJob, error) {
	var job domain.ProcessingJob
	var startedAt string
	var completedAt sql.NullString
	var duration sql.NullInt64
	var errMsg sql.NullString
	if err := r.db.QueryRowContext(context.Background(), query, arg).Scan(&job.ID, &job.PodcastID, &job.Type, &job.Status, &startedAt, &completedAt, &duration, &errMsg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProcessingJob{}, sql.ErrNoRows
		}
		return domain.ProcessingJob{}, fmt.Errorf("sqlite get job: %w", err)
	}
	job.StartedAt = sqliteutil.ParseTime(startedAt)
	if completedAt.Valid {
		v := sqliteutil.ParseTime(completedAt.String)
		job.CompletedAt = &v
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
