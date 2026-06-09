package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"podcast-summarizer/src/core/app/jobs"
	"podcast-summarizer/src/core/domain"
	dbinfra "podcast-summarizer/src/infra/db"
	storageinfra "podcast-summarizer/src/infra/storage"
	jobrepo "podcast-summarizer/src/repo/job"
	podcastrepo "podcast-summarizer/src/repo/podcast"
	processingrepo "podcast-summarizer/src/repo/processing"
)

func TestPodcastProcessingPipelineStagesAndIdempotentArtifacts(t *testing.T) {
	ctx := context.Background()
	db, err := dbinfra.NewSQLite(ctx, filepath.Join(t.TempDir(), "podcast.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	audioURL := "http://audio.test/episode.mp3"
	podcastID, err := podcastrepo.NewPodcastSQLiteRepo(db).UpsertSource(
		"http://podcast.test/episode",
		"Fixture Episode",
		"Fixture Host",
		nil,
		&audioURL,
		nil,
	)
	if err != nil {
		t.Fatalf("upsert episode: %v", err)
	}

	jobRepo := &recordingJobRepo{inner: jobrepo.NewJobSQLiteRepo(db)}
	processingRepo := processingrepo.NewProcessingSQLiteRepo(db)
	manager := jobs.NewManagerWithArtifacts(
		jobRepo,
		&integrationLocker{},
		nil,
		jobs.NewPipelineTranscriptProvider(
			nil,
			&integrationDownloader{dir: t.TempDir()},
			&integrationChunker{dir: t.TempDir()},
			&integrationTranscriber{},
		),
		jobs.NewClientSummaryService(&integrationSummaryClient{}),
		jobs.NewObjectStoragePublisher(storageinfra.NewLocalUploader(filepath.Join(t.TempDir(), "storage"))),
		processingRepo,
	)

	firstJobID, err := manager.StartJob(ctx, podcastID, audioURL, nil)
	if err != nil {
		t.Fatalf("start first job: %v", err)
	}
	firstJob := waitForIntegrationJobStatus(t, jobRepo, firstJobID, "succeeded")
	if firstJob.Error != nil {
		t.Fatalf("first job unexpectedly failed: %s", *firstJob.Error)
	}
	assertStatusSequence(t, jobRepo.history(firstJobID), []string{
		"running",
		"downloading",
		"chunking",
		"transcribing",
		"summarizing",
		"succeeded",
	})
	assertPipelineArtifactCounts(t, db, podcastID, pipelineCounts{
		audioChunks:        2,
		transcriptSegments: 2,
		summarySegments:    1,
		mappings:           2,
	})

	secondJobID, err := manager.StartJob(ctx, podcastID, audioURL, nil)
	if err != nil {
		t.Fatalf("start second job: %v", err)
	}
	secondJob := waitForIntegrationJobStatus(t, jobRepo, secondJobID, "succeeded")
	if secondJob.Error != nil {
		t.Fatalf("second job unexpectedly failed: %s", *secondJob.Error)
	}
	assertPipelineArtifactCounts(t, db, podcastID, pipelineCounts{
		audioChunks:        2,
		transcriptSegments: 2,
		summarySegments:    1,
		mappings:           2,
	})

	mappings, err := processingRepo.GetSummarySourceMappings(podcastID)
	if err != nil {
		t.Fatalf("get summary mappings: %v", err)
	}
	if len(mappings) != 1 || len(mappings[0].TranscriptSegments) != 2 {
		t.Fatalf("expected one grouped summary mapped to two transcript segments, got %+v", mappings)
	}
	if !strings.Contains(mappings[0].Summary.Text, "fixture summary") {
		t.Fatalf("unexpected grouped summary text: %+v", mappings[0].Summary)
	}
}

type recordingJobRepo struct {
	inner jobs.JobRepository
	mu    sync.Mutex
	seen  map[string][]string
}

func (r *recordingJobRepo) Create(job domain.ProcessingJob) (string, error) {
	id, err := r.inner.Create(job)
	if err != nil {
		return "", err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.seen == nil {
		r.seen = map[string][]string{}
	}
	r.seen[id] = append(r.seen[id], job.Status)
	return id, nil
}

func (r *recordingJobRepo) UpdateStatus(id, status string, durationMs *int, errorMessage *string) error {
	if err := r.inner.UpdateStatus(id, status, durationMs, errorMessage); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.seen == nil {
		r.seen = map[string][]string{}
	}
	r.seen[id] = append(r.seen[id], status)
	return nil
}

func (r *recordingJobRepo) Get(id string) (domain.ProcessingJob, error) {
	return r.inner.Get(id)
}

func (r *recordingJobRepo) history(id string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.seen[id]...)
}

type integrationLocker struct{}

func (l *integrationLocker) Acquire(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (l *integrationLocker) Release(ctx context.Context, key string) error {
	return nil
}

type integrationDownloader struct {
	dir string
}

func (d *integrationDownloader) Download(url string) (string, error) {
	path := filepath.Join(d.dir, "episode-original.mp3")
	if err := os.WriteFile(path, []byte("fixture original audio"), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

type integrationChunker struct {
	dir string
}

func (c *integrationChunker) Chunk(inputPath string) ([]string, error) {
	paths := []string{
		filepath.Join(c.dir, "episode-chunk-001.mp3"),
		filepath.Join(c.dir, "episode-chunk-002.mp3"),
	}
	for idx, path := range paths {
		if err := os.WriteFile(path, []byte(fmt.Sprintf("fixture chunk %d", idx+1)), 0o644); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

type integrationTranscriber struct{}

func (t *integrationTranscriber) TranscribeFile(path string) (string, error) {
	return "transcribed " + filepath.Base(path), nil
}

func (t *integrationTranscriber) TranscribeFileDetailed(path string) (domain.TranscriptionResult, error) {
	return domain.TranscriptionResult{
		Text:     "transcribed " + filepath.Base(path),
		Provider: "fixture",
		Model:    "fixture-model",
	}, nil
}

type integrationSummaryClient struct{}

func (c *integrationSummaryClient) Summarize(paragraphs []string) ([]string, error) {
	summaries := make([]string, len(paragraphs))
	for idx, paragraph := range paragraphs {
		summaries[idx] = "fixture summary: " + paragraph
	}
	return summaries, nil
}

func waitForIntegrationJobStatus(t *testing.T, repo jobs.JobRepository, jobID, want string) domain.ProcessingJob {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var last domain.ProcessingJob
	for time.Now().Before(deadline) {
		job, err := repo.Get(jobID)
		if err != nil {
			t.Fatalf("get job: %v", err)
		}
		last = job
		if job.Status == want {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for job %s to reach %s; last=%+v", jobID, want, last)
	return domain.ProcessingJob{}
}

func assertStatusSequence(t *testing.T, got []string, want []string) {
	t.Helper()
	next := 0
	for _, status := range got {
		if next < len(want) && status == want[next] {
			next++
		}
	}
	if next != len(want) {
		t.Fatalf("missing status sequence %v in history %v", want, got)
	}
}

type pipelineCounts struct {
	audioChunks        int
	transcriptSegments int
	summarySegments    int
	mappings           int
}

func assertPipelineArtifactCounts(t *testing.T, db *sql.DB, episodeID string, want pipelineCounts) {
	t.Helper()
	got := pipelineCounts{
		audioChunks:        countRows(t, db, "audio_chunk", episodeID),
		transcriptSegments: countRows(t, db, "transcript_segment", episodeID),
		summarySegments:    countRows(t, db, "summary_segment", episodeID),
		mappings:           countRows(t, db, "transcript_summary_mapping", episodeID),
	}
	if got != want {
		t.Fatalf("artifact counts mismatch: got %+v want %+v", got, want)
	}
}

func countRows(t *testing.T, db *sql.DB, table string, episodeID string) int {
	t.Helper()
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE episode_id = ?", table)
	if err := db.QueryRow(query, episodeID).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}
