package integration

import (
	"context"
	"testing"

	"podcast-summarizer/src/core/domain"
	dbinfra "podcast-summarizer/src/infra/db"
	jobrepo "podcast-summarizer/src/repo/job"
	paragraphrepo "podcast-summarizer/src/repo/paragraph"
	podcastrepo "podcast-summarizer/src/repo/podcast"
	processingrepo "podcast-summarizer/src/repo/processing"
)

func TestSQLiteStoragePersistsPodcastJobAndAlignedParagraphs(t *testing.T) {
	db, err := dbinfra.NewSQLite(context.Background(), t.TempDir()+"/podcast.db")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	podcasts := podcastrepo.NewPodcastSQLiteRepo(db)
	jobs := jobrepo.NewJobSQLiteRepo(db)
	paragraphs := paragraphrepo.NewParagraphSQLiteRepo(db)

	audioURL := "https://example.com/audio.mp3"
	podcastID, err := podcasts.UpsertSource("https://example.com/podcast", "Demo", "Host", nil, &audioURL, nil)
	if err != nil {
		t.Fatalf("upsert podcast: %v", err)
	}
	byURL, err := podcasts.GetSourceByURL("https://example.com/podcast")
	if err != nil {
		t.Fatalf("lookup podcast by url: %v", err)
	}
	if byURL.ID != podcastID || byURL.URL != "https://example.com/podcast" {
		t.Fatalf("unexpected podcast lookup: %+v", byURL)
	}

	jobID, err := jobs.Create(domain.ProcessingJob{PodcastID: podcastID, Type: "ingest", Status: "queued"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	duration := 42
	if err := jobs.UpdateStatus(jobID, "succeeded", &duration, nil); err != nil {
		t.Fatalf("update job: %v", err)
	}
	job, err := jobs.Get(jobID)
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if job.Status != "succeeded" || job.DurationMs == nil || *job.DurationMs != duration {
		t.Fatalf("unexpected job: %+v", job)
	}

	if err := paragraphs.SaveTranscript(podcastID, []domain.Paragraph{{OrderIndex: 1, Text: "Original paragraph"}}); err != nil {
		t.Fatalf("save transcript: %v", err)
	}
	if err := paragraphs.SaveSummaries(podcastID, []domain.Summary{{OrderIndex: 1, Text: "Summary paragraph"}}); err != nil {
		t.Fatalf("save summaries: %v", err)
	}
	aligned, err := paragraphs.GetAligned(podcastID)
	if err != nil {
		t.Fatalf("get aligned: %v", err)
	}
	if len(aligned) != 1 || aligned[0].Text != "Original paragraph" || aligned[0].Summary != "Summary paragraph" {
		t.Fatalf("unexpected aligned paragraphs: %+v", aligned)
	}

	items, err := podcasts.ListSources(10)
	if err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if len(items) != 1 || items[0].LatestJobID == nil || *items[0].LatestJobID != jobID {
		t.Fatalf("unexpected podcast list: %+v", items)
	}
}

func TestSQLiteProcessingModelPersistsArtifactsAndMappings(t *testing.T) {
	db, err := dbinfra.NewSQLite(context.Background(), t.TempDir()+"/podcast.db")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	podcasts := podcastrepo.NewPodcastSQLiteRepo(db)
	jobs := jobrepo.NewJobSQLiteRepo(db)
	processing := processingrepo.NewProcessingSQLiteRepo(db)

	audioURL := "https://example.com/audio.mp3"
	podcastID, err := podcasts.UpsertSource("https://example.com/episode", "Episode", "Local", intPtr(3600), &audioURL, nil)
	if err != nil {
		t.Fatalf("upsert episode: %v", err)
	}
	jobID, err := jobs.Create(domain.ProcessingJob{PodcastID: podcastID, Type: "process", Status: "queued"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	start := 0.0
	mid := 30.5
	end := 61.0
	size := int64(1024)
	chunks, err := processing.SaveAudioChunks(podcastID, []domain.AudioChunk{
		{ProcessingJobID: &jobID, OrderIndex: 1, FilePath: "data/storage/episodes/demo/chunk-001.mp3", StartSeconds: &start, EndSeconds: &mid, DurationSeconds: &mid, ByteSize: &size},
		{ProcessingJobID: &jobID, OrderIndex: 2, FilePath: "data/storage/episodes/demo/chunk-002.mp3", StartSeconds: &mid, EndSeconds: &end, DurationSeconds: &mid},
	})
	if err != nil {
		t.Fatalf("save chunks: %v", err)
	}
	if len(chunks) != 2 || chunks[0].ID == "" || chunks[0].EpisodeID != podcastID {
		t.Fatalf("unexpected saved chunks: %+v", chunks)
	}
	listedChunks, err := processing.ListAudioChunks(podcastID)
	if err != nil {
		t.Fatalf("list chunks: %v", err)
	}
	if len(listedChunks) != 2 || listedChunks[1].OrderIndex != 2 || listedChunks[0].ByteSize == nil || *listedChunks[0].ByteSize != size {
		t.Fatalf("unexpected listed chunks: %+v", listedChunks)
	}

	provider := "fixture"
	transcripts, err := processing.SaveTranscriptSegments(podcastID, []domain.TranscriptSegment{
		{AudioChunkID: &chunks[0].ID, OrderIndex: 1, StartSeconds: &start, EndSeconds: &mid, Text: "First transcript segment.", Provider: &provider},
		{AudioChunkID: &chunks[1].ID, OrderIndex: 2, StartSeconds: &mid, EndSeconds: &end, Text: "Second transcript segment.", Provider: &provider},
	})
	if err != nil {
		t.Fatalf("save transcripts: %v", err)
	}
	summaries, err := processing.SaveSummarySegments(podcastID, []domain.SummarySegment{
		{OrderIndex: 1, Text: "Combined summary.", Provider: &provider},
	})
	if err != nil {
		t.Fatalf("save summaries: %v", err)
	}
	if len(transcripts) != 2 || len(summaries) != 1 || transcripts[0].ID == "" || summaries[0].ID == "" {
		t.Fatalf("unexpected segment ids: transcripts=%+v summaries=%+v", transcripts, summaries)
	}

	if err := processing.SaveTranscriptSummaryMappings(podcastID, []domain.TranscriptSummaryMapping{
		{SummarySegmentID: summaries[0].ID, TranscriptSegmentID: transcripts[0].ID, SourceOrder: 1},
		{SummarySegmentID: summaries[0].ID, TranscriptSegmentID: transcripts[1].ID, SourceOrder: 2},
	}); err != nil {
		t.Fatalf("save mappings: %v", err)
	}
	mappings, err := processing.GetSummarySourceMappings(podcastID)
	if err != nil {
		t.Fatalf("get mappings: %v", err)
	}
	if len(mappings) != 1 || mappings[0].Summary.Text != "Combined summary." || len(mappings[0].TranscriptSegments) != 2 {
		t.Fatalf("unexpected mappings: %+v", mappings)
	}
	if mappings[0].TranscriptSegments[0].Text != "First transcript segment." || mappings[0].TranscriptSegments[1].Text != "Second transcript segment." {
		t.Fatalf("mapping order not preserved: %+v", mappings[0].TranscriptSegments)
	}

	duration := 73
	if err := jobs.UpdateStatus(jobID, "succeeded", &duration, nil); err != nil {
		t.Fatalf("update job: %v", err)
	}
	job, err := jobs.Get(jobID)
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if job.Status != "succeeded" || job.DurationMs == nil || *job.DurationMs != duration {
		t.Fatalf("unexpected completed job: %+v", job)
	}
}

func intPtr(v int) *int {
	return &v
}
