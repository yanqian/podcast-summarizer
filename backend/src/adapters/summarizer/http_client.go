package summarizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient struct {
	URL        string
	Key        string
	HeaderName string
	Client     *http.Client
}

func NewHTTPSummarizer(url, key string) *HTTPClient {
	return &HTTPClient{
		URL:        url,
		Key:        key,
		HeaderName: "Authorization",
		Client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HTTPClient) Summarize(paragraphs []string) ([]string, error) {
	payload := map[string]any{"paragraphs": paragraphs}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Key != "" {
		req.Header.Set(c.HeaderName, c.Key)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("summarize failed: %s (%s)", resp.Status, string(body))
	}
	var out struct {
		Summaries []string `json:"summaries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Summaries) == 0 {
		return nil, fmt.Errorf("empty summaries")
	}
	return out.Summaries, nil
}
