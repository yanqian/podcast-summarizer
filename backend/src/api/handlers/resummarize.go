package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

type ResummarizeProcessingRepository interface {
	ListTranscriptSegments(episodeID string) ([]domain.TranscriptSegment, error)
	SaveSummarySegments(episodeID string, segments []domain.SummarySegment) ([]domain.SummarySegment, error)
	SaveTranscriptSummaryMappings(episodeID string, mappings []domain.TranscriptSummaryMapping) error
}

// ResummarizeHandler re-runs the summarizer on stored transcript segments.
func ResummarizeHandler(repo ResummarizeProcessingRepository, summarySvc *jobs.ClientSummaryService) http.HandlerFunc {
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

		transcripts, err := repo.ListTranscriptSegments(podcastID)
		if err != nil || len(transcripts) == 0 {
			http.Error(w, "transcript not found", http.StatusNotFound)
			return
		}

		result, err := summarySvc.SummarizeSegments(r.Context(), transcripts)
		if err != nil {
			log.Printf("resummarize: summarize error podcast=%s: %v", podcastID, err)
			http.Error(w, "summarize failed", http.StatusInternalServerError)
			return
		}

		count, err := saveSummaryResult(r.Context(), repo, podcastID, result)
		if err != nil {
			log.Printf("resummarize: save summary mappings error podcast=%s: %v", podcastID, err)
			http.Error(w, "persist failed", http.StatusInternalServerError)
			return
		}

		resp := resummarizeResponse{
			PodcastID: podcastID,
			Status:    "ok",
			Count:     count,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func saveSummaryResult(ctx context.Context, repo ResummarizeProcessingRepository, podcastID string, result domain.SummaryResult) (int, error) {
	_ = ctx
	provider := nonEmptyStringPtr(result.Provider)
	model := nonEmptyStringPtr(result.Model)
	summaries := make([]domain.SummarySegment, 0, len(result.Segments))
	sourceIDsByOrder := make([][]string, 0, len(result.Segments))
	for idx, segment := range result.Segments {
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			return 0, fmt.Errorf("summary segment %d is empty", idx+1)
		}
		sourceIDs := compactStrings(segment.SourceTranscriptSegmentIDs)
		if len(sourceIDs) == 0 {
			return 0, fmt.Errorf("summary segment %d has no source transcript segments", idx+1)
		}
		orderIndex := segment.OrderIndex
		if orderIndex <= 0 {
			orderIndex = idx + 1
		}
		summaries = append(summaries, domain.SummarySegment{
			OrderIndex: orderIndex,
			Text:       text,
			Provider:   provider,
			Model:      model,
		})
		sourceIDsByOrder = append(sourceIDsByOrder, sourceIDs)
	}

	savedSummaries, err := repo.SaveSummarySegments(podcastID, summaries)
	if err != nil {
		return 0, err
	}
	if len(savedSummaries) != len(sourceIDsByOrder) {
		return 0, fmt.Errorf("saved %d summary segments for %d mapping groups", len(savedSummaries), len(sourceIDsByOrder))
	}
	mappings := make([]domain.TranscriptSummaryMapping, 0)
	for idx, summary := range savedSummaries {
		for sourceOrder, transcriptID := range sourceIDsByOrder[idx] {
			mappings = append(mappings, domain.TranscriptSummaryMapping{
				SummarySegmentID:    summary.ID,
				TranscriptSegmentID: transcriptID,
				SourceOrder:         sourceOrder + 1,
			})
		}
	}
	return len(savedSummaries), repo.SaveTranscriptSummaryMappings(podcastID, mappings)
}

func nonEmptyStringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
