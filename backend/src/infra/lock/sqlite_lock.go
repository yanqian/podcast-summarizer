package lock

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"podcast-summarizer/src/core/app/jobs"
	"podcast-summarizer/src/internal/sqliteutil"
)

type SQLiteLock struct {
	db  *sql.DB
	ttl time.Duration
}

func NewSQLiteLock(db *sql.DB, ttl time.Duration) *SQLiteLock {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &SQLiteLock{db: db, ttl: ttl}
}

func (l *SQLiteLock) Acquire(ctx context.Context, key string) (bool, error) {
	now := time.Now().UTC()
	if _, err := l.db.ExecContext(ctx, `DELETE FROM job_lock WHERE expires_at <= ?`, sqliteutil.FormatTime(now)); err != nil {
		return false, fmt.Errorf("sqlite prune locks: %w", err)
	}
	_, err := l.db.ExecContext(ctx, `INSERT INTO job_lock (key, expires_at) VALUES (?, ?)`, key, sqliteutil.FormatTime(now.Add(l.ttl)))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "constraint") {
			return false, nil
		}
		return false, fmt.Errorf("sqlite acquire lock: %w", err)
	}
	return true, nil
}

func (l *SQLiteLock) Release(ctx context.Context, key string) error {
	if _, err := l.db.ExecContext(ctx, `DELETE FROM job_lock WHERE key = ?`, key); err != nil {
		return fmt.Errorf("sqlite release lock: %w", err)
	}
	return nil
}

var _ jobs.Locker = (*SQLiteLock)(nil)
