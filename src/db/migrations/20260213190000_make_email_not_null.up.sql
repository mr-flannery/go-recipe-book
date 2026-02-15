-- Migration: Make email column NOT NULL
-- First update any existing users without email to have a generated email
UPDATE users SET email = username || '@example.com' WHERE email IS NULL;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
