package transcription

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// HTTPClient posts audio files to an external transcription API.
// Expects a multipart form with "file" field; API key header optional.
type HTTPClient struct {
	URL        string
	Key        string
	HeaderName string
	Client     *http.Client
}

func NewHTTPClient(url, key string) *HTTPClient {
	return &HTTPClient{
		URL:        url,
		Key:        key,
		HeaderName: "Authorization",
		Client:     &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *HTTPClient) TranscribeFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", path)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.URL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.Key != "" {
		req.Header.Set(c.HeaderName, c.Key)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("transcription failed: %s (%s)", resp.Status, string(b))
	}
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
