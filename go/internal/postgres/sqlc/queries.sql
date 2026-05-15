-- name: ListVideos :many
SELECT * FROM videos;

-- name: FindVideoById :one
SELECT * FROM videos WHERE id = $1;

-- name: CreateVideo :exec
INSERT INTO videos (status, file_path) VALUES ($1, $2);

-- name: FindStatusIdByName :one
SELECT id FROM video_statuses WHERE status = $1 LIMIT 1;