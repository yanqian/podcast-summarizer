package api

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"podcast-summarizer/src/adapters/podcast"
	"podcast-summarizer/src/adapters/summarizer"
	"podcast-summarizer/src/adapters/transcript"
	"podcast-summarizer/src/adapters/transcription"
	"podcast-summarizer/src/api/handlers"
	"podcast-summarizer/src/config"
	"podcast-summarizer/src/core/app"
	"podcast-summarizer/src/core/app/jobs"
	"podcast-summarizer/src/core/domain"
	lockinfra "podcast-summarizer/src/infra/lock"
	mediainfra "podcast-summarizer/src/infra/media"
	jobrepo "podcast-summarizer/src/repo/job"
	paragraphrepo "podcast-summarizer/src/repo/paragraph"
	podcastrepo "podcast-summarizer/src/repo/podcast"
	processingrepo "podcast-summarizer/src/repo/processing"
)

type Middleware func(http.Handler) http.Handler

// Dependencies required to build the router.
type Dependencies struct {
	Config  config.Config
	SQLite  *sql.DB
	Context any
	Storage jobs.ObjectUploader
}

// loggingMiddleware logs method, path, and duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

// recoveryMiddleware prevents panics from crashing the server.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware allows frontend dev server to call the API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter returns a mux with dependencies.
func NewRouter(deps Dependencies) http.Handler {
	itunes := podcast.NewItunesClient()
	var paragraphRepo domain.ParagraphRepository
	var jobRepo jobs.JobRepository
	var lock jobs.Locker

	if deps.SQLite == nil {
		panic("api.NewRouter requires SQLite dependencies")
	}
	podcastRepo := podcastrepo.NewPodcastSQLiteRepo(deps.SQLite)
	paragraphRepo = paragraphrepo.NewParagraphSQLiteRepo(deps.SQLite)
	jobSQLiteRepo := jobrepo.NewJobSQLiteRepo(deps.SQLite)
	jobRepo = jobSQLiteRepo
	lock = lockinfra.NewSQLiteLock(deps.SQLite, 5*time.Minute)
	processingRepo := processingrepo.NewProcessingSQLiteRepo(deps.SQLite)

	transcriptFetcher := transcript.NewFetcher()
	var transcriberClient jobs.TranscriptionFileClient = transcription.NewTranscriber()
	if deps.Config.OpenAIAPIKey != "" {
		transcriberClient = transcription.NewOpenAITranscriber(deps.Config.OpenAIAPIKey, deps.Config.OpenAITranscribeModel)
	} else if deps.Config.TranscribeURL != "" {
		transcriberClient = transcription.NewHTTPClient(deps.Config.TranscribeURL, deps.Config.TranscribeKey)
	}

	var summarizerClient jobs.SummaryClient = summarizer.NewSimpleSummarizer()
	if deps.Config.OpenAIAPIKey != "" {
		summarizerClient = summarizer.NewOpenAISummarizer(deps.Config.OpenAIAPIKey, deps.Config.OpenAISummarizeModel)
	} else if deps.Config.SummarizeURL != "" {
		summarizerClient = summarizer.NewHTTPSummarizer(deps.Config.SummarizeURL, deps.Config.SummarizeKey)
	}

	ingestSvc := app.NewIngestService(itunes, transcriptFetcher, podcastRepo, paragraphRepo)
	notifier := jobs.NewSSENotifier()
	downloader := mediainfra.NewFileDownloader()
	chunker := mediainfra.NewFFmpegChunker(deps.Config.FFMPEGPath)
	transcriptProvider := jobs.NewPipelineTranscriptProvider(transcriptFetcher, downloader, chunker, transcriberClient)
	summarySvc := jobs.NewClientSummaryService(summarizerClient)
	storagePublisher := jobs.NewObjectStoragePublisher(deps.Storage)
	jobManager := jobs.NewManagerWithArtifacts(jobRepo, lock, paragraphRepo, notifier, transcriptProvider, summarySvc, storagePublisher, processingRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/api/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/jobs/") {
			handlers.JobStatusHandler(jobRepo)(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/podcasts/ingest", handlers.IngestHandler(ingestSvc, jobManager))
	mux.HandleFunc("/api/podcasts", handlers.ListPodcastsHandler(podcastRepo))
	mux.HandleFunc("/api/podcasts/", func(w http.ResponseWriter, r *http.Request) {
		// Allow trailing slash for list endpoint.
		if r.URL.Path == "/api/podcasts/" {
			handlers.ListPodcastsHandler(podcastRepo)(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/resummarize") {
			handlers.ResummarizeHandler(paragraphRepo, summarySvc)(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/status") {
			handlers.EpisodeStatusHandler(podcastRepo, jobSQLiteRepo)(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/view") {
			handlers.EpisodeDetailHandler(podcastRepo, jobSQLiteRepo, processingRepo)(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/export") {
			handlers.ExportHandler(paragraphRepo)(w, r)
			return
		}
		if len(strings.Split(strings.Trim(r.URL.Path, "/"), "/")) == 3 {
			handlers.EpisodeDetailHandler(podcastRepo, jobSQLiteRepo, processingRepo)(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/streams/transcript/", handlers.StreamHandler(jobManager))

	return recoveryMiddleware(loggingMiddleware(corsMiddleware(mux)))
}
