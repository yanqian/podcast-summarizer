package transcription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// OpenAITranscriber calls OpenAI audio transcription API.
type OpenAITranscriber struct {
	APIKey string
	Model  string
	Client *http.Client
}

func NewOpenAITranscriber(apiKey, model string) *OpenAITranscriber {
	return &OpenAITranscriber{
		APIKey: apiKey,
		Model:  model,
		Client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (t *OpenAITranscriber) TranscribeFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("model", t.Model); err != nil {
		return "", err
	}
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

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+t.APIKey)

	resp, err := t.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai transcription failed: %s (%s)", resp.Status, string(b))
	}
	var out struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Text == "" {
		return "", fmt.Errorf("openai transcription returned empty text")
	}
	return out.Text, nil
}
