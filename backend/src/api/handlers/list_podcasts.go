package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"podcast-summarizer/src/core/domain"
)

type listResponse struct {
	Items []domain.PodcastSource `json:"items"`
}

// ListPodcastsHandler returns recent podcast sources with latest job status.
func ListPodcastsHandler(repo domain.PodcastRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		items, err := repo.ListSources(50)
		if err != nil {
			log.Printf("list podcasts error: %v", err)
			http.Error(w, "unable to list podcasts", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listResponse{Items: items})
	}
}
