package app

import (
	"fmt"
	"strings"

	"podcast-summarizer/src/core/domain"
)

// FormatTextBundle formats paragraphs with numbering.
func FormatTextBundle(paragraphs []domain.ParagraphWithSummary) (transcript string, summary string) {
	var transcriptLines []string
	var summaryLines []string
	for i, p := range paragraphs {
		n := i + 1
		transcriptLines = append(transcriptLines, fmt.Sprintf("%d. %s", n, p.Text))
		summaryLines = append(summaryLines, fmt.Sprintf("%d. %s", n, p.Summary))
	}
	return strings.Join(transcriptLines, "\n"), strings.Join(summaryLines, "\n")
}
