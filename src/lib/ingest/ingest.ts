import { getSupabaseAdminClient } from '../db/supabase';
import { loadFeedEntries } from './load-feeds';
import { fetchFeed } from './fetch-feed';
import { parseFeedXml } from './parse-feed';
import { normalizeItem } from './normalize';
import { summarizeToJapanese } from './summarize';
import type { FeedEntry } from './types';

export type IngestResult = {
  newCount: number;
  errorCount: number;
};

export async function runIngest(): Promise<IngestResult> {
  const supabase = getSupabaseAdminClient();
  const entries = await loadFeedEntries();
  await syncTopicsAndSources(entries);

  const maxItemsPerFeed = Number(process.env.MAX_ITEMS_PER_FEED ?? '0');
  const minIntervalMs = Number(process.env.GEMINI_MIN_INTERVAL_MS ?? '400');
  const cooldownMs = Number(process.env.GEMINI_COOLDOWN_MS ?? '60000');
  let lastGeminiCallAt = 0;
  let geminiCooldownUntil = 0;

  const runStart = new Date();
  const { data: runRow } = await supabase
    .from('runs')
    .insert({ started_at: runStart.toISOString(), new_count: 0, error_count: 0 })
    .select('id')
    .single();

  let newCount = 0;
  let errorCount = 0;

  for (const entry of entries) {
    try {
      const { data: source } = await supabase
        .from('sources')
        .select('etag,last_modified')
        .eq('id', entry.feed.id)
        .single();

      const fetched = await fetchFeed(entry.feed.url, source?.etag, source?.last_modified);
      if (!fetched.xml || fetched.status === 304) {
        continue;
      }

      const items = await parseFeedXml(fetched.xml, entry.feed.format);
      const normalized = items
        .map(normalizeItem)
        .filter((item): item is NonNullable<typeof item> => Boolean(item));
      const sorted = normalized.sort((a, b) => {
        const aTime = a.publishedAt ? new Date(a.publishedAt).getTime() : 0;
        const bTime = b.publishedAt ? new Date(b.publishedAt).getTime() : 0;
        return bTime - aTime;
      });
      const limited = maxItemsPerFeed > 0 ? sorted.slice(0, maxItemsPerFeed) : sorted;

      for (const item of limited) {
        const { data: existing, error: lookupError } = await supabase
          .from('articles')
          .select('id')
          .or(`url.eq.${item.link},guid.eq.${item.guid}`)
          .maybeSingle();

        if (lookupError) {
          errorCount += 1;
          console.warn('[ingest] lookup failed', { feed: entry.feed.id, url: item.link });
          continue;
        }

        if (existing) {
          continue;
        }

        const summaryRaw = item.summaryRaw ?? item.title;
        let summaryJa = summaryRaw;
        let language = detectLanguage(summaryRaw);

        if (Date.now() < geminiCooldownUntil) {
          // Use fallback summary while cooling down after rate limit.
        } else {
          const waitMs = Math.max(0, minIntervalMs - (Date.now() - lastGeminiCallAt));
          if (waitMs > 0) {
            await sleep(waitMs);
          }

          try {
            const summary = await summarizeToJapanese({ title: item.title, text: summaryRaw });
            summaryJa = summary.summaryJa;
            language = summary.language;
            lastGeminiCallAt = Date.now();
          } catch (error) {
            errorCount += 1;
            if (isRateLimitError(error)) {
              geminiCooldownUntil = Date.now() + cooldownMs;
            }
            console.warn('[ingest] summarize failed', { feed: entry.feed.id, url: item.link }, error);
          }
        }

        const { error } = await supabase.from('articles').insert({
          source_id: entry.feed.id,
          topic_id: entry.topicId,
          guid: item.guid,
          title: item.title,
          url: item.link,
          published_at: item.publishedAt,
          summary_ja: summaryJa,
          summary_raw: summaryRaw,
          language
        });

        if (error) {
          errorCount += 1;
          console.warn('[ingest] insert failed', { feed: entry.feed.id, url: item.link }, error);
        } else {
          newCount += 1;
        }
      }

      await supabase
        .from('sources')
        .update({ etag: fetched.etag ?? null, last_modified: fetched.lastModified ?? null })
        .eq('id', entry.feed.id);
    } catch (error) {
      errorCount += 1;
      console.warn('[ingest] feed failed', { feed: entry.feed.id }, error);
    }
  }

  if (runRow?.id) {
    await supabase
      .from('runs')
      .update({ finished_at: new Date().toISOString(), new_count: newCount, error_count: errorCount })
      .eq('id', runRow.id);
  }

  return { newCount, errorCount };
}

function detectLanguage(text: string) {
  return /[\u3040-\u30ff\u4e00-\u9faf]/.test(text) ? 'ja' : 'unknown';
}

function isRateLimitError(error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  return message.includes('429');
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function syncTopicsAndSources(entries: FeedEntry[]) {
  const supabase = getSupabaseAdminClient();
  const topics = Array.from(
    new Map(entries.map((entry) => [entry.topicId, { id: entry.topicId, label: entry.topicLabel }])).values()
  );

  await supabase.from('topics').upsert(topics, { onConflict: 'id' });

  const sources = entries.map((entry) => ({
    id: entry.feed.id,
    name: entry.feed.name,
    url: entry.feed.url,
    format: entry.feed.format,
    topic_id: entry.topicId,
    active: true
  }));

  await supabase.from('sources').upsert(sources, { onConflict: 'id' });
}
