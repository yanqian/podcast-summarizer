package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"podcast-summarizer/src/core/domain"
)

const openAIChatCompletionsEndpoint = "https://api.openai.com/v1/chat/completions"

// OpenAISummarizer calls OpenAI chat completions to summarize paragraphs.
type OpenAISummarizer struct {
	APIKey   string
	Model    string
	Endpoint string
	Client   *http.Client
}

type structuredSummary struct {
	Discussion string `json:"discussion"`
}

type structuredSummarySegment struct {
	Summary   string   `json:"summary"`
	SourceIDs []string `json:"source_ids"`
}

func NewOpenAISummarizer(apiKey, model string) *OpenAISummarizer {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAISummarizer{
		APIKey:   apiKey,
		Model:    model,
		Endpoint: openAIChatCompletionsEndpoint,
		Client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *OpenAISummarizer) Summarize(paragraphs []string) ([]string, error) {
	const batchSize = 10
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("no paragraphs to summarize")
	}

	// Batch to reduce the chance of output truncation and to keep a stable 1:1 mapping.
	out := make([]string, 0, len(paragraphs))
	for start := 0; start < len(paragraphs); start += batchSize {
		end := start + batchSize
		if end > len(paragraphs) {
			end = len(paragraphs)
		}
		batch := paragraphs[start:end]
		summaries, err := s.summarizeBatch(batch)
		if err != nil {
			return nil, fmt.Errorf("summarize batch %d-%d: %w", start, end, err)
		}
		out = append(out, summaries...)
	}
	return out, nil
}

func (s *OpenAISummarizer) SummarizeTranscriptSegments(ctx context.Context, segments []domain.TranscriptSegment) (domain.SummaryResult, error) {
	if strings.TrimSpace(s.APIKey) == "" {
		return domain.SummaryResult{}, fmt.Errorf("openai summarization requires OPENAI_API_KEY")
	}
	if strings.TrimSpace(s.Model) == "" {
		return domain.SummaryResult{}, fmt.Errorf("openai summarization requires a model")
	}
	if len(segments) == 0 {
		return domain.SummaryResult{}, fmt.Errorf("no transcript segments to summarize")
	}
	for _, segment := range segments {
		if strings.TrimSpace(segment.ID) == "" {
			return domain.SummaryResult{}, fmt.Errorf("transcript segment %d has no id", segment.OrderIndex)
		}
	}

	var transcriptBuilder strings.Builder
	for _, segment := range segments {
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			continue
		}
		transcriptBuilder.WriteString(fmt.Sprintf("- id: %s\n  order: %d\n  text: %s\n", segment.ID, segment.OrderIndex, text))
	}
	if transcriptBuilder.Len() == 0 {
		return domain.SummaryResult{}, fmt.Errorf("no transcript text to summarize")
	}

	payload := map[string]any{
		"model": s.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Return strict JSON only. Output must be a JSON array of objects with summary and source_ids keys.",
			},
			{
				"role": "user",
				"content": "Summarize this podcast transcript into ordered summary segments.\n" +
					"Each summary segment may cover one or more adjacent transcript segments.\n" +
					"Every source_ids entry must be an id from the input, and each input id should appear exactly once.\n" +
					"Return ONLY JSON shaped like [{\"summary\":\"...\",\"source_ids\":[\"...\"]}].\n\nTranscript segments:\n" +
					transcriptBuilder.String(),
			},
		},
		"temperature": 0.2,
	}
	body, _ := json.Marshal(payload)
	endpoint := s.Endpoint
	if endpoint == "" {
		endpoint = openAIChatCompletionsEndpoint
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return domain.SummaryResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return domain.SummaryResult{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return domain.SummaryResult{}, fmt.Errorf("openai summarize failed: %s (%s)", resp.Status, string(b))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return domain.SummaryResult{}, err
	}
	if len(out.Choices) == 0 {
		return domain.SummaryResult{}, fmt.Errorf("no choices returned")
	}
	summarySegments, err := parseStructuredSummarySegments(normalizeContent(out.Choices[0].Message.Content), segments)
	if err != nil {
		return domain.SummaryResult{}, err
	}
	return domain.SummaryResult{
		Segments: summarySegments,
		Provider: "openai",
		Model:    s.Model,
	}, nil
}

