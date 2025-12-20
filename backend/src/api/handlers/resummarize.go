package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"podcast-summarizer/src/core/app/jobs"
	"podcast-summarizer/src/core/domain"
)

type resummarizeResponse struct {
	PodcastID string `json:"podcastId"`
	Status    string `json:"status"`
	Count     int    `json:"summaries"`
}

// ResummarizeHandler re-runs the summarizer on stored transcript paragraphs.
func ResummarizeHandler(repo domain.ParagraphRepository, summarySvc *jobs.ClientSummaryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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
		if err != nil || len(aligned) == 0 {
			http.Error(w, "transcript not found", http.StatusNotFound)
			return
		}

		var paragraphs []domain.Paragraph
		for _, p := range aligned {
			paragraphs = append(paragraphs, domain.Paragraph{OrderIndex: p.OrderIndex, Text: p.Text})
		}

		summaries, err := summarySvc.Summarize(paragraphs)
		if err != nil {
			log.Printf("resummarize: summarize error podcast=%s: %v", podcastID, err)
			http.Error(w, "summarize failed", http.StatusInternalServerError)
			return
		}

		if err := repo.SaveSummaries(podcastID, summaries); err != nil {
			log.Printf("resummarize: save summaries error podcast=%s: %v", podcastID, err)
			http.Error(w, "persist failed", http.StatusInternalServerError)
			return
		}

		resp := resummarizeResponse{
			PodcastID: podcastID,
			Status:    "ok",
			Count:     len(summaries),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
