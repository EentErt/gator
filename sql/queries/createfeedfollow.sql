-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follow(created_at, updated_at, user_id, feed_id) 
    VALUES($1, $2, $3, $4)
    RETURNING *
)
SELECT inserted_feed_follow.*, users.name AS user_name, feed.name AS feed_name
FROM inserted_feed_follow
INNER JOIN users ON inserted_feed_follow.user_id = users.id
INNER JOIN feed ON inserted_feed_follow.feed_id = feed.id;