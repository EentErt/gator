-- name: GetFeedFollowsForUser :many
SELECT feed.name AS feed_name, users.name AS user_name
FROM feed_follow
INNER JOIN users
ON feed_follow.user_id = users.id 
INNER JOIN feed
ON feed_follow.feed_id = feed.id
WHERE users.name = $1;