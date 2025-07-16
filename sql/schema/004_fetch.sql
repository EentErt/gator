-- +goose Up
ALTER TABLE feed ADD last_fetched_at TIMESTAMP;

-- +goose Down
ALTER TABLE feed DROP last_fetched_at;