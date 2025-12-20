package app

import "time"

// RecordTranscription logs timing for transcription pipeline.
func RecordTranscription(duration time.Duration, err error) {
	RecordExternalCall("transcription", time.Now().Add(-duration), err)
}
