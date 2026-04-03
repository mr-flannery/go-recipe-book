ALTER TABLE extraction_jobs DROP COLUMN IF EXISTS retry_after;

DROP INDEX IF EXISTS idx_extraction_jobs_retry;
