-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, refresh_token, expires_at)
VALUES ($1, $2, $3)
RETURNING refresh_token;

-- name: GetRefreshTokenForUser :one
SELECT *
FROM refresh_tokens
WHERE user_id = $1 and revoked = FALSE;