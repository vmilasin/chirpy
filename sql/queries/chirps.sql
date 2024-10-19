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
FROM chirps
ORDER BY 
    CASE WHEN $1 THEN created_at END DESC,
    CASE WHEN NOT $1 THEN created_at END ASC;

-- name: GetChirpsFromAuthor :many
SELECT
    id AS "id", --json:"id"
    body AS "body", --json:"body"
    user_id AS "user_id", --json:"user_id"
    created_at AS "created_at", --json:"created_at"
    updated_at AS "updated_at" --json:"updated_at"
FROM chirps
WHERE user_id = $1
ORDER BY 
    CASE WHEN $2 THEN created_at END DESC,
    CASE WHEN NOT $2 THEN created_at END ASC;


-- name: GetChirpByID :one
SELECT *
FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps 
WHERE id = $1;