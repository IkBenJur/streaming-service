-- name: ListVideos :many
SELECT * FROM videos;

-- name: FindVideoById :one
SELECT * FROM videos WHERE id = $1;