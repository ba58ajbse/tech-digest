import { NextRequest } from 'next/server';
import { runIngest } from '@/lib/ingest/ingest';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  const secret = process.env.CRON_SECRET;
  const token = request.headers.get('x-cron-secret') ?? request.nextUrl.searchParams.get('secret');
  const targetDate = request.nextUrl.searchParams.get('date');

  if (secret && token !== secret) {
    return Response.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const result = await runIngest({ targetDate });
    return Response.json(result);
  } catch (error) {
    return Response.json({ error: (error as Error).message }, { status: 400 });
  }
}
