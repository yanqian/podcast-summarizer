package contract

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"podcast-summarizer/src/core/domain"
	podcastrepo "podcast-summarizer/src/repo/podcast"
	processingrepo "podcast-summarizer/src/repo/processing"
)

func TestExportPayload(t *testing.T) {
	router, db := newTestRouterWithDB(t)
	podcastID := seedExportEpisode(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/podcasts/"+podcastID+"/export", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(rr.Body.Bytes()) == 0 {
		t.Fatalf("expected export payload")
	}
	var resp struct {
		TranscriptText string `json:"transcriptText"`
		SummaryText    string `json:"summaryText"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if !strings.Contains(resp.TranscriptText, "Export transcript.") || !strings.Contains(resp.SummaryText, "Export summary.") {
		t.Fatalf("expected segment export text, got %+v", resp)
	}
}

func TestExportNotFoundWhenPodcastHasNoTranscriptData(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/podcasts/missing/export", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func seedExportEpisode(t *testing.T, db *sql.DB) string {
	t.Helper()

	podcasts := podcastrepo.NewPodcastSQLiteRepo(db)
	processing := processingrepo.NewProcessingSQLiteRepo(db)

	audioURL := "http://example.com/export.mp3"
	podcastID, err := podcasts.UpsertSource("http://example.com/export", "Export episode", "Tester", nil, &audioURL, nil)
	if err != nil {
		t.Fatalf("seed export podcast: %v", err)
	}
	transcripts, err := processing.SaveTranscriptSegments(podcastID, []domain.TranscriptSegment{{OrderIndex: 1, Text: "Export transcript."}})
	if err != nil {
		t.Fatalf("seed export transcript segments: %v", err)
	}
	summaries, err := processing.SaveSummarySegments(podcastID, []domain.SummarySegment{{OrderIndex: 1, Text: "Export summary."}})
	if err != nil {
		t.Fatalf("seed export summary segments: %v", err)
	}
	if err := processing.SaveTranscriptSummaryMappings(podcastID, []domain.TranscriptSummaryMapping{{
		SummarySegmentID:    summaries[0].ID,
		TranscriptSegmentID: transcripts[0].ID,
		SourceOrder:         1,
	}}); err != nil {
		t.Fatalf("seed export summary mappings: %v", err)
	}
	return podcastID
}
