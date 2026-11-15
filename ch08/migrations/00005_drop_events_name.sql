-- +goose Up
ALTER TABLE events DROP COLUMN name;

-- +goose Down
ALTER TABLE events ADD COLUMN name text;
UPDATE events SET name = title;
ALTER TABLE events ALTER COLUMN name SET NOT NULL;
