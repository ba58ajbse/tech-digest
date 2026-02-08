insert into topics (id, label) values
  ('ai', 'AI'),
  ('go', 'Go'),
  ('github', 'GitHub'),
  ('aws', 'AWS'),
  ('security', 'セキュリティ')
on conflict (id) do update set label = excluded.label;

insert into sources (id, name, url, format, topic_id) values
  ('openai-news', 'OpenAI News', 'https://openai.com/news/rss.xml', 'rss', 'ai'),
  ('golang-weekly', 'Golang Weekly', 'https://golangweekly.com/rss/', 'rss', 'go'),
  ('github-changelog', 'GitHub Changelog', 'https://github.blog/changelog/feed/', 'rss', 'github'),
  ('aws-whats-new', 'AWS What''s New', 'https://aws.amazon.com/about-aws/whats-new/recent/feed/', 'rss', 'aws'),
  ('ipa-security-alert', 'IPA セキュリティ情報', 'https://www.ipa.go.jp/security/alert-rss.rdf', 'rdf', 'security'),
  ('jvn-ipedia-new', 'JVN iPedia (New)', 'https://jvndb.jvn.jp/en/rss/jvndb_new.rdf', 'rdf', 'security')
on conflict (id) do update set
  name = excluded.name,
  url = excluded.url,
  format = excluded.format,
  topic_id = excluded.topic_id;
