package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tech-digest/ingest/config"
	"github.com/tech-digest/ingest/db"
	"github.com/tech-digest/ingest/feeds"
	"github.com/tech-digest/ingest/ingest"
	"github.com/tech-digest/ingest/retry"
	"github.com/tech-digest/ingest/server"
)

func main() {
	serveMode := flag.Bool("serve", false, "Run as HTTP server")
	retryMode := flag.Bool("retry", false, "Retry failed summaries")
	retryDays := flag.Int("days", 0, "Days back for retry (default: RETRY_DAYS env, 30)")
	retryMax := flag.Int("max", 0, "Max items for retry (default: RETRY_MAX_ITEMS env, 50)")
	retryMaxRetries := flag.Int("max-retries", 0, "Max Gemini retries per item after cooldown (default: 3)")
	feedsPath := flag.String("feeds", "", "Path to feeds.json (overrides FEEDS_PATH env)")
	flag.Parse()

	// Load .env.local then .env (ignore missing files)
	_ = godotenv.Overload(".env.local")
	_ = godotenv.Load(".env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if *feedsPath != "" {
		cfg.FeedsPath = *feedsPath
	}

	ctx := context.Background()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	// --retry mode
	if *retryMode {
		result, err := retry.Run(ctx, cfg, database, retry.Options{
			Days:       *retryDays,
			MaxItems:   *retryMax,
			MaxRetries: *retryMaxRetries,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "retry error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("updated=%d errors=%d\n", result.Updated, result.Errors)
		return
	}

	feedEntries, err := feeds.Load(cfg.FeedsPath)
	if err != nil {
		log.Fatalf("loading feeds: %v", err)
	}

	// --serve mode
	if *serveMode {
		if err := server.Start(cfg, database, feedEntries); err != nil {
			log.Fatalf("server: %v", err)
		}
		return
	}

	// CLI ingest mode: optional positional argument overrides IngestDate
	if flag.NArg() > 0 {
		cfg.IngestDate = flag.Arg(0)
	}

	result, err := ingest.Run(ctx, cfg, database, feedEntries, ingest.Options{
		TargetDate: cfg.IngestDate,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ingest error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("new=%d errors=%d\n", result.NewCount, result.ErrorCount)
}
