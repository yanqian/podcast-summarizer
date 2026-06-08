package media

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestDownloadFileUsesHTTPResponseBody(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader("small audio fixture")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}

	path, err := downloadFileWithClient(client, "http://fixture.test/episode.mp3")
	if err != nil {
		t.Fatalf("download file: %v", err)
	}
	defer func() {
		_ = os.Remove(path)
	}()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(body) != "small audio fixture" {
		t.Fatalf("downloaded body mismatch: %q", body)
	}
}

func TestDownloadFileReturnsErrorForNonOKStatus(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Status:     "404 Not Found",
			Body:       io.NopCloser(strings.NewReader("missing")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}

	if _, err := downloadFileWithClient(client, "http://fixture.test/missing.mp3"); err == nil {
		t.Fatal("expected non-OK status to fail")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
