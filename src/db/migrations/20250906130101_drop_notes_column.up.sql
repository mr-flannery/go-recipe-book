-- DROP the notes column, we don't need it
ALTER TABLE registration_requests DROP COLUMN IF EXISTS notes;