package podcastrepo

import (
	"context"
	"database/sql"
	"fmt"

	"podcast-summarizer/src/core/domain"
	"podcast-summarizer/src/internal/sqliteutil"
)

type PodcastSQLiteRepo struct {
	db *sql.DB
}

func NewPodcastSQLiteRepo(db *sql.DB) *PodcastSQLiteRepo {
	return &PodcastSQLiteRepo{db: db}
}

func (r *PodcastSQLiteRepo) UpsertSource(url, title, description string, durationSeconds *int, audioURL *string, transcriptURL *string) (string, error) {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("sqlite begin upsert source: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	id := sqliteutil.NewID()
	existingID, err := selectSourceID(tx, url)
	if err != nil {
		return "", fmt.Errorf("sqlite select existing source: %w", err)
	}
	if existingID != "" {
		id = existingID
	}

	hasTranscript := 0
	if transcriptURL != nil && *transcriptURL != "" {
		hasTranscript = 1
	}

	const episodeQ = `
INSERT INTO episode (id, podcast_url, title, description, duration_seconds, audio_url, transcript_url, has_transcript, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(podcast_url) DO UPDATE SET
	title = excluded.title,
	description = excluded.description,
	duration_seconds = excluded.duration_seconds,
	audio_url = excluded.audio_url,
	transcript_url = excluded.transcript_url,
	has_transcript = excluded.has_transcript,
	updated_at = CURRENT_TIMESTAMP;`
	if _, err := tx.ExecContext(context.Background(), episodeQ, id, url, title, description, durationSeconds, sqliteutil.NullableString(audioURL), sqliteutil.NullableString(transcriptURL), hasTranscript); err != nil {
		return "", fmt.Errorf("sqlite upsert episode: %w", err)
	}

	const sourceQ = `
INSERT INTO podcast_source (id, url, title, description, duration_seconds, audio_url, transcript_url, has_transcript, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(url) DO UPDATE SET
	title = excluded.title,
	description = excluded.description,
	duration_seconds = excluded.duration_seconds,
	audio_url = excluded.audio_url,
	transcript_url = excluded.transcript_url,
	has_transcript = excluded.has_transcript,
	updated_at = CURRENT_TIMESTAMP;`
	if _, err := tx.ExecContext(context.Background(), sourceQ, id, url, title, description, durationSeconds, sqliteutil.NullableString(audioURL), sqliteutil.NullableString(transcriptURL), hasTranscript); err != nil {
		return "", fmt.Errorf("sqlite upsert source: %w", err)
	}

	var sourceID string
	if err := tx.QueryRowContext(context.Background(), `SELECT id FROM episode WHERE podcast_url = ?`, url).Scan(&sourceID); err != nil {
		return "", fmt.Errorf("sqlite fetch source id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("sqlite commit source: %w", err)
	}
	return sourceID, nil
}

func (r *PodcastSQLiteRepo) MarkHasTranscript(id string) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("sqlite begin mark transcript: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.ExecContext(context.Background(), `UPDATE episode SET has_transcript = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(context.Background(), `UPDATE podcast_source SET has_transcript = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PodcastSQLiteRepo) GetSourceByURL(url string) (domain.PodcastSource, error) {
	const q = `
SELECT e.id, e.podcast_url, COALESCE(e.title, ''), e.has_transcript, e.created_at,
	(SELECT pj.id FROM processing_job pj WHERE pj.podcast_id = e.id ORDER BY pj.created_at DESC LIMIT 1) AS latest_job_id,
	(SELECT pj.status FROM processing_job pj WHERE pj.podcast_id = e.id ORDER BY pj.created_at DESC LIMIT 1) AS latest_status
FROM episode e
WHERE e.podcast_url = ?;`
	var item domain.PodcastSource
	var hasTranscript int
	var createdAt string
	var latestJobID sql.NullString
	var latestStatus sql.NullString
	if err := r.db.QueryRowContext(context.Background(), q, url).Scan(&item.ID, &item.URL, &item.Title, &hasTranscript, &createdAt, &latestJobID, &latestStatus); err != nil {
		return domain.PodcastSource{}, fmt.Errorf("sqlite get source by url: %w", err)
	}
	item.HasTranscript = hasTranscript == 1
	item.CreatedAt = sqliteutil.ParseTime(createdAt)
	if latestJobID.Valid {
		item.LatestJobID = &latestJobID.String
	}
	if latestStatus.Valid {
		item.LatestStatus = &latestStatus.String
	}
	return item, nil
}

func selectSourceID(tx *sql.Tx, url string) (string, error) {
	var id string
	err := tx.QueryRowContext(context.Background(), `SELECT id FROM episode WHERE podcast_url = ?`, url).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	err = tx.QueryRowContext(context.Background(), `SELECT id FROM podcast_source WHERE url = ?`, url).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err == sql.ErrNoRows {
		return "", nil
	}
	return "", err
}

func (r *PodcastSQLiteRepo) ListSources(limit int) ([]domain.PodcastSource, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
SELECT e.id, e.podcast_url, COALESCE(e.title, ''), e.has_transcript, e.created_at,
	(SELECT pj.id FROM processing_job pj WHERE pj.podcast_id = e.id ORDER BY pj.created_at DESC LIMIT 1) AS latest_job_id,
	(SELECT pj.status FROM processing_job pj WHERE pj.podcast_id = e.id ORDER BY pj.created_at DESC LIMIT 1) AS latest_status
FROM episode e
ORDER BY e.created_at DESC
LIMIT ?;`

	rows, err := r.db.QueryContext(context.Background(), q, limit)
	if err != nil {
		return nil, fmt.Errorf("sqlite list sources: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var items []domain.PodcastSource
	for rows.Next() {
		var item domain.PodcastSource
		var hasTranscript int
		var createdAt string
		var latestJobID sql.NullString
		var latestStatus sql.NullString
		if err := rows.Scan(&item.ID, &item.URL, &item.Title, &hasTranscript, &createdAt, &latestJobID, &latestStatus); err != nil {
			return nil, fmt.Errorf("sqlite scan source: %w", err)
		}
		item.HasTranscript = hasTranscript == 1
		item.CreatedAt = sqliteutil.ParseTime(createdAt)
		if latestJobID.Valid {
			item.LatestJobID = &latestJobID.String
		}
		if latestStatus.Valid {
			item.LatestStatus = &latestStatus.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

var _ domain.PodcastRepository = (*PodcastSQLiteRepo)(nil)
