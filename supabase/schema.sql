create extension if not exists pgcrypto;

create table if not exists topics (
  id text primary key,
  label text not null
);

create table if not exists sources (
  id text primary key,
  name text not null,
  url text not null,
  format text not null,
  topic_id text references topics(id),
  etag text,
  last_modified text,
  active boolean default true,
  created_at timestamptz default now()
);

create table if not exists articles (
  id uuid primary key default gen_random_uuid(),
  source_id text references sources(id),
  topic_id text references topics(id),
  guid text,
  title text not null,
  url text not null,
  published_at timestamptz,
  summary_ja text,
  summary_raw text,
  language text,
  created_at timestamptz default now()
);

create table if not exists runs (
  id uuid primary key default gen_random_uuid(),
  started_at timestamptz not null,
  finished_at timestamptz,
  new_count int default 0,
  error_count int default 0
);

create unique index if not exists articles_unique_url on articles(url);
create unique index if not exists articles_unique_guid on articles(guid);
create index if not exists articles_topic_published on articles(topic_id, published_at);
