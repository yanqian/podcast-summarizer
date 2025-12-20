package contract

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"podcast-summarizer/src/api"
)

func TestViewAlignment(t *testing.T) {
	router := api.NewRouter(api.Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/api/podcasts/demo/view", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(rr.Body.Bytes()) == 0 {
		t.Fatalf("expected response body")
	}
}
