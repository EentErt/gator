-- name: GetFeedByUrl :one
SELECT * FROM feed WHERE url = $1;