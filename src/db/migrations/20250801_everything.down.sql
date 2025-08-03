-- Migration: Revert all changes made by the initial schema and default users migrations

-- Delete all rows from the users table
DELETE FROM users WHERE username IN ('alice', 'bob');

-- Drop all tables in reverse order of creation to handle dependencies
DROP TABLE IF EXISTS proposed_changes;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS recipe_labels;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS recipes;
DROP TABLE IF EXISTS users;
