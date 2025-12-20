package config

import "os"

type Config struct {
	PostgresURL           string
	ValkeyURL             string
	APIBaseURL            string
	FFMPEGPath            string
	TranscribeURL         string
	TranscribeKey         string
	SummarizeURL          string
	SummarizeKey          string
	OpenAIAPIKey          string
	OpenAITranscribeModel string
	OpenAISummarizeModel  string
	R2Endpoint            string
	R2Bucket              string
	R2AccessKey           string
	R2SecretKey           string
	R2PublicBaseURL       string
}

func Load() Config {
	return Config{
		PostgresURL:           os.Getenv("POSTGRES_URL"),
		ValkeyURL:             os.Getenv("VALKEY_URL"),
		APIBaseURL:            os.Getenv("API_BASE_URL"),
		FFMPEGPath:            getenvDefault("FFMPEG_PATH", "ffmpeg"),
		TranscribeURL:         getenvDefault("TRANSCRIBE_URL", ""),
		TranscribeKey:         getenvDefault("TRANSCRIBE_KEY", ""),
		SummarizeURL:          getenvDefault("SUMMARIZE_URL", ""),
		SummarizeKey:          getenvDefault("SUMMARIZE_KEY", ""),
		OpenAIAPIKey:          getenvDefault("OPENAI_API_KEY", ""),
		OpenAITranscribeModel: getenvDefault("OPENAI_TRANSCRIBE_MODEL", "whisper-1"),
		OpenAISummarizeModel:  getenvDefault("OPENAI_SUMMARIZE_MODEL", "gpt-4o-mini"),
		R2Endpoint:            getenvDefault("R2_ENDPOINT", ""),
		R2Bucket:              getenvDefault("R2_BUCKET", ""),
		R2AccessKey:           getenvDefault("R2_ACCESS_KEY", ""),
		R2SecretKey:           getenvDefault("R2_SECRET_KEY", ""),
		R2PublicBaseURL:       getenvDefault("R2_PUBLIC_BASE_URL", ""),
	}
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
