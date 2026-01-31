export type FetchFeedResult = {
  xml: string | null;
  status: number;
  etag?: string;
  lastModified?: string;
};

export async function fetchFeed(url: string, etag?: string | null, lastModified?: string | null): Promise<FetchFeedResult> {
  const headers: HeadersInit = {};
  if (etag) headers['If-None-Match'] = etag;
  if (lastModified) headers['If-Modified-Since'] = lastModified;

  const response = await fetch(url, { headers });
  if (response.status === 304) {
    return { xml: null, status: 304, etag: etag ?? undefined, lastModified: lastModified ?? undefined };
  }

  const xml = await response.text();
  return {
    xml,
    status: response.status,
    etag: response.headers.get('etag') ?? undefined,
    lastModified: response.headers.get('last-modified') ?? undefined
  };
}
