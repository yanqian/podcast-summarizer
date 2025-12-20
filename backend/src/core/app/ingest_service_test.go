package app

import "testing"

func TestExtractIDs(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		wantTrackID   string
		wantShowID    string
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
