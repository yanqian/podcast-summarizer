package app

import (
	"database/sql"
	"testing"

	"podcast-summarizer/src/core/domain"
)

func TestExtractIDs(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantTrackID string
		wantShowID  string
	}{
		{
			name:        "apple podcasts episode url",
			url:         "https://podcasts.apple.com/us/podcast/ilya-sutskever-were-moving-from-the-age-of-scaling/id1516093381?i=1000738363711",
			wantTrackID: "1000738363711",
			wantShowID:  "1516093381",
		},
		{
			name:        "fallback when missing ids",
			url:         "https://example.com/feed/episode.mp3",
			wantTrackID: "",
			wantShowID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackID, showID := extractIDs(tt.url)
			if trackID != tt.wantTrackID {
				t.Fatalf("trackID got %q want %q", trackID, tt.wantTrackID)
			}
			if showID != tt.wantShowID {
				t.Fatalf("showID got %q want %q", showID, tt.wantShowID)
			}
		})
	}
}

func TestIngestReturnsExistingPodcastWithoutLookupOrUpsert(t *testing.T) {
	latestJobID := "job-1"
	latestStatus := "queued"
	repo := &fakePodcastRepo{
		source: domain.PodcastSource{
			ID:           "podcast-1",
			URL:          "https://example.com/episode.mp3",
			Title:        "Existing",
			LatestJobID:  &latestJobID,
			LatestStatus: &latestStatus,
		},
	}
	lookup := &fakeLookup{}
	svc := NewIngestService(lookup, nil, repo, nil)

	result, err := svc.Ingest(" https://example.com/episode.mp3 ")
	if err != nil {
		t.Fatalf("ingest existing: %v", err)
	}
	if !result.Existing || result.PodcastID != "podcast-1" {
		t.Fatalf("expected existing podcast result, got %+v", result)
	}
	if lookup.calls != 0 {
		t.Fatalf("duplicate ingest should not call lookup, got %d calls", lookup.calls)
	}
	if repo.upserts != 0 {
		t.Fatalf("duplicate ingest should not upsert metadata, got %d upserts", repo.upserts)
	}
	if result.Source.LatestJobID == nil || *result.Source.LatestJobID != latestJobID {
		t.Fatalf("expected latest job in duplicate result, got %+v", result.Source)
	}
}

func TestIngestCreatesPodcastWhenURLIsNew(t *testing.T) {
	repo := &fakePodcastRepo{}
	lookup := &fakeLookup{
		meta: &domain.PodcastMetadata{
			TrackID:  "100",
			Title:    "New Episode",
			Artist:   "Host",
			AudioURL: "https://example.com/audio.mp3",
		},
	}
	svc := NewIngestService(lookup, nil, repo, nil)

	result, err := svc.Ingest("https://podcasts.apple.com/us/podcast/demo/id1516093381?i=100")
	if err != nil {
		t.Fatalf("ingest new: %v", err)
	}
	if result.Existing || result.PodcastID != "new-podcast" {
		t.Fatalf("expected new podcast result, got %+v", result)
	}
	if lookup.calls != 1 {
		t.Fatalf("expected one lookup for new Apple URL, got %d", lookup.calls)
	}
	if repo.upserts != 1 {
		t.Fatalf("expected one upsert for new URL, got %d", repo.upserts)
	}
}

type fakeLookup struct {
	calls int
	meta  *domain.PodcastMetadata
}

func (l *fakeLookup) Lookup(trackID string) (*domain.PodcastMetadata, error) {
	l.calls++
	if l.meta != nil {
		return l.meta, nil
	}
	return &domain.PodcastMetadata{
		TrackID:  trackID,
		Title:    "Lookup Episode",
		Artist:   "Lookup Host",
		AudioURL: "https://example.com/audio.mp3",
	}, nil
}

type fakePodcastRepo struct {
	source  domain.PodcastSource
	upserts int
}

func (r *fakePodcastRepo) UpsertSource(url, title, description string, durationSeconds *int, audioURL *string, transcriptURL *string) (string, error) {
	r.upserts++
	return "new-podcast", nil
}

func (r *fakePodcastRepo) GetSourceByURL(url string) (domain.PodcastSource, error) {
	if r.source.URL == url {
		return r.source, nil
	}
	return domain.PodcastSource{}, sql.ErrNoRows
}

func (r *fakePodcastRepo) MarkHasTranscript(id string) error {
	return nil
}

func (r *fakePodcastRepo) ListSources(limit int) ([]domain.PodcastSource, error) {
	return nil, nil
}
