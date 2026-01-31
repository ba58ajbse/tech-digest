import type { NormalizedItem, ParsedItem } from './types';

export function normalizeItem(item: ParsedItem): NormalizedItem | null {
  const title = cleanText(item.title);
  const link = cleanText(item.link);

  if (!title || !link) {
    return null;
  }

  const guid = cleanText(item.guid) ?? link;
  const publishedAt = parseDate(item.isoDate ?? item.pubDate ?? item.dcDate);
  const summaryRaw = pickFirst(
    item.content,
    item.contentSnippet,
    item.summary,
    item.description,
    title
  );

  return {
    guid,
    title,
    link,
    publishedAt: publishedAt ? publishedAt.toISOString() : null,
    summaryRaw: summaryRaw ? cleanText(summaryRaw) : null
  };
}

function pickFirst(...values: Array<string | undefined | null>): string | null {
  for (const value of values) {
    if (value && value.trim().length > 0) {
      return value;
    }
  }
  return null;
}

function cleanText(value?: string | null) {
  if (!value) return null;
  return value.replace(/\s+/g, ' ').trim();
}

function parseDate(value?: string | null) {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}
