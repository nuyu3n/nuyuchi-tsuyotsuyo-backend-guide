-- +goose Up
UPDATE events SET title = name WHERE title IS NULL;
ALTER TABLE events ALTER COLUMN title SET NOT NULL;

-- +goose Down
ALTER TABLE events ALTER COLUMN title DROP NOT NULL;
