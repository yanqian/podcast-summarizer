package contract

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"podcast-summarizer/src/core/domain"
	jobrepo "podcast-summarizer/src/repo/job"
	podcastrepo "podcast-summarizer/src/repo/podcast"
	processingrepo "podcast-summarizer/src/repo/processing"
)

func TestEpisodeDetailReturnsTranscriptSummaryMapping(t *testing.T) {
	router, db := newTestRouterWithDB(t)
	podcastID := seedCompletedEpisode(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/podcasts/"+podcastID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		PodcastID          string `json:"podcastId"`
		Status             string `json:"status"`
		TranscriptSegments []struct {
			ID         string `json:"id"`
			OrderIndex int    `json:"orderIndex"`
			Text       string `json:"text"`
		} `json:"transcriptSegments"`
		SummarySegments []struct {
			ID                         string   `json:"id"`
			Text                       string   `json:"text"`
			SourceTranscriptSegmentIDs []string `json:"sourceTranscriptSegmentIds"`
		} `json:"summarySegments"`
		Paragraphs []struct {
			ParagraphID string `json:"paragraphId"`
			Summary     string `json:"summary"`
		} `json:"paragraphs"`
		LatestJob struct {
			Status string `json:"status"`
		} `json:"latestJob"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.PodcastID != podcastID || resp.Status != "succeeded" || resp.LatestJob.Status != "succeeded" {
		t.Fatalf("unexpected episode status response: %+v", resp)
	}
	if len(resp.TranscriptSegments) != 2 {
		t.Fatalf("expected 2 transcript segments, got %+v", resp.TranscriptSegments)
	}
	if len(resp.SummarySegments) != 1 {
		t.Fatalf("expected 1 summary segment, got %+v", resp.SummarySegments)
	}
	if got := resp.SummarySegments[0].SourceTranscriptSegmentIDs; len(got) != 2 || got[0] != resp.TranscriptSegments[0].ID || got[1] != resp.TranscriptSegments[1].ID {
		t.Fatalf("summary source ids were not mapped to transcript segments: summary=%+v transcripts=%+v", resp.SummarySegments[0], resp.TranscriptSegments)
	}
	if len(resp.Paragraphs) != 2 || resp.Paragraphs[0].Summary != "Grouped summary" || resp.Paragraphs[1].Summary != "Grouped summary" {
		t.Fatalf("expected compatibility paragraphs with mapped summary text, got %+v", resp.Paragraphs)
	}
}

func TestEpisodeViewCompatibilityRouteReturnsDetail(t *testing.T) {
	router, db := newTestRouterWithDB(t)
	podcastID := seedCompletedEpisode(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/podcasts/"+podcastID+"/view", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		PodcastID  string `json:"podcastId"`
		Paragraphs []struct {
			Text string `json:"text"`
		} `json:"paragraphs"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.PodcastID != podcastID || len(resp.Paragraphs) != 2 {
		t.Fatalf("unexpected view response: %+v", resp)
	}
}

func TestEpisodeDetailNotFound(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/podcasts/missing-episode", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestEpisodeStatusIncludesFailedProcessingError(t *testing.T) {
	router, db := newTestRouterWithDB(t)
	podcasts := podcastrepo.NewPodcastSQLiteRepo(db)
	jobs := jobrepo.NewJobSQLiteRepo(db)

	audioURL := "http://example.com/failure.mp3"
	podcastID, err := podcasts.UpsertSource("http://example.com/failure", "Failed episode", "Tester", nil, &audioURL, nil)
	if err != nil {
		t.Fatalf("seed podcast: %v", err)
	}
	jobID, err := jobs.Create(domain.ProcessingJob{PodcastID: podcastID, Type: "ingest", Status: "queued"})
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}
	duration := 12
	errMessage := "transcript pipeline failed: chunk audio: ffmpeg failed"
	if err := jobs.UpdateStatus(jobID, "failed", &duration, &errMessage); err != nil {
		t.Fatalf("mark job failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/podcasts/"+podcastID+"/status", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		PodcastID string `json:"podcastId"`
		Status    string `json:"status"`
		LatestJob struct {
			JobID        string `json:"jobId"`
			Status       string `json:"status"`
			ErrorMessage string `json:"errorMessage"`
		} `json:"latestJob"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.PodcastID != podcastID || resp.Status != "failed" || resp.LatestJob.JobID != jobID || resp.LatestJob.ErrorMessage != errMessage {
		t.Fatalf("unexpected failed status response: %+v", resp)
	}
}

func seedCompletedEpisode(t *testing.T, db *sql.DB) string {
	t.Helper()

	podcasts := podcastrepo.NewPodcastSQLiteRepo(db)
	processing := processingrepo.NewProcessingSQLiteRepo(db)
	jobs := jobrepo.NewJobSQLiteRepo(db)

	audioURL := "http://example.com/completed.mp3"
	podcastID, err := podcasts.UpsertSource("http://example.com/completed", "Completed episode", "Tester", nil, &audioURL, nil)
	if err != nil {
		t.Fatalf("seed podcast: %v", err)
	}
	if err := podcasts.MarkHasTranscript(podcastID); err != nil {
		t.Fatalf("mark transcript: %v", err)
	}
	jobID, err := jobs.Create(domain.ProcessingJob{PodcastID: podcastID, Type: "ingest", Status: "queued"})
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}
	duration := 42
	if err := jobs.UpdateStatus(jobID, "succeeded", &duration, nil); err != nil {
		t.Fatalf("mark job succeeded: %v", err)
	}
	transcripts, err := processing.SaveTranscriptSegments(podcastID, []domain.TranscriptSegment{
		{OrderIndex: 1, Text: "First transcript segment."},
		{OrderIndex: 2, Text: "Second transcript segment."},
	})
	if err != nil {
		t.Fatalf("seed transcript segments: %v", err)
	}
	summaries, err := processing.SaveSummarySegments(podcastID, []domain.SummarySegment{
		{OrderIndex: 1, Text: "Grouped summary"},
	})
	if err != nil {
		t.Fatalf("seed summary segments: %v", err)
	}
	if err := processing.SaveTranscriptSummaryMappings(podcastID, []domain.TranscriptSummaryMapping{
		{SummarySegmentID: summaries[0].ID, TranscriptSegmentID: transcripts[0].ID, SourceOrder: 1},
		{SummarySegmentID: summaries[0].ID, TranscriptSegmentID: transcripts[1].ID, SourceOrder: 2},
	}); err != nil {
		t.Fatalf("seed summary mappings: %v", err)
	}
	return podcastID
}
