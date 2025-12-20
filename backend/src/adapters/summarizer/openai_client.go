package summarizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAISummarizer calls OpenAI chat completions to summarize paragraphs.
type OpenAISummarizer struct {
	APIKey string
	Model  string
	Client *http.Client
}

type structuredSummary struct {
	Discussion string `json:"discussion"`
}

func NewOpenAISummarizer(apiKey, model string) *OpenAISummarizer {
	return &OpenAISummarizer{
		APIKey: apiKey,
		Model:  model,
		Client: &http.Client{Timeout: 60 * time.Second},
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

func (s *OpenAISummarizer) summarizeBatch(paragraphs []string) ([]string, error) {
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
	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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

	summaries, structured := parseSummaries(content)
	if len(summaries) == 0 {
		return nil, fmt.Errorf("unable to parse summaries from OpenAI response")
	}

	summaries = alignSummariesToParagraphs(summaries, paragraphs, structured)
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
func parseSummaries(content string) ([]string, bool) {
	var summaries []string

	// 0) Preferred: array of structured objects we can store as canonical JSON per paragraph.
	var structuredArr []structuredSummary
	if err := json.Unmarshal([]byte(content), &structuredArr); err == nil && len(structuredArr) > 0 {
		for _, v := range structuredArr {
			v.Discussion = strings.TrimSpace(v.Discussion)
			b, err := json.Marshal(v)
			if err != nil {
				continue
			}
			summaries = append(summaries, string(b))
		}
		if len(summaries) > 0 {
			return summaries, true
		}
	}

	// 1) Straight array of strings.
	var arr []string
	if err := json.Unmarshal([]byte(content), &arr); err == nil && len(arr) > 0 {
		for _, v := range arr {
			summaries = append(summaries, strings.TrimSpace(v))
		}
		return summaries, false
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

	return summaries, false
}

// alignSummariesToParagraphs enforces a 1:1 mapping and fills gaps with the source paragraph text.
// When structured=true, fallbacks are emitted as a canonical JSON object string.
func alignSummariesToParagraphs(summaries []string, paragraphs []string, structured bool) []string {
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
		if structured {
			b, err := json.Marshal(structuredSummary{Discussion: p})
			if err == nil {
				out[i] = string(b)
			} else {
				out[i] = p
			}
			continue
		}
		out[i] = p
	}
	return out
}
