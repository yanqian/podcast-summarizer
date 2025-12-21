package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"podcast-summarizer/src/core/domain"
	"podcast-summarizer/src/infra/storage"
)

// JobRepository persists job state.
type JobRepository interface {
	Create(job domain.ProcessingJob) (string, error)
	UpdateStatus(id, status string, durationMs *int, errorMessage *string) error
	Get(id string) (domain.ProcessingJob, error)
}

// Locker prevents duplicate jobs per key.
type Locker interface {
	Acquire(ctx context.Context, key string) (bool, error)
	Release(ctx context.Context, key string) error
}

// ParagraphSaver persists transcript/summaries.
type ParagraphSaver interface {
	SaveTranscript(podcastID string, paragraphs []domain.Paragraph) error
	SaveSummaries(podcastID string, summaries []domain.Summary) error
}

// TranscriptProvider returns paragraphs and any temp file paths produced.
type TranscriptProvider interface {
	Provide(ctx context.Context, audioURL string, transcriptURL *string) (paragraphs []domain.Paragraph, originalPath string, chunkPaths []string, err error)
}

// SummaryService produces paragraph summaries.
type SummaryService interface {
	Summarize(paragraphs []domain.Paragraph) ([]domain.Summary, error)
}

// StoragePublisher uploads generated artifacts; optional.
type StoragePublisher interface {
	UploadOriginal(ctx context.Context, podcastID string, path string) error
	UploadChunk(ctx context.Context, podcastID string, idx int, path string) error
}

// Notifier handles streaming events (SSE).
type Notifier interface {
	Subscribe(jobID string) chan []byte
	NotifyChunk(jobID string, p domain.Paragraph)
	NotifyDone(jobID string)
}

// TranscriptionFileClient turns audio file into text.
type TranscriptionFileClient interface {
	TranscribeFile(path string) (string, error)
}

// SummaryClient produces summaries from text.
type SummaryClient interface {
	Summarize(paragraphs []string) ([]string, error)
}

// FileDownloader fetches remote media to disk.
type FileDownloader interface {
	Download(url string) (string, error)
}

// AudioChunker splits audio into chunks.
type AudioChunker interface {
	Chunk(inputPath string) ([]string, error)
}

// ObjectUploader uploads files to object storage.
type ObjectUploader interface {
	Upload(ctx context.Context, key string, localPath string) (string, error)
}

type Manager struct {
	jobRepo    JobRepository
	lock       Locker
	paras      ParagraphSaver
	notifier   Notifier
	transcript TranscriptProvider
	summary    SummaryService
	storage    StoragePublisher
}

func NewManager(jobRepo JobRepository, lock Locker, paras ParagraphSaver, notifier Notifier, transcript TranscriptProvider, summary SummaryService, storage StoragePublisher) *Manager {
	return &Manager{
		jobRepo:    jobRepo,
		lock:       lock,
		paras:      paras,
		notifier:   notifier,
		transcript: transcript,
		summary:    summary,
		storage:    storage,
	}
}

func (m *Manager) Subscribe(jobID string) chan []byte {
	if m.notifier == nil {
		return nil
	}
	return m.notifier.Subscribe(jobID)
}

func (m *Manager) StartStubJob(ctx context.Context, podcastID string) (string, error) {
	job := domain.ProcessingJob{
		PodcastID: podcastID,
		Type:      "ingest",
		Status:    "queued",
	}
	jobID, err := m.jobRepo.Create(job)
	if err != nil {
		return "", err
	}

	ch := m.Subscribe(jobID)
	go func() {
		start := time.Now()
		// Emit chunks
		for i := 1; i <= 3; i++ {
			if m.notifier != nil {
				m.notifier.NotifyChunk(jobID, domain.Paragraph{OrderIndex: i, Text: "Placeholder transcript chunk"})
			}
			time.Sleep(300 * time.Millisecond)
		}
		// Save placeholder paragraphs/summaries
		_ = m.paras.SaveTranscript(podcastID, []domain.Paragraph{
			{OrderIndex: 1, Text: "Placeholder transcript chunk"},
		})
		_ = m.paras.SaveSummaries(podcastID, []domain.Summary{
			{OrderIndex: 1, Text: "Placeholder summary"},
		})
		// Mark done
		duration := int(time.Since(start).Milliseconds())
		_ = m.jobRepo.UpdateStatus(jobID, "succeeded", &duration, nil)
		if m.notifier != nil {
			m.notifier.NotifyDone(jobID)
		}
		if ch != nil {
			close(ch)
		}
	}()

	return jobID, nil
}

