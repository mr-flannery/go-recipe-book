-- Migration: Insert default users

INSERT INTO users (username, password_hash, created_at)
VALUES
    ('alice', 'password1', CURRENT_TIMESTAMP),
    ('bob', 'password2', CURRENT_TIMESTAMP);
