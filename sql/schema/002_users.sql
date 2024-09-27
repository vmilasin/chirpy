-- +goose Up

ALTER TABLE users
ADD COLUMN password_hash BYTEA NOT NULL;

-- +goose Down

ALTER TABLE users
DROP COLUMN password_hash;