// StartJob runs a basic pipeline: fetch or transcribe then summarize and stream.
func (m *Manager) StartJob(ctx context.Context, podcastID string, audioURL string, transcriptURL *string) (string, error) {
	job := domain.ProcessingJob{
		PodcastID: podcastID,
		Type:      "ingest",
		Status:    "queued",
	}
	jobID, err := m.jobRepo.Create(job)
	if err != nil {
		return "", err
	}

	ch := m.Subscribe(jobID)
	go func() {
		logPrefix := "[job:" + jobID + "]"
		log.Printf("%s started podcast=%s audio=%s transcriptURL=%v", logPrefix, podcastID, audioURL, transcriptURL)
		start := time.Now()
		jobCtx := context.Background()
		lockKey := "job:" + podcastID
		acquired, lErr := m.lock.Acquire(jobCtx, lockKey)
		switch {
		case lErr != nil:
			errMsg := "lock acquisition failed"
			_ = m.jobRepo.UpdateStatus(jobID, "failed", nil, &errMsg)
			log.Printf("%s lock acquire key=%s err=%v", logPrefix, lockKey, lErr)
			if m.notifier != nil {
				m.notifier.NotifyDone(jobID)
			}
			if ch != nil {
				close(ch)
			}
			return
		case !acquired:
			errMsg := "another job is already processing this podcast"
			_ = m.jobRepo.UpdateStatus(jobID, "failed", nil, &errMsg)
			log.Printf("%s lock not acquired key=%s", logPrefix, lockKey)
			if m.notifier != nil {
				m.notifier.NotifyDone(jobID)
			}
			if ch != nil {
				close(ch)
			}
			return
		default:
			log.Printf("%s lock acquired key=%s", logPrefix, lockKey)
		}
		defer func() {
			if err := m.lock.Release(jobCtx, lockKey); err != nil {
				log.Printf("%s lock release failed key=%s err=%v", logPrefix, lockKey, err)
			}
		}()

		paragraphs, originalPath, chunkPaths, err := m.transcript.Provide(jobCtx, audioURL, transcriptURL)
		if err != nil || len(paragraphs) == 0 {
			paragraphs = []domain.Paragraph{{OrderIndex: 1, Text: "Transcribed content from audio " + audioURL}}
		}

		for _, p := range paragraphs {
			if m.notifier != nil {
				m.notifier.NotifyChunk(jobID, p)
			}
		}

		_ = m.paras.SaveTranscript(podcastID, paragraphs)

		summaryModels, err := m.summary.Summarize(paragraphs)
		if err != nil || len(summaryModels) == 0 {
			for _, p := range paragraphs {
				summaryModels = append(summaryModels, domain.Summary{OrderIndex: p.OrderIndex, Text: "Summary: " + p.Text})
			}
		}
		_ = m.paras.SaveSummaries(podcastID, summaryModels)

		if m.storage != nil {
			if originalPath != "" {
				_ = m.storage.UploadOriginal(jobCtx, podcastID, originalPath)
			}
			for idx, p := range chunkPaths {
				_ = m.storage.UploadChunk(jobCtx, podcastID, idx+1, p)
			}
		}
		cleanupFiles(append([]string{originalPath}, chunkPaths...)...)

		duration := int(time.Since(start).Milliseconds())
		_ = m.jobRepo.UpdateStatus(jobID, "succeeded", &duration, nil)
		log.Printf("%s finished in %dms (%d paragraphs)", logPrefix, duration, len(paragraphs))
		if m.notifier != nil {
			m.notifier.NotifyDone(jobID)
		}
		if ch != nil {
			close(ch)
		}
	}()

	return jobID, nil
}

func splitTranscript(text string) []domain.Paragraph {
	lines := strings.Split(text, "\n")
	var paragraphs []domain.Paragraph
	current := ""
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			if current != "" {
				paragraphs = append(paragraphs, domain.Paragraph{OrderIndex: len(paragraphs) + 1, Text: current})
				current = ""
			}
			continue
		}
		if current != "" {
			current += " "
		}
		current += l
	}
	if current != "" {
		paragraphs = append(paragraphs, domain.Paragraph{OrderIndex: len(paragraphs) + 1, Text: current})
	}
	return paragraphs
}

func cleanupFiles(paths ...string) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		_ = os.Remove(p)
	}
}

// SSENotifier manages per-job channels and payload formatting.
type SSENotifier struct {
	mu      sync.Mutex
	streams map[string]chan []byte
}

func NewSSENotifier() *SSENotifier {
	return &SSENotifier{streams: make(map[string]chan []byte)}
}

func (n *SSENotifier) Subscribe(jobID string) chan []byte {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := n.streams[jobID]
	if !ok {
		ch = make(chan []byte, 10)
		n.streams[jobID] = ch
	}
	return ch
}

