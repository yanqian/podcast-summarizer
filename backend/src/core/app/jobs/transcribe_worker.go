package jobs

import (
	"context"

	"podcast-summarizer/src/core/domain"
)

// TranscriptionWorker coordinates transcription then summarization.
type TranscriptionWorker struct {
	Fetcher    domain.TranscriptFetcher
	Client     domain.TranscriptionClient
	Repo       domain.PodcastRepository
	FFmpegPath string
}

func NewTranscriptionWorker(fetcher domain.TranscriptFetcher, client domain.TranscriptionClient, repo domain.PodcastRepository, ffmpegPath string) *TranscriptionWorker {
	return &TranscriptionWorker{Fetcher: fetcher, Client: client, Repo: repo, FFmpegPath: ffmpegPath}
}

func (w *TranscriptionWorker) Run(ctx context.Context, url string) (string, error) {
	if w.Client == nil {
		return "", nil
	}
	_, err := w.Client.Transcribe(url)
	if err != nil {
		return "", err
	}
	return "stub-podcast", nil
}
