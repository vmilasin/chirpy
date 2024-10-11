-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email;

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