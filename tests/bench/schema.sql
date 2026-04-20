-- Minimal schema for the real-PG bench report. One table, one index, one
-- representative shape (UUID pk, text, nullable text, int, timestamptz) —
-- enough to exercise every CRUD path without adding noise from FKs / joins
-- that aren't being measured.

DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    email       TEXT,
    age         INT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX users_age_idx ON users(age);
CREATE INDEX users_created_at_idx ON users(created_at DESC);
