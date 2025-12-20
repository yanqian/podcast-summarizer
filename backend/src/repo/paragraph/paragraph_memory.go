package paragraphrepo

import "podcast-summarizer/src/core/domain"

type ParagraphMemoryRepo struct {
	store map[string][]domain.ParagraphWithSummary
}

func NewParagraphMemoryRepo() *ParagraphMemoryRepo {
	return &ParagraphMemoryRepo{store: make(map[string][]domain.ParagraphWithSummary)}
}

func (r *ParagraphMemoryRepo) SaveTranscript(podcastID string, paragraphs []domain.Paragraph) error {
	// store without summaries for now
	existing := r.store[podcastID]
	var combined []domain.ParagraphWithSummary
	for _, p := range paragraphs {
		combined = append(combined, domain.ParagraphWithSummary{OrderIndex: p.OrderIndex, Text: p.Text})
	}
	r.store[podcastID] = append(existing, combined...)
	return nil
}

func (r *ParagraphMemoryRepo) SaveSummaries(podcastID string, summaries []domain.Summary) error {
	items := r.store[podcastID]
	for i, item := range items {
		for _, s := range summaries {
			if s.OrderIndex == item.OrderIndex {
				items[i].Summary = s.Text
			}
		}
	}
	r.store[podcastID] = items
	return nil
}

func (r *ParagraphMemoryRepo) GetAligned(podcastID string) ([]domain.ParagraphWithSummary, error) {
	return r.store[podcastID], nil
}
