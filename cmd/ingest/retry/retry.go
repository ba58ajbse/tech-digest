package retry

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tech-digest/ingest/config"
	"github.com/tech-digest/ingest/db"
	"github.com/tech-digest/ingest/summarize"
)

type Options struct {
	Days       int // default: RETRY_DAYS env (30)
	MaxItems   int // default: RETRY_MAX_ITEMS env (50)
	MaxRetries int // max Gemini retries per item after cooldown (default: 3)
}

type Result struct {
	Updated int
	Errors  int
}

func Run(ctx context.Context, cfg *config.AppConfig, database *db.DB, opts Options) (*Result, error) {
	days := opts.Days
	if days <= 0 {
		days = cfg.RetryDays
	}
	maxItems := opts.MaxItems
	if maxItems <= 0 {
		maxItems = cfg.RetryMaxItems
	}
	maxRetries := opts.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	since := time.Now().AddDate(0, 0, -days)

	candidates, err := database.GetRetryCandidates(ctx, since, maxItems)
	if err != nil {
		return nil, fmt.Errorf("fetching retry candidates: %w", err)
	}

	log.Printf("[retry] candidates: %d", len(candidates))

	summarizer := summarize.New(
		cfg.GeminiAPIKey,
		cfg.GeminiModel,
		cfg.GeminiMinInterval,
		cfg.GeminiCooldown,
	)

	var updated, errors int

	for _, row := range candidates {
		if !needsRetry(row) {
			continue
		}

		result, err := summarizeWithCooldownRetry(ctx, summarizer, row, maxRetries)
		if err != nil {
			errors++
			log.Printf("[retry] summarize failed id=%s url=%s: %v", row.ID, row.URL, err)
			continue
		}

		if err := database.UpdateArticleSummary(ctx, row.ID, result.SummaryJa, result.Language); err != nil {
			errors++
			log.Printf("[retry] update failed id=%s url=%s: %v", row.ID, row.URL, err)
		} else {
			updated++
			log.Printf("[retry] updated id=%s url=%s", row.ID, row.URL)
		}
	}

	log.Printf("[retry] updated=%d errors=%d", updated, errors)
	return &Result{Updated: updated, Errors: errors}, nil
}

func needsRetry(row db.RetryCandidate) bool {
	if row.SummaryJa == "" {
		return true
	}
	if row.Language == "unknown" {
		return true
	}
	// summary_ja が summary_raw と同じ = Gemini が未処理
	if strings.TrimSpace(row.SummaryJa) == strings.TrimSpace(row.SummaryRaw) {
		return true
	}
	if !containsJapanese(row.SummaryJa) {
		return true
	}
	return false
}

// summarizeWithCooldownRetry calls Summarize and, when the summarizer is in
// cooldown (skipped=true), sleeps until the cooldown expires and retries.
// maxRetries is the maximum number of attempts (including the first call).
func summarizeWithCooldownRetry(ctx context.Context, s *summarize.Summarizer, row db.RetryCandidate, maxRetries int) (*summarize.Result, error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, skipped, err := s.Summarize(ctx, row.Title, row.SummaryRaw)
		if err != nil {
			return nil, err
		}
		if !skipped {
			return result, nil
		}

		if attempt == maxRetries {
			return nil, fmt.Errorf("max retries (%d) exceeded due to cooldown", maxRetries)
		}

		remaining := s.CooldownRemaining()
		if remaining > 0 {
			log.Printf("[retry] cooling down for %s, retry %d/%d id=%s", remaining.Round(time.Second), attempt, maxRetries, row.ID)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(remaining):
			}
		}
	}
	return nil, fmt.Errorf("max retries (%d) exceeded", maxRetries)
}

func containsJapanese(text string) bool {
	for _, r := range text {
		if (r >= 0x3040 && r <= 0x30FF) || (r >= 0x4E00 && r <= 0x9FAF) {
			return true
		}
	}
	return false
}
