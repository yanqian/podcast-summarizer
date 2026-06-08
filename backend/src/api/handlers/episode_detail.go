package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"podcast-summarizer/src/core/domain"
	"podcast-summarizer/src/internal/sqliteutil"
)

type EpisodeRepository interface {
	GetEpisode(id string) (domain.Episode, error)
}

type LatestJobRepository interface {
	GetLatestByPodcastID(podcastID string) (domain.ProcessingJob, error)
}

type ProcessingReadRepository interface {
	ListTranscriptSegments(episodeID string) ([]domain.TranscriptSegment, error)
	ListSummarySegments(episodeID string) ([]domain.SummarySegment, error)
	GetSummarySourceMappings(episodeID string) ([]domain.SummarySourceMapping, error)
}

type episodeJobResponse struct {
	JobID        string  `json:"jobId"`
	PodcastID    string  `json:"podcastId"`
	Type         string  `json:"type"`
	Status       string  `json:"status"`
	StartedAt    string  `json:"startedAt,omitempty"`
	CompletedAt  string  `json:"completedAt,omitempty"`
	DurationMs   *int    `json:"durationMs,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

type transcriptSegmentResponse struct {
	ID           string   `json:"id"`
	OrderIndex   int      `json:"orderIndex"`
	Text         string   `json:"text"`
	AudioChunkID *string  `json:"audioChunkId,omitempty"`
	StartSeconds *float64 `json:"startSeconds,omitempty"`
	EndSeconds   *float64 `json:"endSeconds,omitempty"`
	Provider     *string  `json:"provider,omitempty"`
	Model        *string  `json:"model,omitempty"`
	CreatedAt    string   `json:"createdAt,omitempty"`
}

type summarySegmentResponse struct {
	ID                         string   `json:"id"`
	OrderIndex                 int      `json:"orderIndex"`
	Text                       string   `json:"text"`
	Provider                   *string  `json:"provider,omitempty"`
	Model                      *string  `json:"model,omitempty"`
	CreatedAt                  string   `json:"createdAt,omitempty"`
	SourceTranscriptSegmentIDs []string `json:"sourceTranscriptSegmentIds"`
}

type episodeDetailResponse struct {
	PodcastID          string                      `json:"podcastId"`
	ID                 string                      `json:"id"`
	URL                string                      `json:"url"`
	Title              string                      `json:"title"`
	Description        string                      `json:"description,omitempty"`
	DurationSeconds    *int                        `json:"durationSeconds,omitempty"`
	AudioURL           *string                     `json:"audioUrl,omitempty"`
	TranscriptURL      *string                     `json:"transcriptUrl,omitempty"`
	HasTranscript      bool                        `json:"hasTranscript"`
	CreatedAt          string                      `json:"createdAt,omitempty"`
	UpdatedAt          string                      `json:"updatedAt,omitempty"`
	Status             string                      `json:"status"`
	LatestJob          *episodeJobResponse         `json:"latestJob,omitempty"`
	TranscriptSegments []transcriptSegmentResponse `json:"transcriptSegments"`
	SummarySegments    []summarySegmentResponse    `json:"summarySegments"`
	Paragraphs         []paragraph                 `json:"paragraphs"`
}

type episodeStatusResponse struct {
	PodcastID     string              `json:"podcastId"`
	ID            string              `json:"id"`
	URL           string              `json:"url"`
	Title         string              `json:"title"`
	HasTranscript bool                `json:"hasTranscript"`
	Status        string              `json:"status"`
	LatestJob     *episodeJobResponse `json:"latestJob,omitempty"`
}

// EpisodeDetailHandler returns episode metadata plus transcript-summary mapping data.
func EpisodeDetailHandler(episodes EpisodeRepository, jobs LatestJobRepository, processing ProcessingReadRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		episodeID := podcastIDFromPath(r.URL.Path)
		if episodeID == "" {
			http.NotFound(w, r)
			return
		}
		episode, latestJob, ok := loadEpisodeStatus(w, episodeID, episodes, jobs)
		if !ok {
			return
		}

		transcripts, err := processing.ListTranscriptSegments(episode.ID)
		if err != nil {
			http.Error(w, "unable to load transcript segments", http.StatusInternalServerError)
			return
		}
		summaryMappings, err := processing.GetSummarySourceMappings(episode.ID)
		if err != nil {
			http.Error(w, "unable to load summary mappings", http.StatusInternalServerError)
			return
		}

		resp := buildEpisodeDetailResponse(episode, latestJob, transcripts, summaryMappings)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// EpisodeStatusHandler returns episode metadata and latest processing job state.
func EpisodeStatusHandler(episodes EpisodeRepository, jobs LatestJobRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		episodeID := podcastIDFromPath(r.URL.Path)
		if episodeID == "" {
			http.NotFound(w, r)
			return
		}
		episode, latestJob, ok := loadEpisodeStatus(w, episodeID, episodes, jobs)
		if !ok {
			return
		}

		resp := episodeStatusResponse{
			PodcastID:     episode.ID,
			ID:            episode.ID,
			URL:           episode.PodcastURL,
			Title:         episode.Title,
			HasTranscript: episode.HasTranscript,
			Status:        episodeStatus(episode, latestJob),
			LatestJob:     jobResponse(latestJob),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func buildEpisodeDetailResponse(episode domain.Episode, latestJob *domain.ProcessingJob, transcripts []domain.TranscriptSegment, summaries []domain.SummarySourceMapping) episodeDetailResponse {
	summaryByTranscript := map[string][]string{}
	summaryResponses := make([]summarySegmentResponse, 0, len(summaries))
	for _, item := range summaries {
		sourceIDs := make([]string, 0, len(item.TranscriptSegments))
		for _, source := range item.TranscriptSegments {
			sourceIDs = append(sourceIDs, source.ID)
			summaryByTranscript[source.ID] = append(summaryByTranscript[source.ID], item.Summary.Text)
		}
		summaryResponses = append(summaryResponses, summarySegmentResponse{
			ID:                         item.Summary.ID,
			OrderIndex:                 item.Summary.OrderIndex,
			Text:                       item.Summary.Text,
			Provider:                   item.Summary.Provider,
			Model:                      item.Summary.Model,
			CreatedAt:                  formatOptionalTime(item.Summary.CreatedAt),
			SourceTranscriptSegmentIDs: sourceIDs,
		})
	}

	transcriptResponses := make([]transcriptSegmentResponse, 0, len(transcripts))
	paragraphs := make([]paragraph, 0, len(transcripts))
	for _, item := range transcripts {
		transcriptResponses = append(transcriptResponses, transcriptSegmentResponse{
			ID:           item.ID,
			OrderIndex:   item.OrderIndex,
			Text:         item.Text,
			AudioChunkID: item.AudioChunkID,
			StartSeconds: item.StartSeconds,
			EndSeconds:   item.EndSeconds,
			Provider:     item.Provider,
			Model:        item.Model,
			CreatedAt:    formatOptionalTime(item.CreatedAt),
		})
		paragraphs = append(paragraphs, paragraph{
			ParagraphID: item.ID,
			OrderIndex:  item.OrderIndex,
			Text:        item.Text,
			Timestamp:   item.StartSeconds,
			Summary:     strings.Join(summaryByTranscript[item.ID], "\n\n"),
		})
	}

	return episodeDetailResponse{
		PodcastID:          episode.ID,
		ID:                 episode.ID,
		URL:                episode.PodcastURL,
		Title:              episode.Title,
		Description:        episode.Description,
		DurationSeconds:    episode.DurationSeconds,
		AudioURL:           episode.AudioURL,
		TranscriptURL:      episode.TranscriptURL,
		HasTranscript:      episode.HasTranscript,
		CreatedAt:          formatOptionalTime(episode.CreatedAt),
		UpdatedAt:          formatOptionalTime(episode.UpdatedAt),
		Status:             episodeStatus(episode, latestJob),
		LatestJob:          jobResponse(latestJob),
		TranscriptSegments: transcriptResponses,
		SummarySegments:    summaryResponses,
		Paragraphs:         paragraphs,
	}
}

func loadEpisodeStatus(w http.ResponseWriter, episodeID string, episodes EpisodeRepository, jobs LatestJobRepository) (domain.Episode, *domain.ProcessingJob, bool) {
	episode, err := episodes.GetEpisode(episodeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "episode not found", http.StatusNotFound)
			return domain.Episode{}, nil, false
		}
		http.Error(w, "unable to load episode", http.StatusInternalServerError)
		return domain.Episode{}, nil, false
	}

	var latestJob *domain.ProcessingJob
	if jobs != nil {
		job, err := jobs.GetLatestByPodcastID(episode.ID)
		if err == nil {
			latestJob = &job
		} else if !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "unable to load episode status", http.StatusInternalServerError)
			return domain.Episode{}, nil, false
		}
	}
	return episode, latestJob, true
}

func episodeStatus(episode domain.Episode, latestJob *domain.ProcessingJob) string {
	if latestJob != nil && latestJob.Status != "" {
		return latestJob.Status
	}
	if episode.HasTranscript {
		return "succeeded"
	}
	return "not_started"
}

func jobResponse(job *domain.ProcessingJob) *episodeJobResponse {
	if job == nil {
		return nil
	}
	resp := &episodeJobResponse{
		JobID:        job.ID,
		PodcastID:    job.PodcastID,
		Type:         job.Type,
		Status:       job.Status,
		DurationMs:   job.DurationMs,
		ErrorMessage: job.Error,
	}
	if !job.StartedAt.IsZero() {
		resp.StartedAt = sqliteutil.FormatTime(job.StartedAt)
	}
	if job.CompletedAt != nil && !job.CompletedAt.IsZero() {
		resp.CompletedAt = sqliteutil.FormatTime(*job.CompletedAt)
	}
	return resp
}

func podcastIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 || parts[0] != "api" || parts[1] != "podcasts" {
		return ""
	}
	return parts[2]
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return sqliteutil.FormatTime(value)
}
