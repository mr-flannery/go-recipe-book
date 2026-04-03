ALTER TABLE extraction_jobs ADD COLUMN retry_after TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_extraction_jobs_retry ON extraction_jobs(status, retry_after) WHERE status = 'pending';
