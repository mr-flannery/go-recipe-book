-- Migration: Make email column NOT NULL
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
