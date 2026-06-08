package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIngestWithExistingTranscript(t *testing.T) {
	router := newTestRouter(t)

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

func TestIngestDuplicateURLReturnsExistingEpisodeWithoutNewJob(t *testing.T) {
	router, db := newTestRouterWithDB(t)

	first := postIngest(t, router, `{"url":"http://example.com/duplicate-podcast"}`)
	if first.code != http.StatusAccepted {
		t.Fatalf("expected first submit 202, got %d body=%s", first.code, first.body)
	}
	firstPodcastID, ok := first.response["podcastId"].(string)
	if !ok || firstPodcastID == "" {
		t.Fatalf("expected first response podcastId, got %+v", first.response)
	}
	firstJobID, ok := first.response["jobId"].(string)
	if !ok || firstJobID == "" {
		t.Fatalf("expected first response jobId, got %+v", first.response)
	}

	second := postIngest(t, router, `{"url":"http://example.com/duplicate-podcast"}`)
	if second.code != http.StatusOK {
		t.Fatalf("expected duplicate submit 200, got %d body=%s", second.code, second.body)
	}
	if second.response["status"] != "existing" || second.response["existing"] != true {
		t.Fatalf("expected duplicate existing response, got %+v", second.response)
	}
	if second.response["podcastId"] != firstPodcastID {
		t.Fatalf("duplicate returned different podcastId: first=%s second=%v", firstPodcastID, second.response["podcastId"])
	}
	if second.response["jobId"] != firstJobID {
		t.Fatalf("duplicate returned different jobId: first=%s second=%v", firstJobID, second.response["jobId"])
	}

	var jobCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM processing_job WHERE podcast_id = ?`, firstPodcastID).Scan(&jobCount); err != nil {
		t.Fatalf("count processing jobs: %v", err)
	}
	if jobCount != 1 {
		t.Fatalf("expected duplicate submit to leave one processing job, got %d", jobCount)
	}
}

type ingestTestResponse struct {
	code     int
	body     string
	response map[string]any
}

func postIngest(t *testing.T, router http.Handler, body string) ingestTestResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/podcasts/ingest", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v body=%s", err, rr.Body.String())
	}
	return ingestTestResponse{code: rr.Code, body: rr.Body.String(), response: resp}
}