func (s *OpenAISummarizer) summarizeBatch(paragraphs []string) ([]string, error) {
	if strings.TrimSpace(s.APIKey) == "" {
		return nil, fmt.Errorf("openai summarization requires OPENAI_API_KEY")
	}
	var promptBuilder strings.Builder
	promptBuilder.WriteString("You are summarizing human dialogue.\n\nReturn ONLY a valid JSON array (no code fences, no extra text) with EXACTLY the same number of items as paragraphs, in the same order.\n\nEach item MUST be an object with EXACTLY this key:\n- \"discussion\": a string, less than 10 sentences(based on the paragraph length), information-dense, preserving key requests, answers, decisions, follow-ups, action items, and any numbers/quotes/facts. Do not speculate.\n\nOutput schema example:\n[{\"discussion\":\"...\"}]\n\nParagraphs:\n")
	for i, p := range paragraphs {
		promptBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, p))
	}
	prompt := promptBuilder.String()
	payload := map[string]any{
		"model": s.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "Return strict JSON only. Output must be a JSON array of objects with a single key: discussion."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.4,
	}
	b, _ := json.Marshal(payload)
	endpoint := s.Endpoint
	if endpoint == "" {
		endpoint = openAIChatCompletionsEndpoint
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai summarize failed: %s (%s)", resp.Status, string(body))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}
	content := strings.TrimSpace(out.Choices[0].Message.Content)
	content = normalizeContent(content)

	summaries := parseSummaries(content)
	if len(summaries) == 0 {
		return nil, fmt.Errorf("unable to parse summaries from OpenAI response")
	}

	summaries = alignSummariesToParagraphs(summaries, paragraphs)
	return summaries, nil
}

// normalizeContent strips code fences and extracts the JSON array portion if present.
func normalizeContent(content string) string {
	clean := strings.TrimSpace(content)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```JSON")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSpace(strings.TrimSuffix(clean, "```"))
	if i := strings.Index(clean, "["); i != -1 {
		if j := strings.LastIndex(clean, "]"); j > i {
			return clean[i : j+1]
		}
	}
	return clean
}

// parseSummaries attempts to extract summary strings from various response shapes.
func parseSummaries(content string) []string {
	var summaries []string

	// 0) Preferred: array of structured objects with discussion text.
	var structuredArr []structuredSummary
	if err := json.Unmarshal([]byte(content), &structuredArr); err == nil && len(structuredArr) > 0 {
		for _, v := range structuredArr {
			if text := strings.TrimSpace(v.Discussion); text != "" {
				summaries = append(summaries, text)
			}
		}
		if len(summaries) > 0 {
			return summaries
		}
	}

	// 1) Straight array of strings.
	var arr []string
	if err := json.Unmarshal([]byte(content), &arr); err == nil && len(arr) > 0 {
		for _, v := range arr {
			summaries = append(summaries, strings.TrimSpace(v))
		}
		return summaries
	}

	// 2) Fallback: split lines or bullet list.
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(strings.TrimPrefix(l, "-"))
		l = strings.Trim(l, "[]{}\" ")
		if l != "" {
			summaries = append(summaries, l)
		}
	}

	return summaries
}

func parseStructuredSummarySegments(content string, transcripts []domain.TranscriptSegment) ([]domain.SummaryResultSegment, error) {
	var raw []structuredSummarySegment
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("parse summary segments: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("openai returned no summary segments")
	}

	validIDs := make(map[string]struct{}, len(transcripts))
	for _, transcript := range transcripts {
		if transcript.ID != "" {
			validIDs[transcript.ID] = struct{}{}
		}
	}

	result := make([]domain.SummaryResultSegment, 0, len(raw))
	seen := map[string]struct{}{}
	for idx, item := range raw {
		text := strings.TrimSpace(item.Summary)
		if text == "" {
			return nil, fmt.Errorf("summary segment %d is empty", idx+1)
		}
		sourceIDs := make([]string, 0, len(item.SourceIDs))
		for _, sourceID := range item.SourceIDs {
			sourceID = strings.TrimSpace(sourceID)
			if sourceID == "" {
				continue
			}
			if _, ok := validIDs[sourceID]; !ok {
				return nil, fmt.Errorf("summary segment %d references unknown transcript segment %q", idx+1, sourceID)
			}
			if _, exists := seen[sourceID]; exists {
				return nil, fmt.Errorf("summary segment %d repeats transcript segment %q", idx+1, sourceID)
			}
			seen[sourceID] = struct{}{}
			sourceIDs = append(sourceIDs, sourceID)
		}
		if len(sourceIDs) == 0 {
			return nil, fmt.Errorf("summary segment %d has no source transcript segments", idx+1)
		}
		result = append(result, domain.SummaryResultSegment{
			OrderIndex:                 idx + 1,
			Text:                       text,
			SourceTranscriptSegmentIDs: sourceIDs,
		})
	}
	return result, nil
}

// alignSummariesToParagraphs enforces a 1:1 mapping and fills gaps with the source paragraph text.
func alignSummariesToParagraphs(summaries []string, paragraphs []string) []string {
	out := make([]string, len(paragraphs))
	for i := range paragraphs {
		if i < len(summaries) {
			if cleaned := strings.TrimSpace(strings.Trim(summaries[i], "\"")); cleaned != "" {
				out[i] = cleaned
				continue
			}
		}
		// Fallback: use the paragraph text (trimmed) to avoid empty summaries.
		p := strings.TrimSpace(paragraphs[i])
		if len(p) > 600 {
			p = p[:600] + "..."
		}
		out[i] = p
	}
	return out
}
