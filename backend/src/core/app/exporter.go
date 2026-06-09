package app

import (
	"fmt"
	"strings"

	"podcast-summarizer/src/core/domain"
)

// FormatSegmentBundle formats current transcript and summary segments for export.
func FormatSegmentBundle(transcripts []domain.TranscriptSegment, summaries []domain.SummarySourceMapping) (transcript string, summary string) {
	var transcriptLines []string
	for i, p := range transcripts {
		n := i + 1
		transcriptLines = append(transcriptLines, fmt.Sprintf("%d. %s", n, p.Text))
	}

	var summaryLines []string
	for i, item := range summaries {
		n := i + 1
		summaryLines = append(summaryLines, fmt.Sprintf("%d. %s", n, item.Summary.Text))
	}
	return strings.Join(transcriptLines, "\n"), strings.Join(summaryLines, "\n")
}
