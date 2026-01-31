import { runIngest } from '../src/lib/ingest/ingest';

runIngest()
  .then((result) => {
    console.log(`[ingest] new: ${result.newCount}, errors: ${result.errorCount}`);
    process.exit(0);
  })
  .catch((error) => {
    console.error('[ingest] failed', error);
    process.exit(1);
  });
