import Parser from 'rss-parser';
import { XMLParser } from 'fast-xml-parser';
import type { ParsedItem } from './types';
import type { FeedFormat } from '../config/feeds';

const rssParser = new Parser();

export async function parseFeedXml(xml: string, format: FeedFormat): Promise<ParsedItem[]> {
  if (format === 'rdf') {
    return parseRdf(xml);
  }

  const feed = await rssParser.parseString(xml);
  return (feed.items ?? []).map((item) => ({
    title: item.title ?? undefined,
    link: item.link ?? undefined,
    guid: item.guid ?? undefined,
    pubDate: item.pubDate ?? item.isoDate ?? undefined,
    isoDate: item.isoDate ?? undefined,
    content: (item as Record<string, unknown>)['content:encoded'] as string | undefined,
    contentSnippet: item.contentSnippet ?? undefined,
    summary: item.summary ?? undefined,
    description: item.contentSnippet ?? undefined
  }));
}

function parseRdf(xml: string): ParsedItem[] {
  const parser = new XMLParser({
    ignoreAttributes: false,
    attributeNamePrefix: '@_'
  });
  const doc = parser.parse(xml) as Record<string, any>;
  const root = doc['rdf:RDF'] ?? doc['rss'] ?? doc;
  const items = root?.item ?? root?.channel?.item ?? [];
  const list = Array.isArray(items) ? items : [items];

  return list
    .filter(Boolean)
    .map((item) => ({
      title: asText(item.title),
      link: asText(item.link),
      guid: asText(item.guid ?? item['@_rdf:about']),
      pubDate: asText(item['dc:date'] ?? item.pubDate),
      isoDate: asText(item['dc:date'] ?? item.isoDate),
      description: asText(item.description ?? item['content:encoded'])
    }));
}

function asText(value: unknown): string | undefined {
  if (!value) return undefined;
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return String(value);
  if (typeof value === 'object' && '#text' in (value as Record<string, unknown>)) {
    const text = (value as Record<string, unknown>)['#text'];
    return typeof text === 'string' ? text : undefined;
  }
  return undefined;
}
