-- name: CreateFeedFollow :many
WITH inserted_feed_follow AS (
    INSERT INTO feed_follow(created_at, updated_at, user_id, feed_id) 
    VALUES($1, $2, $3, $4)
    RETURNING *
)
SELECT inserted_feed_follow.*, users.name AS user_name, feed.name AS feed_name
FROM inserted_feed_follow
INNER JOIN users ON user_id
INNER JOIN feed ON feed_id;