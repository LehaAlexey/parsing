CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_ux ON users (email);

CREATE TABLE IF NOT EXISTS user_urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    normalized_url TEXT NOT NULL,
    polling_interval_seconds INT NOT NULL DEFAULT 3600,
    next_run_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS user_urls_user_norm_ux ON user_urls (user_id, normalized_url);
CREATE INDEX IF NOT EXISTS user_urls_user_id_idx ON user_urls (user_id);
CREATE INDEX IF NOT EXISTS user_urls_next_run_idx ON user_urls (next_run_at);
