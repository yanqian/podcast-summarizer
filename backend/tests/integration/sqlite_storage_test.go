package integration

import (
	"context"
	"testing"

	"podcast-summarizer/src/core/domain"
	dbinfra "podcast-summarizer/src/infra/db"
	jobrepo "podcast-summarizer/src/repo/job"
	paragraphrepo "podcast-summarizer/src/repo/paragraph"
	podcastrepo "podcast-summarizer/src/repo/podcast"
)

func TestSQLiteStoragePersistsPodcastJobAndAlignedParagraphs(t *testing.T) {
	db, err := dbinfra.NewSQLite(context.Background(), t.TempDir()+"/podcast.db")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	podcasts := podcastrepo.NewPodcastSQLiteRepo(db)
	jobs := jobrepo.NewJobSQLiteRepo(db)
	paragraphs := paragraphrepo.NewParagraphSQLiteRepo(db)

	audioURL := "https://example.com/audio.mp3"
	podcastID, err := podcasts.UpsertSource("https://example.com/podcast", "Demo", "Host", nil, &audioURL, nil)
	if err != nil {
		t.Fatalf("upsert podcast: %v", err)
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
