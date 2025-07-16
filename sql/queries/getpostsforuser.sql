-- name: GetPostsForUser :many
WITH get_feed_id (feed_id) AS (
    SELECT feed_id
    FROM feed_follow
    WHERE user_id = $1
)
SELECT *
FROM posts
INNER JOIN get_feed_id
ON posts.feed_id = get_feed_id.feed_id
ORDER BY posts.published_at DESC
LIMIT $2;