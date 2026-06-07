package contract

import (
	"context"
	"net/http"
	"testing"

	"podcast-summarizer/src/api"
	dbinfra "podcast-summarizer/src/infra/db"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	db, err := dbinfra.NewSQLite(context.Background(), t.TempDir()+"/podcast.db")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return api.NewRouter(api.Dependencies{SQLite: db})
}
