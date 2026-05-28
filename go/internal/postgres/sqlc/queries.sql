-- name: ListVideos :many
SELECT * FROM videos;

-- name: FindVideoById :one
SELECT * FROM videos WHERE id = $1;

-- name: CreateVideo :one
INSERT INTO videos (status, file_extension) VALUES ($1, $2) RETURNING id;

-- name: VideoHasValidStatusToStartProcessingJob :one
SELECT COUNT(*) > 0
FROM videos video
JOIN video_statuses vstatus ON vstatus.id = video.status
WHERE video.id = $1 AND vstatus.status = 'pending';

-- name: UpdateVideoStatus :exec
UPDATE videos SET status = $2 WHERE id = $1;

-- name: UpdateVideoProgress :exec
UPDATE videos SET progress = $2 WHERE id = $1;