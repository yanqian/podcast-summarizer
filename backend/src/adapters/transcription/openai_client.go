package transcription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"podcast-summarizer/src/core/domain"
)

const openAITranscriptionsEndpoint = "https://api.openai.com/v1/audio/transcriptions"

// OpenAITranscriber calls OpenAI audio transcription API.
type OpenAITranscriber struct {
	APIKey   string
	Model    string
	Endpoint string
	Client   *http.Client
}

func NewOpenAITranscriber(apiKey, model string) *OpenAITranscriber {
	if model == "" {
		model = "whisper-1"
	}
	return &OpenAITranscriber{
		APIKey:   apiKey,
		Model:    model,
		Endpoint: openAITranscriptionsEndpoint,
		Client:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (t *OpenAITranscriber) TranscribeFile(path string) (string, error) {
	result, err := t.TranscribeFileDetailed(path)
	if err != nil {
		return "", err
	}
	return result.Text, nil
}

func (t *OpenAITranscriber) TranscribeFileDetailed(path string) (domain.TranscriptionResult, error) {
	if strings.TrimSpace(t.APIKey) == "" {
		return domain.TranscriptionResult{}, fmt.Errorf("openai transcription requires OPENAI_API_KEY")
	}
	if strings.TrimSpace(t.Model) == "" {
		return domain.TranscriptionResult{}, fmt.Errorf("openai transcription requires a model")
	}
	f, err := os.Open(path)
	if err != nil {
		return domain.TranscriptionResult{}, err
	}
	defer func() {
		_ = f.Close()
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("model", t.Model); err != nil {
		return domain.TranscriptionResult{}, err
	}
	responseFormat := responseFormatForModel(t.Model)
	if err := writer.WriteField("response_format", responseFormat); err != nil {
		return domain.TranscriptionResult{}, err
	}
	if responseFormat == "verbose_json" {
		if err := writer.WriteField("timestamp_granularities[]", "segment"); err != nil {
			return domain.TranscriptionResult{}, err
		}
	}
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return domain.TranscriptionResult{}, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return domain.TranscriptionResult{}, err
	}
	if err := writer.Close(); err != nil {
		return domain.TranscriptionResult{}, err
	}

	endpoint := t.Endpoint
	if endpoint == "" {
		endpoint = openAITranscriptionsEndpoint
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return domain.TranscriptionResult{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+t.APIKey)

	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return domain.TranscriptionResult{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return domain.TranscriptionResult{}, fmt.Errorf("openai transcription failed: %s (%s)", resp.Status, string(b))
	}
	var out struct {
		Text     string `json:"text"`
		Segments []struct {
			Start *float64 `json:"start"`
			End   *float64 `json:"end"`
			Text  string   `json:"text"`
		} `json:"segments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return domain.TranscriptionResult{}, err
	}
	segments := make([]domain.TranscriptionResultSegment, 0, len(out.Segments))
	for _, segment := range out.Segments {
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			continue
		}
		segments = append(segments, domain.TranscriptionResultSegment{
			OrderIndex:   len(segments) + 1,
			StartSeconds: segment.Start,
			EndSeconds:   segment.End,
			Text:         text,
		})
	}
	text := strings.TrimSpace(out.Text)
	if text == "" && len(segments) > 0 {
		parts := make([]string, 0, len(segments))
		for _, segment := range segments {
			parts = append(parts, segment.Text)
		}
		text = strings.Join(parts, " ")
	}
	if text == "" {
		return domain.TranscriptionResult{}, fmt.Errorf("openai transcription returned empty text")
	}
	return domain.TranscriptionResult{
		Text:     text,
		Segments: segments,
		Provider: "openai",
		Model:    t.Model,
	}, nil
}

func responseFormatForModel(model string) string {
	if model == "whisper-1" || strings.HasPrefix(model, "whisper-1-") {
		return "verbose_json"
	}
	return "json"
}
