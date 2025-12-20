package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"podcast-summarizer/src/api"
)

func TestIngestWithExistingTranscript(t *testing.T) {
	router := api.NewRouter(api.Dependencies{})

	body := []byte(`{"url":"http://example.com/podcast"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/podcasts/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp["jobId"] == nil || resp["podcastId"] == nil {
		t.Fatalf("expected jobId and podcastId in response")
	}
}
