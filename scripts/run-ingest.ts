import 'dotenv/config';
import { runIngest } from '../src/lib/ingest/ingest';

const targetDate = process.argv[2] ?? process.env.INGEST_DATE ?? null;

runIngest({ targetDate })
  .then((result) => {
    console.log(`[ingest] new: ${result.newCount}, errors: ${result.errorCount}`);
    process.exit(0);
  })
  .catch((error) => {
    console.error('[ingest] failed', error);
    process.exit(1);
  });
