package app

import (
	"database/sql"
	"errors"
	"log"
	neturl "net/url"
	"strings"

	"podcast-summarizer/src/core/domain"
)

// IngestService coordinates podcast lookup, transcript fetch, or audio download/transcription.
type IngestService struct {
	Lookup        domain.PodcastLookup
	Fetcher       domain.TranscriptFetcher
	Repo          domain.PodcastRepository
	ParagraphRepo domain.ParagraphRepository
}

type IngestResult struct {
	PodcastID string
	Metadata  *domain.PodcastMetadata
	Source    domain.PodcastSource
	Existing  bool
}

func NewIngestService(lookup domain.PodcastLookup, fetcher domain.TranscriptFetcher, repo domain.PodcastRepository, paraRepo domain.ParagraphRepository) *IngestService {
	return &IngestService{Lookup: lookup, Fetcher: fetcher, Repo: repo, ParagraphRepo: paraRepo}
}

// Ingest handles URL by extracting track ID, running lookup, and saving metadata; returns podcast ID.
func (s *IngestService) Ingest(rawURL string) (IngestResult, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return IngestResult{}, errors.New("url required")
	}

	existing, err := s.Repo.GetSourceByURL(rawURL)
	if err == nil {
		return IngestResult{
			PodcastID: existing.ID,
			Source:    existing,
			Existing:  true,
		}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return IngestResult{}, err
	}

	trackID, showID := extractIDs(rawURL)
	var meta *domain.PodcastMetadata
	if trackID == "" {
		// Fallback metadata when track id is missing; useful for tests or non-Apple URLs.
		meta = &domain.PodcastMetadata{
			TrackID:  "local",
			Title:    "Unknown",
			Artist:   "Unknown",
			AudioURL: rawURL,
		}
	} else {
		meta, err = s.Lookup.Lookup(trackID)
		if err != nil && showID != "" {
			if epLookup, ok := s.Lookup.(domain.PodcastEpisodeLookup); ok {
				meta, err = epLookup.LookupEpisode(showID, trackID)
			}
		}
		if err != nil {
			log.Printf("lookup failed for track %s: %v; using fallback metadata", trackID, err)
			meta = &domain.PodcastMetadata{
				TrackID:  trackID,
				Title:    "Unknown",
				Artist:   "Unknown",
				AudioURL: rawURL,
			}
		}
	}
	podcastID, err := s.Repo.UpsertSource(rawURL, meta.Title, meta.Artist, nil, &meta.AudioURL, meta.TranscriptURL)
	if err != nil {
		return IngestResult{}, err
	}
	return IngestResult{
		PodcastID: podcastID,
		Metadata:  meta,
		Source: domain.PodcastSource{
			ID:    podcastID,
			URL:   rawURL,
			Title: meta.Title,
		},
	}, nil
}

// extractIDs matches the Python helper: episode id from ?i=, show id from path id<digits>.
func extractIDs(rawURL string) (trackID, showID string) {
	if parsed, err := neturl.Parse(rawURL); err == nil {
		if id := parsed.Query().Get("i"); id != "" {
			trackID = id
		}
		segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		for _, seg := range segments {
			if strings.HasPrefix(seg, "id") && len(seg) > 2 && isNumeric(seg[2:]) {
				showID = seg[2:]
				break
			}
		}
	}
	if trackID == "" {
		if i := strings.Index(rawURL, "?i="); i != -1 {
			trackID = rawURL[i+3:]
		}
	}
	if showID == "" {
		parts := strings.Split(strings.Trim(rawURL, "/"), "/")
		for _, p := range parts {
			if strings.HasPrefix(p, "id") {
				id := strings.TrimPrefix(p, "id")
				if isNumeric(id) {
					showID = id
					break
				}
			}
		}
	}
	return
}

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
