-- +goose Up
CREATE TABLE IF NOT EXISTS video_statuses(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL
);

INSERT INTO video_statuses (status) VALUES ('pending'), ('processing'), ('finished'), ('failed');

CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status UUID NOT NULL REFERENCES video_statuses(id),
    progress INT,
    file_path TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- +goose Down
DROP TABLE videos;
DROP TABLE video_statuses;
