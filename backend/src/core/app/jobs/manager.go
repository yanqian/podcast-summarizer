package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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

// TranscriptOutput contains transcript paragraphs and generated audio artifacts.
type TranscriptOutput struct {
	Paragraphs   []domain.Paragraph
	OriginalPath string
	ChunkPaths   []string
	Segments     []domain.TranscriptSegment
}

// StageReporter persists user-visible pipeline progress for the current job.
type StageReporter func(status string)

// TranscriptProvider returns paragraphs and any temp file paths produced.
type TranscriptProvider interface {
	Provide(ctx context.Context, audioURL string, transcriptURL *string, report StageReporter) (TranscriptOutput, error)
}

// SummaryService produces paragraph summaries.
type SummaryService interface {
	Summarize(paragraphs []domain.Paragraph) ([]domain.Summary, error)
}

// TranscriptSummaryService produces grouped summaries from persisted transcript segments.
type TranscriptSummaryService interface {
	SummarizeSegments(ctx context.Context, segments []domain.TranscriptSegment) (domain.SummaryResult, error)
}

// StoragePublisher uploads generated artifacts; optional.
type StoragePublisher interface {
	UploadOriginal(ctx context.Context, podcastID string, path string) (string, error)
	UploadChunk(ctx context.Context, podcastID string, idx int, path string) (string, error)
}

// AudioChunkRepository persists durable local chunk references.
type AudioChunkRepository interface {
	SaveAudioChunks(episodeID string, chunks []domain.AudioChunk) ([]domain.AudioChunk, error)
}

// TranscriptSegmentRepository persists enriched generated transcript segments.
type TranscriptSegmentRepository interface {
	SaveTranscriptSegments(episodeID string, segments []domain.TranscriptSegment) ([]domain.TranscriptSegment, error)
}

