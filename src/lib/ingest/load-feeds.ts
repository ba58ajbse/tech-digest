import { loadFeedsConfig } from '../config/feeds';
import type { FeedEntry } from './types';

export async function loadFeedEntries(): Promise<FeedEntry[]> {
  const config = await loadFeedsConfig();
  return config.topics.flatMap((topic) =>
    topic.feeds.map((feed) => ({
      topicId: topic.id,
      topicLabel: topic.label,
      feed
    }))
  );
}
