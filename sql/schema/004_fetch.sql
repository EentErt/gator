-- +goose Up
ALTER TABLE feed ADD last_fetched_at;