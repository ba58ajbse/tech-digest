package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/tech-digest/ingest/config"
	"github.com/tech-digest/ingest/db"
	"github.com/tech-digest/ingest/feeds"
	"github.com/tech-digest/ingest/ingest"
)

func Start(cfg *config.AppConfig, database *db.DB, entries []feeds.FeedEntry) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", handler(cfg, database, entries))

	addr := ":" + cfg.Port
	log.Printf("[server] listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func handler(cfg *config.AppConfig, database *db.DB, entries []feeds.FeedEntry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authenticate
		if cfg.CronSecret != "" {
			secret := r.Header.Get("x-cron-secret")
			if secret == "" {
				secret = r.URL.Query().Get("secret")
			}
			if secret != cfg.CronSecret {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}

		targetDate := r.URL.Query().Get("date")

		result, err := ingest.Run(r.Context(), cfg, database, entries, ingest.Options{
			TargetDate: targetDate,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{
			"new_count":   result.NewCount,
			"error_count": result.ErrorCount,
		})
	}
}
