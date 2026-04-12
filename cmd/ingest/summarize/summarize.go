package summarize

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Result struct {
	SummaryJa string
	Language  string
}

type Summarizer struct {
	apiKey      string
	model       string
	minInterval time.Duration
	cooldown    time.Duration
	lastCallAt  time.Time
	coolUntil   time.Time
	mu          sync.Mutex
}

func New(apiKey, model string, minInterval, cooldown time.Duration) *Summarizer {
	return &Summarizer{
		apiKey:      apiKey,
		model:       model,
		minInterval: minInterval,
		cooldown:    cooldown,
	}
}

// CooldownRemaining returns how long until the cooldown expires (0 if not cooling down).
func (s *Summarizer) CooldownRemaining() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	remaining := time.Until(s.coolUntil)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *Summarizer) Summarize(ctx context.Context, title, text string) (result *Result, skipped bool, err error) {
	if s.apiKey == "" {
		return Fallback(text), true, nil
	}

	s.mu.Lock()
	inCooldown := time.Now().Before(s.coolUntil)
	if !inCooldown {
		wait := s.minInterval - time.Since(s.lastCallAt)
		s.mu.Unlock()
		if wait > 0 {
			time.Sleep(wait)
		}
	} else {
		s.mu.Unlock()
		return Fallback(text), true, nil
	}

	res, err := s.callAPI(ctx, title, text)
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			s.mu.Lock()
			s.coolUntil = time.Now().Add(s.cooldown)
			s.mu.Unlock()
		}
		return nil, false, err
	}

	s.mu.Lock()
	s.lastCallAt = time.Now()
	s.mu.Unlock()

	return res, false, nil
}

func (s *Summarizer) callAPI(ctx context.Context, title, text string) (*Result, error) {
	prompt := buildPrompt(title, text)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", s.model, s.apiKey)

	body, _ := json.Marshal(map[string]any{
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]any{{"text": prompt}}},
		},
		"generationConfig": map[string]any{
			"temperature":     0.2,
			"maxOutputTokens": 256,
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !isSuccess(resp.StatusCode) {
		return nil, fmt.Errorf("Gemini API error: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("parsing Gemini response: %w", err)
	}

	var parts []string
	if len(payload.Candidates) > 0 {
		for _, p := range payload.Candidates[0].Content.Parts {
			parts = append(parts, p.Text)
		}
	}
	rawText := strings.Join(parts, "")

	parsed := safeJSONParse(rawText)
	if parsed == nil || parsed["summary_ja"] == "" {
		return Fallback(text), nil
	}

	language := parsed["language"]
	if language == "" {
		language = "unknown"
	}
	return &Result{
		SummaryJa: strings.TrimSpace(parsed["summary_ja"]),
		Language:  strings.TrimSpace(language),
	}, nil
}

func buildPrompt(title, text string) string {
	combined := (title + "\n\n" + text)
	if len(combined) > 6000 {
		combined = combined[:6000]
	}
	return strings.Join([]string{
		"You are a technical news summarizer.",
		"Detect the language of the input and output a concise Japanese summary.",
		"Return JSON only in this format:",
		`{"language":"en","summary_ja":"..."}`,
		"Guidelines:",
		"- Summary must be 2-4 sentences in Japanese.",
		"- If the input is Japanese, keep the summary in Japanese.",
		"- Do not include markdown or code fences.",
		"",
		"INPUT:\n" + combined,
	}, "\n")
}

func safeJSONParse(text string) map[string]string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || start >= end {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(text[start:end+1]), &m); err != nil {
		return nil
	}
	return m
}

func Fallback(text string) *Result {
	trimmed := strings.Join(strings.Fields(text), " ")
	if len([]rune(trimmed)) > 300 {
		runes := []rune(trimmed)
		trimmed = string(runes[:300])
	}
	lang := "unknown"
	if isJapanese(trimmed) {
		lang = "ja"
	}
	return &Result{SummaryJa: trimmed, Language: lang}
}

func isJapanese(text string) bool {
	for _, r := range text {
		if (r >= 0x3040 && r <= 0x30FF) || (r >= 0x4E00 && r <= 0x9FAF) {
			return true
		}
	}
	return false
}

func isSuccess(code int) bool {
	return code >= 200 && code < 300
}