// SummarySegmentRepository persists grouped summaries and their source transcript mappings.
type SummarySegmentRepository interface {
	SaveSummarySegments(episodeID string, segments []domain.SummarySegment) ([]domain.SummarySegment, error)
	SaveTranscriptSummaryMappings(episodeID string, mappings []domain.TranscriptSummaryMapping) error
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

// DetailedTranscriptionFileClient returns provider/model and optional segment timing metadata.
type DetailedTranscriptionFileClient interface {
	TranscribeFileDetailed(path string) (domain.TranscriptionResult, error)
}

// SummaryClient produces summaries from text.
type SummaryClient interface {
	Summarize(paragraphs []string) ([]string, error)
}

// TranscriptSegmentSummaryClient can group persisted transcript segments into mapped summary segments.
type TranscriptSegmentSummaryClient interface {
	SummarizeTranscriptSegments(ctx context.Context, segments []domain.TranscriptSegment) (domain.SummaryResult, error)
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
	jobRepo            JobRepository
	lock               Locker
	notifier           Notifier
	transcript         TranscriptProvider
	summary            SummaryService
	transcriptSummary  TranscriptSummaryService
	storage            StoragePublisher
	audioChunks        AudioChunkRepository
	transcriptSegments TranscriptSegmentRepository
	summarySegments    SummarySegmentRepository
}

func NewManager(jobRepo JobRepository, lock Locker, notifier Notifier, transcript TranscriptProvider, summary SummaryService, storage StoragePublisher) *Manager {
	return NewManagerWithArtifacts(jobRepo, lock, notifier, transcript, summary, storage, nil)
}

func NewManagerWithArtifacts(jobRepo JobRepository, lock Locker, notifier Notifier, transcript TranscriptProvider, summary SummaryService, storage StoragePublisher, audioChunks AudioChunkRepository) *Manager {
	var transcriptSegments TranscriptSegmentRepository
	if repo, ok := audioChunks.(TranscriptSegmentRepository); ok {
		transcriptSegments = repo
	}
	var summarySegments SummarySegmentRepository
	if repo, ok := audioChunks.(SummarySegmentRepository); ok {
		summarySegments = repo
	}
	var transcriptSummary TranscriptSummaryService
	if svc, ok := summary.(TranscriptSummaryService); ok {
		transcriptSummary = svc
	}
	return &Manager{
		jobRepo:            jobRepo,
		lock:               lock,
		notifier:           notifier,
		transcript:         transcript,
		summary:            summary,
		transcriptSummary:  transcriptSummary,
		storage:            storage,
		audioChunks:        audioChunks,
		transcriptSegments: transcriptSegments,
		summarySegments:    summarySegments,
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
		segments, err := m.saveTranscriptSegments(podcastID, TranscriptOutput{Paragraphs: []domain.Paragraph{{OrderIndex: 1, Text: "Placeholder transcript chunk"}}}, nil)
		if err == nil {
			_ = m.saveSummarySegments(context.Background(), podcastID, segments)
		}
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
		m.updateJobStatus(jobID, "running", logPrefix)

		output, err := m.transcript.Provide(jobCtx, audioURL, transcriptURL, func(status string) {
			m.updateJobStatus(jobID, status, logPrefix)
		})
		defer cleanupFiles(append([]string{output.OriginalPath}, output.ChunkPaths...)...)
		if err != nil {
			m.failJob(jobID, ch, start, fmt.Sprintf("transcript pipeline failed: %v", err))
			log.Printf("%s transcript pipeline failed: %v", logPrefix, err)
			return
		}
		if len(output.Paragraphs) == 0 {
			m.failJob(jobID, ch, start, "transcript pipeline produced no paragraphs")
			log.Printf("%s transcript pipeline produced no paragraphs", logPrefix)
			return
		}
		storedChunks, err := m.storeAudioArtifacts(jobCtx, podcastID, jobID, output.OriginalPath, output.ChunkPaths)
		if err != nil {
			m.failJob(jobID, ch, start, fmt.Sprintf("audio artifact storage failed: %v", err))
			log.Printf("%s audio artifact storage failed: %v", logPrefix, err)
			return
		}

		for _, p := range output.Paragraphs {
			if m.notifier != nil {
				m.notifier.NotifyChunk(jobID, p)
			}
		}

		savedSegments, err := m.saveTranscriptSegments(podcastID, output, storedChunks)
		if err != nil {
			m.failJob(jobID, ch, start, fmt.Sprintf("save transcript segments failed: %v", err))
			log.Printf("%s save transcript segments failed: %v", logPrefix, err)
			return
		}

		m.updateJobStatus(jobID, "summarizing", logPrefix)
		if m.summary == nil {
			m.failJob(jobID, ch, start, "summarization failed: no summary service configured")
			log.Printf("%s summarization failed: no summary service configured", logPrefix)
			return
		}
		if err := m.saveSummarySegments(jobCtx, podcastID, savedSegments); err != nil {
			m.failJob(jobID, ch, start, fmt.Sprintf("save summary mappings failed: %v", err))
			log.Printf("%s save summary mappings failed: %v", logPrefix, err)
			return
		}

		duration := int(time.Since(start).Milliseconds())
		_ = m.jobRepo.UpdateStatus(jobID, "succeeded", &duration, nil)
		log.Printf("%s finished in %dms (%d paragraphs)", logPrefix, duration, len(output.Paragraphs))
		if m.notifier != nil {
			m.notifier.NotifyDone(jobID)
		}
		if ch != nil {
			close(ch)
		}
	}()

	return jobID, nil
}

func (m *Manager) updateJobStatus(jobID, status, logPrefix string) {
	if strings.TrimSpace(status) == "" {
		return
	}
	if err := m.jobRepo.UpdateStatus(jobID, status, nil, nil); err != nil {
		log.Printf("%s status update %q failed: %v", logPrefix, status, err)
	}
}

func (m *Manager) failJob(jobID string, ch chan []byte, start time.Time, message string) {
	duration := int(time.Since(start).Milliseconds())
	_ = m.jobRepo.UpdateStatus(jobID, "failed", &duration, &message)
	if m.notifier != nil {
		m.notifier.NotifyDone(jobID)
	}
	if ch != nil {
		close(ch)
	}
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

func (m *Manager) storeAudioArtifacts(ctx context.Context, podcastID, jobID, originalPath string, chunkPaths []string) ([]domain.AudioChunk, error) {
	if m.storage == nil && m.audioChunks == nil {
		return nil, nil
	}

	if m.storage != nil && originalPath != "" {
		if _, err := m.storage.UploadOriginal(ctx, podcastID, originalPath); err != nil {
			return nil, fmt.Errorf("store original audio: %w", err)
		}
	}

	var storedChunks []string
	for idx, path := range chunkPaths {
		if path == "" {
			continue
		}
		storedPath := path
		if m.storage != nil {
			uploadedPath, err := m.storage.UploadChunk(ctx, podcastID, idx+1, path)
			if err != nil {
				return nil, fmt.Errorf("store audio chunk %d: %w", idx+1, err)
			}
			storedPath = uploadedPath
		}
		if storedPath != "" {
			storedChunks = append(storedChunks, storedPath)
		}
	}

	if m.audioChunks == nil || len(storedChunks) == 0 {
		return nil, nil
	}
	chunks, err := buildAudioChunkRecords(jobID, storedChunks)
	if err != nil {
		return nil, err
	}
	savedChunks, err := m.audioChunks.SaveAudioChunks(podcastID, chunks)
	if err != nil {
		return nil, fmt.Errorf("persist audio chunks: %w", err)
	}
	return savedChunks, nil
}

func (m *Manager) saveTranscriptSegments(podcastID string, output TranscriptOutput, chunks []domain.AudioChunk) ([]domain.TranscriptSegment, error) {
	if m.transcriptSegments == nil {
		return nil, fmt.Errorf("transcript segment repository is not configured")
	}
	segments := output.Segments
	if len(segments) == 0 {
		segments = transcriptSegmentsFromParagraphs(output.Paragraphs)
	}
	if len(segments) == 0 {
		return nil, nil
	}
	chunksByOrder := map[int]string{}
	for _, chunk := range chunks {
		chunksByOrder[chunk.OrderIndex] = chunk.ID
	}
	for idx := range segments {
		if segments[idx].OrderIndex <= 0 {
			segments[idx].OrderIndex = idx + 1
		}
		if segments[idx].AudioChunkID == nil {
			if chunkID, ok := chunksByOrder[segments[idx].OrderIndex]; ok && chunkID != "" {
				segments[idx].AudioChunkID = &chunkID
			}
		}
	}
	savedSegments, err := m.transcriptSegments.SaveTranscriptSegments(podcastID, segments)
	if err != nil {
		return nil, err
	}
	return savedSegments, nil
}

func (m *Manager) saveSummarySegments(ctx context.Context, podcastID string, transcripts []domain.TranscriptSegment) error {
	if m.transcriptSummary == nil || m.summarySegments == nil || len(transcripts) == 0 {
		return fmt.Errorf("summary segment persistence is not configured")
	}
	result, err := m.transcriptSummary.SummarizeSegments(ctx, transcripts)
	if err != nil {
		return err
	}
	if len(result.Segments) == 0 {
		return fmt.Errorf("summary service returned no segments")
	}

	provider := nonEmptyStringPtr(result.Provider)
	model := nonEmptyStringPtr(result.Model)
	summaries := make([]domain.SummarySegment, 0, len(result.Segments))
	sourceIDsByOrder := make([][]string, 0, len(result.Segments))
	for idx, segment := range result.Segments {
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			return fmt.Errorf("summary segment %d is empty", idx+1)
		}
		sourceIDs := compactStrings(segment.SourceTranscriptSegmentIDs)
		if len(sourceIDs) == 0 {
			return fmt.Errorf("summary segment %d has no source transcript segments", idx+1)
		}
		orderIndex := segment.OrderIndex
		if orderIndex <= 0 {
			orderIndex = idx + 1
		}
		summaries = append(summaries, domain.SummarySegment{
			OrderIndex: orderIndex,
			Text:       text,
			Provider:   provider,
			Model:      model,
		})
		sourceIDsByOrder = append(sourceIDsByOrder, sourceIDs)
	}

	savedSummaries, err := m.summarySegments.SaveSummarySegments(podcastID, summaries)
	if err != nil {
		return err
	}
	if len(savedSummaries) != len(sourceIDsByOrder) {
		return fmt.Errorf("saved %d summary segments for %d mapping groups", len(savedSummaries), len(sourceIDsByOrder))
	}
	mappings := make([]domain.TranscriptSummaryMapping, 0)
	for idx, summary := range savedSummaries {
		for sourceOrder, transcriptID := range sourceIDsByOrder[idx] {
			mappings = append(mappings, domain.TranscriptSummaryMapping{
				SummarySegmentID:    summary.ID,
				TranscriptSegmentID: transcriptID,
				SourceOrder:         sourceOrder + 1,
			})
		}
	}
	return m.summarySegments.SaveTranscriptSummaryMappings(podcastID, mappings)
}

func buildAudioChunkRecords(jobID string, paths []string) ([]domain.AudioChunk, error) {
	result := make([]domain.AudioChunk, 0, len(paths))
	for idx, path := range paths {
		size, checksum, err := fileMetadata(path)
		if err != nil {
			return nil, fmt.Errorf("read chunk metadata %s: %w", path, err)
		}
		result = append(result, domain.AudioChunk{
			ProcessingJobID: &jobID,
			OrderIndex:      idx + 1,
			FilePath:        path,
			ByteSize:        &size,
			Checksum:        &checksum,
		})
	}
	return result, nil
}

func fileMetadata(path string) (int64, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, "", err
	}
	defer func() {
		_ = file.Close()
	}()

	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return 0, "", err
	}
	return size, hex.EncodeToString(hash.Sum(nil)), nil
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

