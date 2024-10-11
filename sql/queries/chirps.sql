-- name: CreateChirp :one
INSERT INTO chirps (user_id, body)
VALUES ($1, $2)
RETURNING id, body;

-- name: GetChirpAll :many
SELECT *
FROM chirps;

-- name: GetChirpByID :one
SELECT *
FROM chirps
WHERE id = $1;