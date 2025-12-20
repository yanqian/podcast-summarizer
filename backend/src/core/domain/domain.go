package domain

import "time"

// Adapter and repository contracts.

type TranscriptFetcher interface {
	Fetch(url string) ([]byte, error)
}

type PodcastLookup interface {
	Lookup(trackID string) (*PodcastMetadata, error)
}

// PodcastEpisodeLookup allows fetching episode metadata when only show+episode IDs are known.
type PodcastEpisodeLookup interface {
	LookupEpisode(showID, episodeID string) (*PodcastMetadata, error)
}

type TranscriptionClient interface {
	Transcribe(url string) ([]byte, error)
}

type SummaryClient interface {
	Summarize(paragraphs []string) ([]string, error)
}

type PodcastRepository interface {
	UpsertSource(url, title, description string, durationSeconds *int, audioURL *string, transcriptURL *string) (string, error)
	MarkHasTranscript(id string) error
	ListSources(limit int) ([]PodcastSource, error)
}

type ParagraphRepository interface {
	SaveTranscript(podcastID string, paragraphs []Paragraph) error
	SaveSummaries(podcastID string, summaries []Summary) error
	GetAligned(podcastID string) ([]ParagraphWithSummary, error)
}

type JobRepository interface {
	Create(job ProcessingJob) (string, error)
	UpdateStatus(id, status string, durationMs *int, errorMessage *string) error
	Get(id string) (ProcessingJob, error)
}

type Cache interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}

type Paragraph struct {
	OrderIndex int
	Text       string
}

type Summary struct {
	OrderIndex int
	Text       string
}

type ParagraphWithSummary struct {
	OrderIndex int
	Text       string
	Summary    string
}

type PodcastMetadata struct {
	TrackID       string
	Title         string
	Artist        string
	AudioURL      string
	TranscriptURL *string
}

type PodcastSource struct {
	ID            string    `json:"id"`
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	HasTranscript bool      `json:"hasTranscript"`
	CreatedAt     time.Time `json:"createdAt"`
	LatestJobID   *string   `json:"latestJobId"`
	LatestStatus  *string   `json:"latestStatus"`
}

type ProcessingJob struct {
	ID         string
	PodcastID  string
	Type       string
	Status     string
	DurationMs *int
	Error      *string
}
