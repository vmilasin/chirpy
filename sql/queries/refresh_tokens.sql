-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, refresh_token, expires_at)
VALUES ($1, $2, $3)
RETURNING refresh_token;

-- name: GetRefreshTokenForUser :one
SELECT *
FROM refresh_tokens
WHERE user_id = $1 and revoked_at IS NULL;

-- name: CheckRefreshTokenValidity :one
SELECT revoked_at, user_id
FROM refresh_tokens
where refresh_token =$1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
    revoked_at = CURRENT_TIMESTAMP
WHERE refresh_token =$1;