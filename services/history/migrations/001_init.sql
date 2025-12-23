CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS price_history (
    event_id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    price BIGINT NOT NULL,
    currency TEXT NOT NULL,
    parsed_at TIMESTAMPTZ NOT NULL,
    source_url TEXT NOT NULL,
    meta_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS price_history_product_parsed_idx ON price_history (product_id, parsed_at DESC);

