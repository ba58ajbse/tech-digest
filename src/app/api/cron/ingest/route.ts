import { NextRequest } from 'next/server';
import { runIngest } from '@/lib/ingest/ingest';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  const secret = process.env.CRON_SECRET;
  const token = request.headers.get('x-cron-secret') ?? request.nextUrl.searchParams.get('secret');

  if (secret && token !== secret) {
    return Response.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const result = await runIngest();
  return Response.json(result);
}