func (p *PipelineTranscriptProvider) Provide(ctx context.Context, audioURL string, transcriptURL *string, report StageReporter) (TranscriptOutput, error) {
	var paragraphs []domain.Paragraph

	if transcriptURL != nil && *transcriptURL != "" && p.fetcher != nil {
		reportStage(report, "fetching_transcript")
		if data, err := p.fetcher.Fetch(*transcriptURL); err == nil {
			paragraphs = splitTranscript(string(data))
		}
	}

	output := TranscriptOutput{Paragraphs: paragraphs}
	var originalPath string
	var chunkPaths []string

	if len(paragraphs) == 0 && audioURL != "" && p.downloader != nil && p.chunker != nil && p.transcriber != nil {
		reportStage(report, "downloading")
		path, err := p.downloader.Download(audioURL)
		if err != nil {
			return TranscriptOutput{}, fmt.Errorf("download audio: %w", err)
		}
		originalPath = path
		reportStage(report, "chunking")
		chunks, err := p.chunker.Chunk(path)
		if err != nil {
			return TranscriptOutput{OriginalPath: originalPath}, fmt.Errorf("chunk audio: %w", err)
		}
		chunkPaths = chunks
		reportStage(report, "transcribing")
		for idx, c := range chunks {
			result, terr := transcribeChunk(p.transcriber, c)
			if terr != nil {
				return TranscriptOutput{OriginalPath: originalPath, ChunkPaths: chunkPaths}, fmt.Errorf("transcribe chunk %d: %w", idx+1, terr)
			}
			text := strings.TrimSpace(result.Text)
			if text == "" {
				return TranscriptOutput{OriginalPath: originalPath, ChunkPaths: chunkPaths}, fmt.Errorf("transcribe chunk %d returned empty text", idx+1)
			}
			paragraphs = append(paragraphs, domain.Paragraph{OrderIndex: idx + 1, Text: text})
			output.Segments = append(output.Segments, transcriptSegmentFromResult(idx+1, result))
		}
	}

	output.Paragraphs = paragraphs
	output.OriginalPath = originalPath
	output.ChunkPaths = chunkPaths
	return output, nil
}

