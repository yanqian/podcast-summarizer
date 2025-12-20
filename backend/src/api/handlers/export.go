package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"podcast-summarizer/src/core/app"
	"podcast-summarizer/src/core/domain"
)

type exportResponse struct {
	PodcastID      string `json:"podcastId"`
	TranscriptText string `json:"transcriptText"`
	SummaryText    string `json:"summaryText"`
}

// ExportHandler returns transcript and summary text bundles.
func ExportHandler(repo domain.ParagraphRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		podcastID := parts[2]
		aligned, err := repo.GetAligned(podcastID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		transcriptText, summaryText := app.FormatTextBundle(aligned)
		resp := exportResponse{
			PodcastID:      podcastID,
			TranscriptText: transcriptText,
			SummaryText:    summaryText,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
