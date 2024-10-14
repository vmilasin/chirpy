-- name: CreateChirp :one
INSERT INTO chirps (user_id, body)
VALUES ($1, $2)
RETURNING *;

-- name: GetChirpAll :many
SELECT
    id AS "id", --json:"id"
    body AS "body", --json:"body"
    user_id AS "user_id", --json:"user_id"
    created_at AS "created_at", --json:"created_at"
    updated_at AS "updated_at" --json:"updated_at"
FROM chirps;

-- name: GetChirpByID :one
SELECT *
FROM chirps
WHERE id = $1;