func reportStage(report StageReporter, status string) {
	if report != nil {
		report(status)
	}
}

func transcribeChunk(client TranscriptionFileClient, path string) (domain.TranscriptionResult, error) {
	if detailed, ok := client.(DetailedTranscriptionFileClient); ok {
		return detailed.TranscribeFileDetailed(path)
	}
	text, err := client.TranscribeFile(path)
	if err != nil {
		return domain.TranscriptionResult{}, err
	}
	return domain.TranscriptionResult{Text: text}, nil
}

func transcriptSegmentFromResult(orderIndex int, result domain.TranscriptionResult) domain.TranscriptSegment {
	text := strings.TrimSpace(result.Text)
	segment := domain.TranscriptSegment{
		OrderIndex: orderIndex,
		Text:       text,
		Provider:   nonEmptyStringPtr(result.Provider),
		Model:      nonEmptyStringPtr(result.Model),
	}
	if len(result.Segments) > 0 {
		first := result.Segments[0]
		last := result.Segments[len(result.Segments)-1]
		segment.StartSeconds = first.StartSeconds
		segment.EndSeconds = last.EndSeconds
		if text == "" {
			parts := make([]string, 0, len(result.Segments))
			for _, part := range result.Segments {
				partText := strings.TrimSpace(part.Text)
				if partText != "" {
					parts = append(parts, partText)
				}
			}
			segment.Text = strings.Join(parts, " ")
		}
	}
	return segment
}

