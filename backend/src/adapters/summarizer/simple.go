package summarizer

// SimpleSummarizer returns a short summary prefix for each paragraph.
// TODO: Replace with real summarization service (e.g., SUMMARIZE_URL + SUMMARIZE_KEY).
type SimpleSummarizer struct{}

func NewSimpleSummarizer() *SimpleSummarizer {
	return &SimpleSummarizer{}
}

func (s *SimpleSummarizer) Summarize(paragraphs []string) ([]string, error) {
	result := make([]string, len(paragraphs))
	for i, p := range paragraphs {
		if len(p) > 80 {
			result[i] = "Summary: " + p[:80]
		} else {
			result[i] = "Summary: " + p
		}
	}
	return result, nil
}
