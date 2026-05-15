-- name: ListVideos :many
SELECT * FROM videos;

-- name: FindVideoById :one
SELECT * FROM videos WHERE id = $1;

-- name: CreateVideo :one
INSERT INTO videos (status, file_extension) VALUES ($1, $2) RETURNING id;

-- name: FindStatusIdByName :one
SELECT id FROM video_statuses WHERE status = $1 LIMIT 1;