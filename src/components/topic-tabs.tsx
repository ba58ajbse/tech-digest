'use client';

import { useMemo, useState } from 'react';
import type { FeedTopic } from '@/lib/config/feeds';
import type { Article } from '@/lib/db/articles';

type RangeOption = {
  id: string;
  label: string;
  days: number;
};

const ranges: RangeOption[] = [
  { id: '7', label: '7日', days: 7 },
  { id: '30', label: '30日', days: 30 },
  { id: '90', label: '90日', days: 90 }
];

export default function TopicTabs({ topics, articles }: { topics: FeedTopic[]; articles: Article[] }) {
  const [activeTopic, setActiveTopic] = useState(topics[0]?.id ?? '');
  const [query, setQuery] = useState('');
  const [range, setRange] = useState(ranges[1].id);

  const activeRange = ranges.find((item) => item.id === range) ?? ranges[1];

  const filtered = useMemo(() => {
    const lower = query.trim().toLowerCase();
    const since = new Date();
    since.setDate(since.getDate() - activeRange.days);

    return articles.filter((article) => {
      if (article.topic_id !== activeTopic) return false;
      const dateValue = article.published_at ?? article.created_at;
      if (dateValue && new Date(dateValue) < since) return false;

      if (!lower) return true;
      const haystack = `${article.title} ${article.summary_ja ?? ''}`.toLowerCase();
      return haystack.includes(lower);
    });
  }, [articles, activeRange.days, activeTopic, query]);

  return (
    <section className="digest">
      <header className="digest-header">
        <div>
          <p className="eyebrow">Topics</p>
          <h2>トピック別ダイジェスト</h2>
        </div>
        <div className="filters">
          <input
            className="search"
            placeholder="検索キーワード"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
          <div className="range">
            {ranges.map((item) => (
              <button
                key={item.id}
                type="button"
                className={item.id === range ? 'active' : ''}
                onClick={() => setRange(item.id)}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>
      </header>

      <div className="tabs">
        {topics.map((topic) => (
          <button
            key={topic.id}
            type="button"
            className={topic.id === activeTopic ? 'active' : ''}
            onClick={() => setActiveTopic(topic.id)}
          >
            {topic.label}
            <span>{articles.filter((article) => article.topic_id === topic.id).length}</span>
          </button>
        ))}
      </div>

      <div className="cards">
        {filtered.length === 0 ? (
          <div className="empty">該当する記事がありません。</div>
        ) : (
          filtered.map((article, index) => (
            <article className="card" key={article.id} style={{ animationDelay: `${index * 40}ms` }}>
              <div className="card-meta">
                <span>{article.source_name}</span>
                <span>•</span>
                <time>{formatDate(article.published_at ?? article.created_at)}</time>
                {article.language && article.language !== 'ja' ? (
                  <span className="lang">{article.language}</span>
                ) : null}
              </div>
              <h3>{article.title}</h3>
              <p>{article.summary_ja ?? article.summary_raw ?? ''}</p>
              <a href={article.url} target="_blank" rel="noreferrer">
                原文を読む
              </a>
            </article>
          ))
        )}
      </div>
    </section>
  );
}

function formatDate(value: string | null) {
  if (!value) return '日付不明';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '日付不明';
  return date.toLocaleDateString('ja-JP', { month: 'short', day: 'numeric' });
}
