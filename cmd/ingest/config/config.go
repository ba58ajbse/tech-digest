package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type AppConfig struct {
	DatabaseURL       string
	GeminiAPIKey      string
	GeminiModel       string
	GeminiMinInterval time.Duration
	GeminiCooldown    time.Duration
	CronSecret        string
	MaxItemsPerFeed   int
	RetryDays         int
	RetryMaxItems     int
	IngestDate        string
	FeedsPath         string
	Port              string
}

func Load() (*AppConfig, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	cfg := &AppConfig{
		DatabaseURL:       databaseURL,
		GeminiAPIKey:      os.Getenv("GEMINI_API_KEY"),
		GeminiModel:       getEnvOrDefault("GEMINI_MODEL", "gemini-2.5-flash-lite"),
		GeminiMinInterval: parseDurationMs("GEMINI_MIN_INTERVAL_MS", 400),
		GeminiCooldown:    parseDurationMs("GEMINI_COOLDOWN_MS", 60000),
		CronSecret:        os.Getenv("CRON_SECRET"),
		MaxItemsPerFeed:   parseInt("MAX_ITEMS_PER_FEED", 0),
		RetryDays:         parseInt("RETRY_DAYS", 30),
		RetryMaxItems:     parseInt("RETRY_MAX_ITEMS", 50),
		IngestDate:        os.Getenv("INGEST_DATE"),
		FeedsPath:         getEnvOrDefault("FEEDS_PATH", "feeds.json"),
		Port:              getEnvOrDefault("PORT", "8080"),
	}
	return cfg, nil
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseDurationMs(key string, defMs int64) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(defMs) * time.Millisecond
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return time.Duration(defMs) * time.Millisecond
	}
	return time.Duration(n) * time.Millisecond
}
