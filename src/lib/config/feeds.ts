import { readFile } from 'node:fs/promises';
import path from 'node:path';

export type FeedFormat = 'rss' | 'rdf';

export type FeedConfig = {
  id: string;
  name: string;
  url: string;
  format: FeedFormat;
};

export type FeedTopic = {
  id: string;
  label: string;
  feeds: FeedConfig[];
};

export type FeedsConfig = {
  topics: FeedTopic[];
};

export async function loadFeedsConfig(): Promise<FeedsConfig> {
  const filePath = path.join(process.cwd(), 'feeds.json');
  const raw = await readFile(filePath, 'utf-8');
  return JSON.parse(raw) as FeedsConfig;
}
