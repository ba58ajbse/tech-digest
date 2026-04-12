package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (d *DB) Close() {
	d.pool.Close()
}

// --- Topics & Sources ---

type Topic struct {
	ID    string
	Label string
}

type Source struct {
	ID      string
	Name    string
	URL     string
	Format  string
	TopicID string
}

func (d *DB) UpsertTopics(ctx context.Context, topics []Topic) error {
	for _, t := range topics {
		_, err := d.pool.Exec(ctx,
			`INSERT INTO topics (id, label) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET label = EXCLUDED.label`,
			t.ID, t.Label,
		)
		if err != nil {
			return fmt.Errorf("upserting topic %s: %w", t.ID, err)
		}
	}
	return nil
}

func (d *DB) UpsertSources(ctx context.Context, sources []Source) error {
	for _, s := range sources {
		_, err := d.pool.Exec(ctx,
			`INSERT INTO sources (id, name, url, format, topic_id, active)
			 VALUES ($1, $2, $3, $4, $5, true)
			 ON CONFLICT (id) DO UPDATE SET
			   name = EXCLUDED.name,
			   url = EXCLUDED.url,
			   format = EXCLUDED.format,
			   topic_id = EXCLUDED.topic_id,
			   active = true`,
			s.ID, s.Name, s.URL, s.Format, s.TopicID,
		)
		if err != nil {
			return fmt.Errorf("upserting source %s: %w", s.ID, err)
		}
	}
	return nil
}

func (d *DB) MarkInactiveSources(ctx context.Context, activeIDs []string) error {
	if len(activeIDs) == 0 {
		return nil
	}
	_, err := d.pool.Exec(ctx,
		`UPDATE sources SET active = false WHERE NOT (id = ANY($1::text[]))`,
		activeIDs,
	)
	return err
}

func (d *DB) GetSourceCacheHeaders(ctx context.Context, sourceID string) (etag, lastModified string, err error) {
	row := d.pool.QueryRow(ctx,
		`SELECT COALESCE(etag, ''), COALESCE(last_modified, '') FROM sources WHERE id = $1`,
		sourceID,
	)
	err = row.Scan(&etag, &lastModified)
	if err == pgx.ErrNoRows {
		return "", "", nil
	}
	return etag, lastModified, err
}

func (d *DB) UpdateSourceCacheHeaders(ctx context.Context, sourceID, etag, lastModified string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE sources SET etag = $1, last_modified = $2 WHERE id = $3`,
		nullableStr(etag), nullableStr(lastModified), sourceID,
	)
	return err
}

// --- Articles ---

type Article struct {
	SourceID    string
	TopicID     string
	GUID        string
	Title       string
	URL         string
	PublishedAt *time.Time
	SummaryJa   string
	SummaryRaw  string
	Language    string
}

func (d *DB) ArticleExistsByURLOrGUID(ctx context.Context, url, guid string) (bool, error) {
	var exists bool
	err := d.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM articles WHERE url = $1 OR guid = $2)`,
		url, guid,
	).Scan(&exists)
	return exists, err
}

type RetryCandidate struct {
	ID         string
	Title      string
	URL        string
	SummaryRaw string
	SummaryJa  string
	Language   string
}

func (d *DB) GetRetryCandidates(ctx context.Context, since time.Time, limit int) ([]RetryCandidate, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT id, title, COALESCE(url,''), COALESCE(summary_raw,''), COALESCE(summary_ja,''), COALESCE(language,'')
		 FROM articles
		 WHERE created_at >= $1
		   AND summary_raw IS NOT NULL
		   AND (language = 'unknown' OR summary_ja IS NULL)
		 ORDER BY created_at DESC
		 LIMIT $2`,
		since, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []RetryCandidate
	for rows.Next() {
		var c RetryCandidate
		if err := rows.Scan(&c.ID, &c.Title, &c.URL, &c.SummaryRaw, &c.SummaryJa, &c.Language); err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

func (d *DB) UpdateArticleSummary(ctx context.Context, id, summaryJa, language string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE articles SET summary_ja = $1, language = $2 WHERE id = $3`,
		summaryJa, language, id,
	)
	return err
}

func (d *DB) InsertArticle(ctx context.Context, a Article) error {
	_, err := d.pool.Exec(ctx,
		`INSERT INTO articles
		   (source_id, topic_id, guid, title, url, published_at, summary_ja, summary_raw, language)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		a.SourceID, a.TopicID, a.GUID, a.Title, a.URL,
		a.PublishedAt, a.SummaryJa, a.SummaryRaw, a.Language,
	)
	return err
}

// --- Runs ---

func (d *DB) CreateRun(ctx context.Context, startedAt time.Time) (string, error) {
	var id string
	err := d.pool.QueryRow(ctx,
		`INSERT INTO runs (started_at, new_count, error_count) VALUES ($1, 0, 0) RETURNING id`,
		startedAt,
	).Scan(&id)
	return id, err
}

func (d *DB) FinishRun(ctx context.Context, runID string, newCount, errorCount int) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE runs SET finished_at = now(), new_count = $1, error_count = $2 WHERE id = $3`,
		newCount, errorCount, runID,
	)
	return err
}

// --- helpers ---

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
