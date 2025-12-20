package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"podcast-summarizer/src/core/domain"
)

type paragraph struct {
	ParagraphID string   `json:"paragraphId"`
	OrderIndex  int      `json:"orderIndex"`
	Text        string   `json:"text"`
	Timestamp   *float64 `json:"timestampSeconds,omitempty"`
	Summary     string   `json:"summary"`
}

type transcriptResponse struct {
	PodcastID  string      `json:"podcastId"`
	Title      string      `json:"title,omitempty"`
	Paragraphs []paragraph `json:"paragraphs"`
}

// ViewHandler returns aligned transcript/summaries.
func ViewHandler(repo domain.ParagraphRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			http.NotFound(w, r)
			return
		}
		podcastID := pathParts[2]
		aligned, err := repo.GetAligned(podcastID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		resp := transcriptResponse{
			PodcastID: podcastID,
		}
		for _, p := range aligned {
			resp.Paragraphs = append(resp.Paragraphs, paragraph{
				ParagraphID: "",
				OrderIndex:  p.OrderIndex,
				Text:        p.Text,
				Summary:     p.Summary,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
