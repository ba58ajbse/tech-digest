import { loadFeedsConfig } from '@/lib/config/feeds';
import { getLatestArticles } from '@/lib/db/articles';
import TopicTabs from '@/components/topic-tabs';
import ThemeToggle from '@/components/theme-toggle';

export const dynamic = 'force-dynamic';

export default async function HomePage() {
  const config = await loadFeedsConfig();
  const articles = await getLatestArticles({ days: 30, limit: 240 });

  return (
    <main className="page">
      <section className="hero">
        <div className="hero-header">
          <p className="eyebrow">Daily Briefing</p>
          <ThemeToggle />
        </div>
        <h1>Tech Digest</h1>
        <p className="subhead">
          RSSベースの技術動向を日本語で要約。気になる記事は原文へ。
        </p>
        <div className="hero-meta">
          <span>更新頻度: 1日1回</span>
          <span>要約: Gemini</span>
        </div>
      </section>

      <TopicTabs topics={config.topics} articles={articles} />
    </main>
  );
}
