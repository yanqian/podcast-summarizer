package transcript

import (
	"errors"
	"io"
	"net/http"
	"time"
)

// Fetcher retrieves transcripts/captions via HTTP GET.
type Fetcher struct {
	client *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{client: &http.Client{Timeout: 15 * time.Second}}
}

func (f *Fetcher) Fetch(url string) ([]byte, error) {
	if url == "" {
		return nil, errors.New("missing url")
	}
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch transcript")
	}
	return io.ReadAll(resp.Body)
}
