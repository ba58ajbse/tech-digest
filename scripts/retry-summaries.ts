import 'dotenv/config';
import { getSupabaseAdminClient } from '../src/lib/db/supabase';
import { summarizeToJapanese } from '../src/lib/ingest/summarize';

type ArticleRow = {
  id: string;
  title: string;
  summary_raw: string | null;
  summary_ja: string | null;
  language: string | null;
  created_at: string | null;
};

async function retrySummaries() {
  const supabase = getSupabaseAdminClient();
  const minIntervalMs = Number(process.env.GEMINI_MIN_INTERVAL_MS ?? '400');
  const cooldownMs = Number(process.env.GEMINI_COOLDOWN_MS ?? '60000');
  const retryDays = Number(process.env.RETRY_DAYS ?? '30');
  const retryMax = Number(process.env.RETRY_MAX_ITEMS ?? '50');

  let lastGeminiCallAt = 0;
  let geminiCooldownUntil = 0;

  const since = new Date();
  since.setDate(since.getDate() - retryDays);

  const { data, error } = await supabase
    .from('articles')
    .select('id,title,summary_raw,summary_ja,language,created_at')
    .gte('created_at', since.toISOString())
    .or('language.eq.unknown,summary_ja.is.null')
    .not('summary_raw', 'is', null)
    .order('created_at', { ascending: false })
    .limit(retryMax);

  if (error || !data) {
    throw error ?? new Error('Failed to load retry candidates');
  }

  let updated = 0;
  let errors = 0;

  for (const row of data as ArticleRow[]) {
    if (!row.summary_raw) continue;

    if (Date.now() < geminiCooldownUntil) {
      continue;
    }

    const waitMs = Math.max(0, minIntervalMs - (Date.now() - lastGeminiCallAt));
    if (waitMs > 0) {
      await sleep(waitMs);
    }

    try {
      const summary = await summarizeToJapanese({ title: row.title, text: row.summary_raw });
      lastGeminiCallAt = Date.now();

      const { error: updateError } = await supabase
        .from('articles')
        .update({ summary_ja: summary.summaryJa, language: summary.language })
        .eq('id', row.id);

      if (updateError) {
        errors += 1;
      } else {
        updated += 1;
      }
    } catch (err) {
      errors += 1;
      if (isRateLimitError(err)) {
        geminiCooldownUntil = Date.now() + cooldownMs;
      }
    }
  }

  console.log(`[retry] updated: ${updated}, errors: ${errors}`);
}

function isRateLimitError(error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  return message.includes('429');
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

retrySummaries()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error('[retry] failed', error);
    process.exit(1);
  });
