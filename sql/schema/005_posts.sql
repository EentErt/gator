-- +goose Up
CREATE TABLE posts(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT,
    published_at TIMESTAMP,
    feed_id INTEGER NOT NULL,
    CONSTRAINT fk_feed_id
    FOREIGN KEY (feed_id)
    REFERENCES feed(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;