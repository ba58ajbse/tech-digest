import type { FeedConfig } from '../config/feeds';

export type FeedEntry = {
  topicId: string;
  topicLabel: string;
  feed: FeedConfig;
};

export type ParsedItem = {
  title?: string;
  link?: string;
  guid?: string;
  pubDate?: string;
  isoDate?: string;
  content?: string;
  contentSnippet?: string;
  summary?: string;
  description?: string;
  dcDate?: string;
};

export type NormalizedItem = {
  guid: string;
  title: string;
  link: string;
  publishedAt: string | null;
  summaryRaw: string | null;
};
