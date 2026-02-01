import { getSupabaseServerClient } from './supabase';

export type Article = {
  id: string;
  title: string;
  url: string;
  summary_ja: string | null;
  summary_raw: string | null;
  language: string | null;
  published_at: string | null;
  created_at: string | null;
  topic_id: string;
  source_name: string;
};

export async function getLatestArticles(options?: { days?: number; limit?: number }) {
  const { days = 30, limit = 200 } = options ?? {};
  const supabase = getSupabaseServerClient();
  const since = new Date();
  since.setDate(since.getDate() - days);

  const { data, error } = await supabase
    .from('articles')
    .select('id,title,url,summary_ja,summary_raw,language,published_at,created_at,topic_id,sources(name)')
    .gte('published_at', since.toISOString())
    .order('published_at', { ascending: false })
    .limit(limit);

  if (error || !data) {
    return [] as Article[];
  }

  return data.map((row) => {
    const source = Array.isArray(row.sources) ? row.sources[0] : row.sources;

    return {
      id: row.id,
      title: row.title,
      url: row.url,
      summary_ja: row.summary_ja,
      summary_raw: row.summary_raw,
      language: row.language,
      published_at: row.published_at,
      created_at: row.created_at,
      topic_id: row.topic_id,
      source_name: source?.name ?? 'Unknown'
    };
  });
}
