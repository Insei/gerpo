-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    done        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS tasks_created_at_idx ON tasks (created_at DESC);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION tasks_touch_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS tasks_touch ON tasks;
CREATE TRIGGER tasks_touch
    BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE PROCEDURE tasks_touch_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS tasks_touch ON tasks;
DROP FUNCTION IF EXISTS tasks_touch_updated_at();
DROP TABLE IF EXISTS tasks;
