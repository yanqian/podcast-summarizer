package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"podcast-summarizer/src/core/domain"
)

func TestOpenAISummarizerRequestsGroupedSummaryMappings(t *testing.T) {
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", req.Method)
		}
		if req.URL.String() != "https://api.openai.test/v1/chat/completions" {
			t.Fatalf("endpoint = %s", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization header = %q", got)
		}
		var payload struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Temperature float64 `json:"temperature"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Model != "gpt-4o-mini" {
			t.Fatalf("model = %q", payload.Model)
		}
		if len(payload.Messages) != 2 || payload.Messages[1].Role != "user" {
			t.Fatalf("unexpected messages: %+v", payload.Messages)
		}
		if !strings.Contains(payload.Messages[1].Content, "seg-1") || !strings.Contains(payload.Messages[1].Content, "source_ids") {
			t.Fatalf("prompt missing transcript ids or schema: %s", payload.Messages[1].Content)
		}

		return jsonResponse([]byte(`{
			"choices": [{
				"message": {
					"content": "[{\"summary\":\"Opening context\",\"source_ids\":[\"seg-1\",\"seg-2\"]},{\"summary\":\"Decision and follow-up\",\"source_ids\":[\"seg-3\"]}]"
				}
			}]
		}`)), nil
	})

	client := NewOpenAISummarizer("test-key", "gpt-4o-mini")
	client.Endpoint = "https://api.openai.test/v1/chat/completions"
	client.Client = &http.Client{Transport: transport}

	result, err := client.SummarizeTranscriptSegments(context.Background(), []domain.TranscriptSegment{
		{ID: "seg-1", OrderIndex: 1, Text: "The host introduces the problem."},
		{ID: "seg-2", OrderIndex: 2, Text: "The guest explains the constraints."},
		{ID: "seg-3", OrderIndex: 3, Text: "They agree on the next step."},
	})
	if err != nil {
		t.Fatalf("summarize segments: %v", err)
	}
	if result.Provider != "openai" || result.Model != "gpt-4o-mini" {
		t.Fatalf("provider metadata mismatch: %+v", result)
	}
	if len(result.Segments) != 2 {
		t.Fatalf("expected 2 summary segments, got %+v", result.Segments)
	}
	if result.Segments[0].Text != "Opening context" {
		t.Fatalf("unexpected first summary: %+v", result.Segments[0])
	}
	if strings.Join(result.Segments[0].SourceTranscriptSegmentIDs, ",") != "seg-1,seg-2" {
		t.Fatalf("unexpected first mapping ids: %+v", result.Segments[0].SourceTranscriptSegmentIDs)
	}
	if strings.Join(result.Segments[1].SourceTranscriptSegmentIDs, ",") != "seg-3" {
		t.Fatalf("unexpected second mapping ids: %+v", result.Segments[1].SourceTranscriptSegmentIDs)
	}
}

func TestOpenAISummarizerRequiresAPIKeyForSegmentSummaries(t *testing.T) {
	client := NewOpenAISummarizer("", "gpt-4o-mini")
	_, err := client.SummarizeTranscriptSegments(context.Background(), []domain.TranscriptSegment{
		{ID: "seg-1", OrderIndex: 1, Text: "Transcript."},
	})
	if err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("expected API key error, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(body []byte) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}
