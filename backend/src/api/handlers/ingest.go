package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"podcast-summarizer/src/core/app"
	"podcast-summarizer/src/core/app/jobs"
)

type ingestRequest struct {
	URL string `json:"url"`
}

type jobAccepted struct {
	JobID        string  `json:"jobId,omitempty"`
	PodcastID    string  `json:"podcastId"`
	Status       string  `json:"status"`
	Existing     bool    `json:"existing"`
	LatestStatus *string `json:"latestStatus,omitempty"`
}

// IngestHandler accepts a podcast URL, runs lookup, and enqueues processing.
func IngestHandler(svc *app.IngestService, mgr *jobs.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // cap request size
		var req ingestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
			log.Printf("ingest decode error: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if !isValidURL(req.URL) {
			log.Printf("ingest invalid url: %q", req.URL)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		result, err := svc.Ingest(req.URL)
		if err != nil {
			log.Printf("ingest service error url=%q: %v", req.URL, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if result.Existing {
			log.Printf("ingest duplicate url=%q podcast=%s status=%v", req.URL, result.PodcastID, result.Source.LatestStatus)
			resp := jobAccepted{
				PodcastID:    result.PodcastID,
				Status:       "existing",
				Existing:     true,
				LatestStatus: result.Source.LatestStatus,
			}
			if result.Source.LatestJobID != nil {
				resp.JobID = *result.Source.LatestJobID
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		audioURL := ""
		var transcriptURL *string
		if result.Metadata != nil {
			audioURL = result.Metadata.AudioURL
			transcriptURL = result.Metadata.TranscriptURL
		}
		jobID, err := mgr.StartJob(r.Context(), result.PodcastID, audioURL, transcriptURL)
		if err != nil {
			log.Printf("ingest enqueue error podcast=%s url=%s: %v", result.PodcastID, req.URL, err)
			http.Error(w, "unable to enqueue job", http.StatusInternalServerError)
			return
		}
		log.Printf("[job:%s] enqueued podcast=%s url=%s audio=%s", jobID, result.PodcastID, req.URL, audioURL)

		resp := jobAccepted{
			JobID:     jobID,
			PodcastID: result.PodcastID,
			Status:    "queued",
			Existing:  false,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func isValidURL(raw string) bool {
	if len(raw) == 0 || len(raw) > 2048 {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	return true
}
