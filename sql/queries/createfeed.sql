-- name: CreateFeed :exec
INSERT INTO feed(created_at, updated_at, name, url, user_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;