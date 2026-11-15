-- +goose Up
ALTER TABLE events ADD COLUMN kind text NOT NULL DEFAULT 'festival';

-- +goose Down
ALTER TABLE events DROP COLUMN kind;
