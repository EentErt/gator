-- name: GetFeed :one
SELECT * FROM feed WHERE name = $1;