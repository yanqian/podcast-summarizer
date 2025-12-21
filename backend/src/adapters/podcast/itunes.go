package podcast

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"podcast-summarizer/src/core/domain"
)

type LookupResponse struct {
	ResultCount int           `json:"resultCount"`
	Results     []LookupEntry `json:"results"`
}

type LookupEntry struct {
	TrackID    int64  `json:"trackId"`
	TrackName  string `json:"trackName"`
	ArtistName string `json:"artistName"`
	PreviewURL string `json:"previewUrl"`
	FeedURL    string `json:"feedUrl"`
	EpisodeURL string `json:"episodeUrl"`
	// Some feeds expose captions/transcripts via additional fields; keep flexible map.
	Raw map[string]any `json:"-"`
}

// ItunesClient fetches podcast metadata via iTunes lookup.
type ItunesClient struct {
	HTTP *http.Client
}

func NewItunesClient() *ItunesClient {
	return &ItunesClient{
		HTTP: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *ItunesClient) Lookup(trackID string) (*domain.PodcastMetadata, error) {
	url := fmt.Sprintf("https://itunes.apple.com/lookup?id=%s", trackID)
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lookup failed: %s", resp.Status)
	}
	var lr LookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, err
	}
	if lr.ResultCount == 0 || len(lr.Results) == 0 {
		return nil, fmt.Errorf("no results for track id %s", trackID)
	}
	entry := lr.Results[0]
	// Capture raw for potential transcript fields.
	entry.Raw = map[string]any{}
	meta := &domain.PodcastMetadata{
		TrackID:  fmt.Sprintf("%d", entry.TrackID),
		Title:    entry.TrackName,
		Artist:   entry.ArtistName,
		AudioURL: entry.PreviewURL,
	}
	// If transcript URL exists in raw, set it (not common in lookup).
	if entry.PreviewURL == "" && entry.FeedURL != "" {
		meta.AudioURL = entry.FeedURL
	}
	return meta, nil
}

// LookupEpisode resolves metadata for a specific episode by show and episode IDs.
func (c *ItunesClient) LookupEpisode(showID, episodeID string) (*domain.PodcastMetadata, error) {
	url := fmt.Sprintf("https://itunes.apple.com/lookup?id=%s&entity=podcastEpisode&limit=200", showID)
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lookup failed: %s", resp.Status)
	}
	var lr LookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, err
	}
	if lr.ResultCount == 0 || len(lr.Results) == 0 {
		return nil, fmt.Errorf("no results for show id %s", showID)
	}

	var match *LookupEntry
	for i := range lr.Results {
		entry := lr.Results[i]
		if fmt.Sprintf("%d", entry.TrackID) == episodeID {
			match = &entry
			break
		}
	}
	if match == nil {
		return nil, fmt.Errorf("no episode %s under show %s", episodeID, showID)
	}

	audioURL := match.EpisodeURL
	if audioURL == "" {
		audioURL = match.PreviewURL
	}
	if audioURL == "" {
		audioURL = match.FeedURL
	}
	if audioURL == "" {
		return nil, fmt.Errorf("episode %s has no audio URL", episodeID)
	}

	meta := &domain.PodcastMetadata{
		TrackID:  fmt.Sprintf("%d", match.TrackID),
		Title:    match.TrackName,
		Artist:   match.ArtistName,
		AudioURL: audioURL,
	}
	return meta, nil
}
