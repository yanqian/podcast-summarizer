package integration

import (
	"testing"

	"podcast-summarizer/src/core/app/jobs"
)

// Placeholder integration test to ensure pipeline wiring compiles.
func TestTranscriptionPipelineStub(t *testing.T) {
	worker := jobs.NewTranscriptionWorker(nil, nil, nil, "")
	if worker == nil {
		t.Fatalf("expected worker")
	}
}
