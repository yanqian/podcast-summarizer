package config

import "os"

type Config struct {
	SQLitePath            string
	APIBaseURL            string
	FFMPEGPath            string
	OpenAIAPIKey          string
	OpenAITranscribeModel string
	OpenAISummarizeModel  string
	LocalStoragePath      string
}

func Load() Config {
	return Config{
		SQLitePath:            getenvDefault("SQLITE_PATH", "data/podcast.db"),
		APIBaseURL:            os.Getenv("API_BASE_URL"),
		FFMPEGPath:            getenvDefault("FFMPEG_PATH", "ffmpeg"),
		OpenAIAPIKey:          getenvDefault("OPENAI_API_KEY", ""),
		OpenAITranscribeModel: getenvDefault("OPENAI_TRANSCRIBE_MODEL", "whisper-1"),
		OpenAISummarizeModel:  getenvDefault("OPENAI_SUMMARIZE_MODEL", "gpt-4o-mini"),
		LocalStoragePath:      getenvDefault("LOCAL_STORAGE_PATH", "data/storage"),
	}
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
