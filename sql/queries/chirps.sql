-- name: CreateChirp :one
INSERT INTO chirps (user_id, body)
VALUES ($1, $2)
RETURNING id, body;

-- name: GetChirpAll :one
SELECT *
FROM chirps;

-- name: GetChirpByID :one
SELECT *
FROM chirps
WHERE id = $1;

-- name: GetChirpFromUser :one
SELECT *
FROM chirps
WHERE user_id = $1;