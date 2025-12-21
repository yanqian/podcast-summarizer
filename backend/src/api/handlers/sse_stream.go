package handlers

import (
	"log"
	"net/http"
	"strings"

	"podcast-summarizer/src/core/app/jobs"
)

// StreamHandler streams transcript chunk events for a job via SSE.
func StreamHandler(mgr *jobs.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) == 0 {
			http.Error(w, "missing job id", http.StatusBadRequest)
			return
		}
		jobID := parts[len(parts)-1]
		ch := mgr.Subscribe(jobID)
		if ch == nil {
			log.Printf("sse subscribe failed: job %s channel nil", jobID)
			http.Error(w, "stream unavailable", http.StatusServiceUnavailable)
			return
		}
		for msg := range ch {
			// SSE requires data prefix and double newline terminator.
			if _, err := w.Write([]byte("data: ")); err != nil {
				log.Printf("sse write prefix failed job=%s err=%v", jobID, err)
				return
			}
			if _, err := w.Write(msg); err != nil {
				log.Printf("sse write body failed job=%s err=%v", jobID, err)
				return
			}
			if _, err := w.Write([]byte("\n\n")); err != nil {
				log.Printf("sse write suffix failed job=%s err=%v", jobID, err)
				return
			}
			flusher.Flush()
		}
	}
}
