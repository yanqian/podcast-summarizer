package transcription

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAITranscriberRequestsVerboseJSONAndParsesSegments(t *testing.T) {
	audioPath := filepath.Join(t.TempDir(), "chunk.mp3")
	if err := os.WriteFile(audioPath, []byte("fixture audio"), 0o644); err != nil {
		t.Fatalf("write fixture audio: %v", err)
	}

	handler := func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization header = %q", got)
		}
		form, err := readMultipartForm(r)
		if err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		assertFormValue(t, form, "model", "whisper-1")
		assertFormValue(t, form, "response_format", "verbose_json")
		assertFormValue(t, form, "timestamp_granularities[]", "segment")
		files := form.File["file"]
		if len(files) != 1 || files[0].Filename != "chunk.mp3" {
			t.Fatalf("unexpected uploaded files: %+v", files)
		}

		body, err := json.Marshal(map[string]any{
			"text": "First sentence. Second sentence.",
			"segments": []map[string]any{
				{"start": 0.0, "end": 1.25, "text": "First sentence."},
				{"start": 1.25, "end": 2.5, "text": "Second sentence."},
			},
		})
		if err != nil {
			t.Fatalf("marshal response: %v", err)
		}
		return jsonResponse(body), nil
	}

	client := NewOpenAITranscriber("test-key", "whisper-1")
	client.Endpoint = "https://api.openai.test/v1/audio/transcriptions"
	client.Client = &http.Client{Transport: roundTripFunc(handler)}

	result, err := client.TranscribeFileDetailed(audioPath)
	if err != nil {
		t.Fatalf("transcribe file: %v", err)
	}
	if result.Text != "First sentence. Second sentence." {
		t.Fatalf("text = %q", result.Text)
	}
	if result.Provider != "openai" || result.Model != "whisper-1" {
		t.Fatalf("provider/model mismatch: %+v", result)
	}
	if len(result.Segments) != 2 {
		t.Fatalf("segments = %+v", result.Segments)
	}
	if result.Segments[0].StartSeconds == nil || *result.Segments[0].StartSeconds != 0.0 {
		t.Fatalf("first segment start not parsed: %+v", result.Segments[0])
	}
	if result.Segments[1].EndSeconds == nil || *result.Segments[1].EndSeconds != 2.5 {
		t.Fatalf("second segment end not parsed: %+v", result.Segments[1])
	}
}

func TestOpenAITranscriberUsesJSONForModernTranscribeModels(t *testing.T) {
	audioPath := filepath.Join(t.TempDir(), "chunk.mp3")
	if err := os.WriteFile(audioPath, []byte("fixture audio"), 0o644); err != nil {
		t.Fatalf("write fixture audio: %v", err)
	}

	handler := func(r *http.Request) (*http.Response, error) {
		form, err := readMultipartForm(r)
		if err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		assertFormValue(t, form, "model", "gpt-4o-mini-transcribe")
		assertFormValue(t, form, "response_format", "json")
		if got := form.Value["timestamp_granularities[]"]; len(got) != 0 {
			t.Fatalf("timestamp_granularities[] should be omitted, got %q", got)
		}
		return jsonResponse([]byte(`{"text":"Modern model transcript."}`)), nil
	}

	client := NewOpenAITranscriber("test-key", "gpt-4o-mini-transcribe")
	client.Endpoint = "https://api.openai.test/v1/audio/transcriptions"
	client.Client = &http.Client{Transport: roundTripFunc(handler)}

	text, err := client.TranscribeFile(audioPath)
	if err != nil {
		t.Fatalf("transcribe file: %v", err)
	}
	if text != "Modern model transcript." {
		t.Fatalf("text = %q", text)
	}
}

func TestOpenAITranscriberRequiresAPIKey(t *testing.T) {
	audioPath := filepath.Join(t.TempDir(), "chunk.mp3")
	if err := os.WriteFile(audioPath, []byte("fixture audio"), 0o644); err != nil {
		t.Fatalf("write fixture audio: %v", err)
	}
	client := NewOpenAITranscriber("", "whisper-1")
	if _, err := client.TranscribeFile(audioPath); err == nil {
		t.Fatalf("expected missing API key error")
	}
}

func assertFormValue(t *testing.T, form *multipart.Form, key, want string) {
	t.Helper()
	got := form.Value[key]
	if len(got) != 1 || got[0] != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func readMultipartForm(r *http.Request) (*multipart.Form, error) {
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	reader := multipart.NewReader(r.Body, params["boundary"])
	return reader.ReadForm(1 << 20)
}

func jsonResponse(body []byte) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(string(body))),
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