func transcriptSegmentsFromParagraphs(paragraphs []domain.Paragraph) []domain.TranscriptSegment {
	segments := make([]domain.TranscriptSegment, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		segments = append(segments, domain.TranscriptSegment{
			OrderIndex: paragraph.OrderIndex,
			Text:       paragraph.Text,
		})
	}
	return segments
}

func nonEmptyStringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
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

func (s *ClientSummaryService) SummarizeSegments(ctx context.Context, segments []domain.TranscriptSegment) (domain.SummaryResult, error) {
	if s.client == nil {
		return domain.SummaryResult{}, fmt.Errorf("no summary client configured")
	}
	if len(segments) == 0 {
		return domain.SummaryResult{}, fmt.Errorf("no transcript segments to summarize")
	}
	if client, ok := s.client.(TranscriptSegmentSummaryClient); ok {
		return client.SummarizeTranscriptSegments(ctx, segments)
	}

	groups := groupTranscriptSegments(segments, 2)
	texts := make([]string, 0, len(groups))
	for _, group := range groups {
		parts := make([]string, 0, len(group))
		for _, segment := range group {
			if text := strings.TrimSpace(segment.Text); text != "" {
				parts = append(parts, text)
			}
		}
		texts = append(texts, strings.Join(parts, "\n\n"))
	}

	resp, err := s.client.Summarize(texts)
	if err != nil {
		return domain.SummaryResult{}, err
	}
	if len(resp) == 0 {
		return domain.SummaryResult{}, fmt.Errorf("summary client returned 0 summaries")
	}

	result := domain.SummaryResult{Segments: make([]domain.SummaryResultSegment, 0, len(groups))}
	for idx, group := range groups {
		text := ""
		if idx < len(resp) {
			text = strings.TrimSpace(resp[idx])
		}
		if text == "" {
			text = fallbackGroupSummary(group)
		}
		sourceIDs := make([]string, 0, len(group))
		for _, transcript := range group {
			if transcript.ID != "" {
				sourceIDs = append(sourceIDs, transcript.ID)
			}
		}
		if len(sourceIDs) == 0 {
			return domain.SummaryResult{}, fmt.Errorf("summary group %d has no source transcript ids", idx+1)
		}
		result.Segments = append(result.Segments, domain.SummaryResultSegment{
			OrderIndex:                 idx + 1,
			Text:                       text,
			SourceTranscriptSegmentIDs: sourceIDs,
		})
	}
	return result, nil
}

func groupTranscriptSegments(segments []domain.TranscriptSegment, groupSize int) [][]domain.TranscriptSegment {
	if groupSize <= 0 {
		groupSize = 1
	}
	groups := make([][]domain.TranscriptSegment, 0, (len(segments)+groupSize-1)/groupSize)
	for start := 0; start < len(segments); start += groupSize {
		end := start + groupSize
		if end > len(segments) {
			end = len(segments)
		}
		groups = append(groups, segments[start:end])
	}
	return groups
}

func fallbackGroupSummary(group []domain.TranscriptSegment) string {
	parts := make([]string, 0, len(group))
	for _, segment := range group {
		if text := strings.TrimSpace(segment.Text); text != "" {
			parts = append(parts, text)
		}
	}
	text := strings.Join(parts, " ")
	if len(text) > 600 {
		text = text[:600] + "..."
	}
	return text
}

// StoragePublisher implementation backed by ObjectUploader.
type ObjectStoragePublisher struct {
	uploader ObjectUploader
}

func NewObjectStoragePublisher(uploader ObjectUploader) *ObjectStoragePublisher {
	return &ObjectStoragePublisher{uploader: uploader}
}

func (p *ObjectStoragePublisher) UploadOriginal(ctx context.Context, podcastID string, path string) (string, error) {
	if p.uploader == nil || path == "" {
		return "", nil
	}
	return p.uploader.Upload(ctx, storage.BuildObjectKey(podcastID, "original.mp3"), path)
}

func (p *ObjectStoragePublisher) UploadChunk(ctx context.Context, podcastID string, idx int, path string) (string, error) {
	if p.uploader == nil || path == "" {
		return "", nil
	}
	key := storage.BuildObjectKey(podcastID, fmt.Sprintf("chunks/chunk-%03d.mp3", idx))
	return p.uploader.Upload(ctx, key, path)
}
