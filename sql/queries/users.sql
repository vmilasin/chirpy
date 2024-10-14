-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, password_hash
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id
FROM users
WHERE email = $1;

-- name: GetPWHash :one
SELECT password_hash
FROM users
WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
SET
    email = COALESCE(NULLIF($1, ''), email),  -- If email is provided, update it, otherwise keep the existing value
    password_hash = COALESCE(NULLIF($2, ''), password_hash)  -- If password is provided, update it, otherwise keep the existing value
WHERE id = $3
RETURNING id, email;