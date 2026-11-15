-- +goose Up
ALTER TABLE events ADD COLUMN title text;
UPDATE events SET title = name WHERE title IS NULL;

-- +goose Down
ALTER TABLE events DROP COLUMN title;
