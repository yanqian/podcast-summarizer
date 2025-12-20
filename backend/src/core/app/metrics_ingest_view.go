package app

import "time"

// Metric hooks placeholders.

func RecordIngest(duration time.Duration, err error) {
	RecordExternalCall("ingest", time.Now().Add(-duration), err)
}

func RecordView(duration time.Duration, err error) {
	RecordExternalCall("view", time.Now().Add(-duration), err)
}
