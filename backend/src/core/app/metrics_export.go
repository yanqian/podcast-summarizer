package app

import "time"

// RecordExport logs timing for export operations.
func RecordExport(duration time.Duration, err error) {
	RecordExternalCall("export", time.Now().Add(-duration), err)
}
