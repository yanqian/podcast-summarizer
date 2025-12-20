package app

import (
	"log"
	"time"
)

// Timed logs duration for a named operation.
func Timed(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	log.Printf("%s completed in %s (err=%v)", name, time.Since(start), err)
	return err
}

// RecordExternalCall logs external call latency; placeholder for metrics integration.
func RecordExternalCall(name string, start time.Time, err error) {
	log.Printf("external_call=%s duration=%s error=%v", name, time.Since(start), err)
}
