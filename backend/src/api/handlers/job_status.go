package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"podcast-summarizer/src/core/app/jobs"
	"podcast-summarizer/src/internal/sqliteutil"
)

type jobStatusResponse struct {
	JobID        string  `json:"jobId"`
	PodcastID    string  `json:"podcastId"`
	Type         string  `json:"type"`
	Status       string  `json:"status"`
	StartedAt    string  `json:"startedAt,omitempty"`
	CompletedAt  string  `json:"completedAt,omitempty"`
	DurationMs   *int    `json:"durationMs,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// JobStatusHandler returns stub status for a job.
func JobStatusHandler(repo jobs.JobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		jobID := parts[len(parts)-1]
		job, err := repo.Get(jobID)
		if err != nil {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		resp := jobStatusResponse{
			JobID:        job.ID,
			PodcastID:    job.PodcastID,
			Type:         job.Type,
			Status:       job.Status,
			DurationMs:   job.DurationMs,
			ErrorMessage: job.Error,
		}
		if !job.StartedAt.IsZero() {
			resp.StartedAt = sqliteutil.FormatTime(job.StartedAt)
		}
		if job.CompletedAt != nil && !job.CompletedAt.IsZero() {
			resp.CompletedAt = sqliteutil.FormatTime(*job.CompletedAt)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
