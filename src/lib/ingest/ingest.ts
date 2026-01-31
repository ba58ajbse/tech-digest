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

      for (const item of normalized) {
        const { data: existing } = await supabase
          .from('articles')
          .select('id')
          .eq('url', item.link)
          .maybeSingle();

        if (existing) {
          continue;
        }

        const summaryRaw = item.summaryRaw ?? item.title;
        const summary = await summarizeToJapanese({ title: item.title, text: summaryRaw });

        const { error } = await supabase.from('articles').insert({
          source_id: entry.feed.id,
          topic_id: entry.topicId,
          guid: item.guid,
          title: item.title,
          url: item.link,
          published_at: item.publishedAt,
          summary_ja: summary.summaryJa,
          summary_raw: summaryRaw,
          language: summary.language
        });

        if (error) {
          errorCount += 1;
        } else {
          newCount += 1;
        }
      }

      await supabase
        .from('sources')
        .update({ etag: fetched.etag ?? null, last_modified: fetched.lastModified ?? null })
        .eq('id', entry.feed.id);
    } catch {
      errorCount += 1;
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
