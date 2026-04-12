package ingest

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/tech-digest/ingest/config"
	"github.com/tech-digest/ingest/db"
	"github.com/tech-digest/ingest/feeds"
	"github.com/tech-digest/ingest/fetch"
	"github.com/tech-digest/ingest/normalize"
	"github.com/tech-digest/ingest/parse"
	"github.com/tech-digest/ingest/summarize"
)

type Options struct {
	TargetDate string // YYYY-MM-DD, empty = yesterday
}

type Result struct {
	NewCount   int
	ErrorCount int
}

func Run(ctx context.Context, cfg *config.AppConfig, database *db.DB, entries []feeds.FeedEntry, opts Options) (*Result, error) {
	if err := syncTopicsAndSources(ctx, database, entries); err != nil {
		return nil, fmt.Errorf("syncing topics/sources: %w", err)
	}

	targetRange, err := getTargetRange(opts.TargetDate)
	if err != nil {
		return nil, err
	}

	runID, err := database.CreateRun(ctx, time.Now())
	if err != nil {
		log.Printf("[ingest] failed to create run record: %v", err)
	}

	summarizer := summarize.New(
		cfg.GeminiAPIKey,
		cfg.GeminiModel,
		cfg.GeminiMinInterval,
		cfg.GeminiCooldown,
	)

	var newCount, errorCount int

	for _, entry := range entries {
		if err := processEntry(ctx, cfg, database, summarizer, entry, targetRange, &newCount, &errorCount); err != nil {
			errorCount++
			log.Printf("[ingest] feed failed feed=%s: %v", entry.Feed.ID, err)
		}
	}

	// Mark sources not in feeds.json as inactive
	activeIDs := make([]string, 0, len(entries))
	for _, e := range entries {
		activeIDs = append(activeIDs, e.Feed.ID)
	}
	if err := database.MarkInactiveSources(ctx, activeIDs); err != nil {
		log.Printf("[ingest] failed to mark inactive sources: %v", err)
	}

	if runID != "" {
		if err := database.FinishRun(ctx, runID, newCount, errorCount); err != nil {
			log.Printf("[ingest] failed to finish run record: %v", err)
		}
	}

	return &Result{NewCount: newCount, ErrorCount: errorCount}, nil
}

func processEntry(
	ctx context.Context,
	cfg *config.AppConfig,
	database *db.DB,
	summarizer *summarize.Summarizer,
	entry feeds.FeedEntry,
	targetRange targetRangeT,
	newCount, errorCount *int,
) error {
	etag, lastModified, err := database.GetSourceCacheHeaders(ctx, entry.Feed.ID)
	if err != nil {
		return fmt.Errorf("getting cache headers: %w", err)
	}

	fetched, err := fetch.Fetch(entry.Feed.URL, etag, lastModified)
	if err != nil {
		return fmt.Errorf("fetching feed: %w", err)
	}
	if fetched.NotModified {
		log.Printf("[ingest] not modified feed=%s", entry.Feed.ID)
		return nil
	}

	parsedItems, err := parse.Parse(fetched.XML, entry.Feed.Format)
	if err != nil {
		return fmt.Errorf("parsing feed: %w", err)
	}

	var normalized []*normalize.Item
	for _, pi := range parsedItems {
		ni := normalize.Normalize(pi)
		if ni != nil {
			normalized = append(normalized, ni)
		}
	}

	// Sort descending by published date
	sort.Slice(normalized, func(i, j int) bool {
		ti := timeVal(normalized[i].PublishedAt)
		tj := timeVal(normalized[j].PublishedAt)
		return ti.After(tj)
	})

	// Filter to target date range
	var filtered []*normalize.Item
	for _, item := range normalized {
		if isWithinRange(item.PublishedAt, targetRange) {
			filtered = append(filtered, item)
		}
	}

	// Apply per-feed limit
	if cfg.MaxItemsPerFeed > 0 && len(filtered) > cfg.MaxItemsPerFeed {
		filtered = filtered[:cfg.MaxItemsPerFeed]
	}

	for _, item := range filtered {
		exists, err := database.ArticleExistsByURLOrGUID(ctx, item.Link, item.GUID)
		if err != nil {
			*errorCount++
			log.Printf("[ingest] lookup failed feed=%s url=%s: %v", entry.Feed.ID, item.Link, err)
			continue
		}
		if exists {
			continue
		}

		summaryRaw := item.SummaryRaw
		if summaryRaw == "" {
			summaryRaw = item.Title
		}

		summaryResult, _, sumErr := summarizer.Summarize(ctx, item.Title, summaryRaw)
		var summaryJa, language string
		if sumErr != nil {
			*errorCount++
			log.Printf("[ingest] summarize failed feed=%s url=%s: %v", entry.Feed.ID, item.Link, sumErr)
			fb := summarize.Fallback(summaryRaw)
			summaryJa = fb.SummaryJa
			language = fb.Language
		} else {
			summaryJa = summaryResult.SummaryJa
			language = summaryResult.Language
		}

		if err := database.InsertArticle(ctx, db.Article{
			SourceID:    entry.Feed.ID,
			TopicID:     entry.TopicID,
			GUID:        item.GUID,
			Title:       item.Title,
			URL:         item.Link,
			PublishedAt: item.PublishedAt,
			SummaryJa:   summaryJa,
			SummaryRaw:  summaryRaw,
			Language:    language,
		}); err != nil {
			*errorCount++
			log.Printf("[ingest] insert failed feed=%s url=%s: %v", entry.Feed.ID, item.Link, err)
		} else {
			*newCount++
		}
	}

	// Update ETag cache headers
	if err := database.UpdateSourceCacheHeaders(ctx, entry.Feed.ID, fetched.ETag, fetched.LastModified); err != nil {
		log.Printf("[ingest] failed to update cache headers feed=%s: %v", entry.Feed.ID, err)
	}

	return nil
}

type targetRangeT struct {
	start time.Time
	end   time.Time
}

func getTargetRange(targetDate string) (targetRangeT, error) {
	var base time.Time
	if targetDate != "" {
		parsed, err := parseLocalDate(targetDate)
		if err != nil {
			return targetRangeT{}, fmt.Errorf("invalid target date %q: %w", targetDate, err)
		}
		base = parsed
	} else {
		base = time.Now().AddDate(0, 0, -1)
	}

	loc := time.Local
	start := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, loc)
	end := time.Date(base.Year(), base.Month(), base.Day(), 23, 59, 59, 999000000, loc)
	return targetRangeT{start: start, end: end}, nil
}

func parseLocalDate(value string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func isWithinRange(t *time.Time, r targetRangeT) bool {
	if t == nil {
		return false
	}
	return !t.Before(r.start) && !t.After(r.end)
}

func timeVal(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func syncTopicsAndSources(ctx context.Context, database *db.DB, entries []feeds.FeedEntry) error {
	// Deduplicate topics
	seen := make(map[string]bool)
	var topics []db.Topic
	for _, e := range entries {
		if !seen[e.TopicID] {
			topics = append(topics, db.Topic{ID: e.TopicID, Label: e.TopicLabel})
			seen[e.TopicID] = true
		}
	}

	if err := database.UpsertTopics(ctx, topics); err != nil {
		return err
	}

	var sources []db.Source
	for _, e := range entries {
		sources = append(sources, db.Source{
			ID:      e.Feed.ID,
			Name:    e.Feed.Name,
			URL:     e.Feed.URL,
			Format:  e.Feed.Format,
			TopicID: e.TopicID,
		})
	}

	return database.UpsertSources(ctx, sources)
}
