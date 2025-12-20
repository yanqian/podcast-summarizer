package transcription

import (
	"fmt"
)

// Transcriber calls an external transcription service for an audio chunk.
// TODO: Replace stub with real service endpoint/auth (e.g., TRANSCRIBE_URL + TRANSCRIBE_KEY).
type Transcriber struct{}

func NewTranscriber() *Transcriber {
	return &Transcriber{}
}

func (t *Transcriber) Transcribe(url string) ([]byte, error) {
	return []byte(fmt.Sprintf("Transcribed content from %s", url)), nil
}

// TranscribeFile is a placeholder to transcribe a local file; replace with real implementation.
func (t *Transcriber) TranscribeFile(path string) (string, error) {
	return fmt.Sprintf("Transcribed text for %s", path), nil
}
