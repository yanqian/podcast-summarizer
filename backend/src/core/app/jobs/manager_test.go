package jobs_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"podcast-summarizer/src/core/app/jobs"
	"podcast-summarizer/src/core/domain"
	dbinfra "podcast-summarizer/src/infra/db"
	storageinfra "podcast-summarizer/src/infra/storage"
	jobrepo "podcast-summarizer/src/repo/job"
	paragraphrepo "podcast-summarizer/src/repo/paragraph"
	podcastrepo "podcast-summarizer/src/repo/podcast"
	processingrepo "podcast-summarizer/src/repo/processing"
)

func TestManagerStoresLocalAudioArtifactsAndPersistsChunkMetadata(t *testing.T) {
	ctx := context.Background()
	db, err := dbinfra.NewSQLite(ctx, filepath.Join(t.TempDir(), "podcast.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	podcastID := createEpisode(t, db)
	jobRepo := jobrepo.NewJobSQLiteRepo(db)
	paragraphRepo := paragraphrepo.NewParagraphSQLiteRepo(db)
	processingRepo := processingrepo.NewProcessingSQLiteRepo(db)
	fixturesDir := t.TempDir()
	storageRoot := filepath.Join(t.TempDir(), "storage")
	provider := jobs.NewPipelineTranscriptProvider(
		nil,
		&fixtureDownloader{dir: fixturesDir},
		&fixtureChunker{dir: fixturesDir},
		&fixtureTranscriber{},
	)
	manager := jobs.NewManagerWithArtifacts(
		jobRepo,
		&allowLocker{},
		paragraphRepo,
		nil,
		provider,
		&mirrorSummary{},
		jobs.NewObjectStoragePublisher(storageinfra.NewLocalUploader(storageRoot)),
		processingRepo,
	)

	jobID, err := manager.StartJob(ctx, podcastID, "http://audio.test/episode.mp3", nil)
	if err != nil {
		t.Fatalf("start job: %v", err)
	}
	job := waitForJobStatus(t, jobRepo, jobID, "succeeded")
	if job.Error != nil {
		t.Fatalf("job unexpectedly failed: %s", *job.Error)
	}

	originalPath := filepath.Join(storageRoot, podcastID, "original.mp3")
	if got, err := os.ReadFile(originalPath); err != nil || string(got) != "original audio fixture" {
		t.Fatalf("stored original mismatch: body=%q err=%v", got, err)
	}

	chunks, err := processingRepo.ListAudioChunks(podcastID)
	if err != nil {
		t.Fatalf("list chunks: %v", err)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %+v", chunks)
	}
	for idx, chunk := range chunks {
		wantOrder := idx + 1
		wantPath := filepath.Join(storageRoot, podcastID, "chunks", fmt.Sprintf("chunk-%03d.mp3", wantOrder))
		if chunk.OrderIndex != wantOrder {
			t.Fatalf("chunk order mismatch at %d: %+v", idx, chunk)
		}
		if chunk.FilePath != wantPath {
			t.Fatalf("chunk path mismatch: got %q want %q", chunk.FilePath, wantPath)
		}
		if chunk.ProcessingJobID == nil || *chunk.ProcessingJobID != jobID {
			t.Fatalf("chunk missing processing job id: %+v", chunk)
		}
		if chunk.ByteSize == nil || *chunk.ByteSize == 0 {
			t.Fatalf("chunk missing byte size: %+v", chunk)
		}
		if chunk.Checksum == nil || len(*chunk.Checksum) != 64 {
			t.Fatalf("chunk missing sha256 checksum: %+v", chunk)
		}
		if _, err := os.Stat(chunk.FilePath); err != nil {
			t.Fatalf("stored chunk does not exist: %v", err)
		}
	}

	segments, err := processingRepo.ListTranscriptSegments(podcastID)
	if err != nil {
		t.Fatalf("list transcript segments: %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("expected 2 transcript segments, got %+v", segments)
	}
	for idx, segment := range segments {
		wantOrder := idx + 1
		if segment.OrderIndex != wantOrder {
			t.Fatalf("segment order mismatch at %d: %+v", idx, segment)
		}
		if segment.AudioChunkID == nil || *segment.AudioChunkID != chunks[idx].ID {
			t.Fatalf("segment missing matching audio chunk id: segment=%+v chunks=%+v", segment, chunks)
		}
		if segment.Provider == nil || *segment.Provider != "fixture" {
			t.Fatalf("segment missing provider metadata: %+v", segment)
		}
		if segment.Model == nil || *segment.Model != "fixture-model" {
			t.Fatalf("segment missing model metadata: %+v", segment)
		}
		if !strings.Contains(segment.Text, fmt.Sprintf("chunk-%c.mp3", 'a'+idx)) {
			t.Fatalf("segment text mismatch: %+v", segment)
		}
	}
}

func TestManagerMarksJobFailedWhenChunkingFails(t *testing.T) {
	ctx := context.Background()
	db, err := dbinfra.NewSQLite(ctx, filepath.Join(t.TempDir(), "podcast.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	podcastID := createEpisode(t, db)
	jobRepo := jobrepo.NewJobSQLiteRepo(db)
	processingRepo := processingrepo.NewProcessingSQLiteRepo(db)
	fixturesDir := t.TempDir()
	provider := jobs.NewPipelineTranscriptProvider(
		nil,
		&fixtureDownloader{dir: fixturesDir},
		&fixtureChunker{err: errors.New("fixture chunker failed")},
		&fixtureTranscriber{},
	)
	manager := jobs.NewManagerWithArtifacts(
		jobRepo,
		&allowLocker{},
		paragraphrepo.NewParagraphSQLiteRepo(db),
		nil,
		provider,
		&mirrorSummary{},
		jobs.NewObjectStoragePublisher(storageinfra.NewLocalUploader(filepath.Join(t.TempDir(), "storage"))),
		processingRepo,
	)

	jobID, err := manager.StartJob(ctx, podcastID, "http://audio.test/episode.mp3", nil)
	if err != nil {
		t.Fatalf("start job: %v", err)
	}
	job := waitForJobStatus(t, jobRepo, jobID, "failed")
	if job.Error == nil || !strings.Contains(*job.Error, "chunk audio") {
		t.Fatalf("expected chunk failure in job error, got %+v", job)
	}

	chunks, err := processingRepo.ListAudioChunks(podcastID)
	if err != nil {
		t.Fatalf("list chunks: %v", err)
	}
	if len(chunks) != 0 {
		t.Fatalf("failed chunking should not persist chunk metadata: %+v", chunks)
	}
}

func TestManagerMarksJobFailedWhenTranscriptionFails(t *testing.T) {
	ctx := context.Background()
	db, err := dbinfra.NewSQLite(ctx, filepath.Join(t.TempDir(), "podcast.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	podcastID := createEpisode(t, db)
	jobRepo := jobrepo.NewJobSQLiteRepo(db)
	processingRepo := processingrepo.NewProcessingSQLiteRepo(db)
	fixturesDir := t.TempDir()
	provider := jobs.NewPipelineTranscriptProvider(
		nil,
		&fixtureDownloader{dir: fixturesDir},
		&fixtureChunker{dir: fixturesDir},
		&fixtureTranscriber{err: errors.New("fixture transcription failed")},
	)
	manager := jobs.NewManagerWithArtifacts(
		jobRepo,
		&allowLocker{},
		paragraphrepo.NewParagraphSQLiteRepo(db),
		nil,
		provider,
		&mirrorSummary{},
		jobs.NewObjectStoragePublisher(storageinfra.NewLocalUploader(filepath.Join(t.TempDir(), "storage"))),
		processingRepo,
	)

	jobID, err := manager.StartJob(ctx, podcastID, "http://audio.test/episode.mp3", nil)
	if err != nil {
		t.Fatalf("start job: %v", err)
	}
	job := waitForJobStatus(t, jobRepo, jobID, "failed")
	if job.Error == nil || !strings.Contains(*job.Error, "transcribe chunk 1") {
		t.Fatalf("expected transcription failure in job error, got %+v", job)
	}

	segments, err := processingRepo.ListTranscriptSegments(podcastID)
	if err != nil {
		t.Fatalf("list transcript segments: %v", err)
	}
	if len(segments) != 0 {
		t.Fatalf("failed transcription should not persist transcript segments: %+v", segments)
	}
}

func createEpisode(t *testing.T, db *sql.DB) string {
	t.Helper()
	audioURL := "http://audio.test/episode.mp3"
	id, err := podcastrepo.NewPodcastSQLiteRepo(db).UpsertSource("http://podcast.test/episode", "Episode", "Host", nil, &audioURL, nil)
	if err != nil {
		t.Fatalf("upsert episode: %v", err)
	}
	return id
}

func waitForJobStatus(t *testing.T, repo jobs.JobRepository, jobID, want string) domain.ProcessingJob {
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

type allowLocker struct{}

func (l *allowLocker) Acquire(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (l *allowLocker) Release(ctx context.Context, key string) error {
	return nil
}

type fixtureDownloader struct {
	dir string
	err error
}

func (d *fixtureDownloader) Download(url string) (string, error) {
	if d.err != nil {
		return "", d.err
	}
	path := filepath.Join(d.dir, "original-input.mp3")
	if err := os.WriteFile(path, []byte("original audio fixture"), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

type fixtureChunker struct {
	dir string
	err error
}

func (c *fixtureChunker) Chunk(inputPath string) ([]string, error) {
	if c.err != nil {
		return nil, c.err
	}
	first := filepath.Join(c.dir, "chunk-a.mp3")
	second := filepath.Join(c.dir, "chunk-b.mp3")
	if err := os.WriteFile(first, []byte("first chunk"), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(second, []byte("second chunk"), 0o644); err != nil {
		return nil, err
	}
	return []string{first, second}, nil
}

type fixtureTranscriber struct {
	err error
}

func (t *fixtureTranscriber) TranscribeFile(path string) (string, error) {
	if t.err != nil {
		return "", t.err
	}
	return "transcribed " + filepath.Base(path), nil
}

func (t *fixtureTranscriber) TranscribeFileDetailed(path string) (domain.TranscriptionResult, error) {
	if t.err != nil {
		return domain.TranscriptionResult{}, t.err
	}
	return domain.TranscriptionResult{
		Text:     "transcribed " + filepath.Base(path),
		Provider: "fixture",
		Model:    "fixture-model",
	}, nil
}

type mirrorSummary struct{}

func (s *mirrorSummary) Summarize(paragraphs []domain.Paragraph) ([]domain.Summary, error) {
	summaries := make([]domain.Summary, len(paragraphs))
	for idx, paragraph := range paragraphs {
		summaries[idx] = domain.Summary{OrderIndex: paragraph.OrderIndex, Text: "summary " + paragraph.Text}
	}
	return summaries, nil
}