func (n *SSENotifier) NotifyChunk(jobID string, p domain.Paragraph) {
	n.mu.Lock()
	ch, ok := n.streams[jobID]
	n.mu.Unlock()
	if !ok {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"event": "chunk",
		"data": map[string]any{
			"order": p.OrderIndex,
			"text":  p.Text,
		},
	})
	ch <- payload
}

func (n *SSENotifier) NotifyDone(jobID string) {
	n.mu.Lock()
	ch, ok := n.streams[jobID]
	n.mu.Unlock()
	if !ok {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"event": "done",
		"data":  map[string]any{},
	})
	ch <- payload
}

// TranscriptProvider implementation that handles fetch, download, chunk, transcribe.
type PipelineTranscriptProvider struct {
	fetcher     domain.TranscriptFetcher
	downloader  FileDownloader
	chunker     AudioChunker
	transcriber TranscriptionFileClient
}

func NewPipelineTranscriptProvider(fetcher domain.TranscriptFetcher, downloader FileDownloader, chunker AudioChunker, transcriber TranscriptionFileClient) *PipelineTranscriptProvider {
	return &PipelineTranscriptProvider{fetcher: fetcher, downloader: downloader, chunker: chunker, transcriber: transcriber}
}

func (p *PipelineTranscriptProvider) Provide(ctx context.Context, audioURL string, transcriptURL *string) ([]domain.Paragraph, string, []string, error) {
	var paragraphs []domain.Paragraph

	if transcriptURL != nil && *transcriptURL != "" && p.fetcher != nil {
		if data, err := p.fetcher.Fetch(*transcriptURL); err == nil {
			paragraphs = splitTranscript(string(data))
		}
	}

	var originalPath string
	var chunkPaths []string

	if len(paragraphs) == 0 && audioURL != "" && p.downloader != nil && p.chunker != nil && p.transcriber != nil {
		if path, err := p.downloader.Download(audioURL); err == nil {
			originalPath = path
			if chunks, err := p.chunker.Chunk(path); err == nil {
				chunkPaths = chunks
				for idx, c := range chunks {
					text, terr := p.transcriber.TranscribeFile(c)
					if terr != nil {
						text = "Transcription failed: " + terr.Error()
					}
					paragraphs = append(paragraphs, domain.Paragraph{OrderIndex: idx + 1, Text: text})
				}
			}
		}
	}

	return paragraphs, originalPath, chunkPaths, nil
}

// SummaryService implementation backed by SummaryClient.
type ClientSummaryService struct {
	client SummaryClient
}

func NewClientSummaryService(client SummaryClient) *ClientSummaryService {
	return &ClientSummaryService{client: client}
}

func (s *ClientSummaryService) Summarize(paragraphs []domain.Paragraph) ([]domain.Summary, error) {
	if s.client == nil {
		return nil, fmt.Errorf("no summary client configured")
	}
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("no paragraphs to summarize")
	}

	texts := make([]string, len(paragraphs))
	for i, p := range paragraphs {
		texts[i] = p.Text
	}

	resp, err := s.client.Summarize(texts)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, fmt.Errorf("summary client returned 0 summaries")
	}

	summaries := make([]domain.Summary, len(paragraphs))
	for i, p := range paragraphs {
		// Prefer the model output; fall back to the original transcript text when empty or missing.
		summary := ""
		if i < len(resp) {
			summary = strings.TrimSpace(resp[i])
		}
		if summary == "" {
			summary = strings.TrimSpace(p.Text)
		}
		if len(summary) > 600 {
			summary = summary[:600] + "..."
		}
		summaries[i] = domain.Summary{OrderIndex: p.OrderIndex, Text: summary}
	}

	return summaries, nil
}

// StoragePublisher implementation backed by ObjectUploader.
type ObjectStoragePublisher struct {
	uploader ObjectUploader
}

func NewObjectStoragePublisher(uploader ObjectUploader) *ObjectStoragePublisher {
	return &ObjectStoragePublisher{uploader: uploader}
}

func (p *ObjectStoragePublisher) UploadOriginal(ctx context.Context, podcastID string, path string) error {
	if p.uploader == nil || path == "" {
		return nil
	}
	_, err := p.uploader.Upload(ctx, storage.BuildObjectKey(podcastID, "original.mp3"), path)
	return err
}

func (p *ObjectStoragePublisher) UploadChunk(ctx context.Context, podcastID string, idx int, path string) error {
	if p.uploader == nil || path == "" {
		return nil
	}
	key := storage.BuildObjectKey(podcastID, fmt.Sprintf("chunks/chunk-%03d.mp3", idx))
	_, err := p.uploader.Upload(ctx, key, path)
	return err
}
