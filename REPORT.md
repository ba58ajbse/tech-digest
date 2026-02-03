# Implementation Report

## 2026-01-31
- Added project scaffolding for Next.js and ingestion scripts.
- Added Supabase schema and seed files for topics, sources, and articles.
- Implemented RSS/RDF ingestion with Gemini-based summarization.
- Implemented a topic-tab UI with search and time filtering.
- Fixed ingestion script to load local `.env` values via dotenv.
- Added per-item error handling and fallback summaries to keep ingest resilient.
- Added Gemini rate-limit backoff and optional per-feed item limiting to reduce 429 errors.
- Limited ingestion to items published on the previous day (local time).
- Added a retry script to re-summarize failed articles via Gemini.
- Added retry logging and a stricter fallback-only filter for re-summarization.
- Updated retry flow to wait out Gemini rate limits instead of skipping candidates.
- Added PWA manifest, icons, and service worker registration for installable mobile experience.
- Updated UI theme to a dark color palette while keeping layout unchanged.
- Added a GitHub Actions workflow to trigger daily ingest at 09:00 JST.
- Added Vercel protection bypass header support for scheduled ingest.
- Removed NEXT_PUBLIC Supabase keys to keep database access server-side only.
- Added an RLS enablement script for Supabase with no public policies.
- Switched Google Fonts loading to CSS import to avoid build-time fetch failures.
- Fixed Supabase relation mapping for source names during article fetch.
- Moved themeColor metadata to viewport export to satisfy Next.js warnings.
- Added HTTP status logging and failure handling to the daily ingest workflow.

## 2026-02-01
- No functional changes; commit to trigger deployment.
