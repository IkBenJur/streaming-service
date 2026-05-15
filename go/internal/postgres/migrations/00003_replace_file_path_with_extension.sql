-- +goose Up
ALTER TABLE videos DROP COLUMN file_path;
ALTER TABLE videos ADD COLUMN file_extension VARCHAR(4) NOT NULL CHECK (file_extension IN ('webm', 'mp4'));

-- +goose Down
ALTER TABLE videos DROP COLUMN file_extension;
ALTER TABLE videos ADD COLUMN file_path TEXT;
