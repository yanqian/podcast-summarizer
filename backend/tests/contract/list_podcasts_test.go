package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPodcastsEmptyResponseUsesArray(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/podcasts", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.Items == nil {
		t.Fatalf("expected items to be an empty array, got nil")
	}
	if len(resp.Items) != 0 {
		t.Fatalf("expected no items, got %d", len(resp.Items))
	}
}
