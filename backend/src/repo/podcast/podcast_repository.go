package podcastrepo

import (
	"errors"

	"podcast-summarizer/src/core/domain"
)

type PodcastRepository struct{}

func NewPodcastRepository() *PodcastRepository {
	return &PodcastRepository{}
}

func (r *PodcastRepository) UpsertSource(url, title, description string, durationSeconds *int, audioURL *string, transcriptURL *string) (string, error) {
	if url == "" {
		return "", errors.New("url required")
	}
	return "stub-podcast", nil
}

func (r *PodcastRepository) MarkHasTranscript(id string) error {
	if id == "" {
		return errors.New("id required")
	}
	return nil
}

func (r *PodcastRepository) ListSources(limit int) ([]domain.PodcastSource, error) {
	return []domain.PodcastSource{}, nil
}
