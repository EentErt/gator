-- +goose Up
CREATE TABLE feed_follow(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
	updated_at TIMESTAMP,
    user_id UUID,
    feed_id INTEGER,
    CONSTRAINT fk_user_id
    FOREIGN KEY (user_id)
    REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_feed_id
    FOREIGN KEY (feed_id)
    REFERENCES feed(id) ON DELETE CASCADE,
    UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follow;