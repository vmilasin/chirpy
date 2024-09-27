-- +goose Up
-- Create table with id, user's id as foreign key, refresh_token string, created_at, expires_at and revoked flag
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    refresh_token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Enable pg_cron extension for refresh token lifecycle management
CREATE EXTENSION IF NOT EXISTS "pg_cron";

-- Create a function that revokes refresh tokens after they expire
CREATE OR REPLACE FUNCTION revoke_expired_tokens() 
RETURNS VOID AS $$
BEGIN UPDATE refresh_tokens SET revoked = TRUE WHERE expires_at < NOW() AND revoked = FALSE; END; $$ LANGUAGE plpgsql;

-- Schedule pg_cron to invalidate expired tokens every day
SELECT cron.schedule('0 0 * * *', 'SELECT revoke_expired_tokens()');

-- Create a function that deletes 3 day old expired tokens
CREATE OR REPLACE FUNCTION delete_old_revoked_tokens() 
RETURNS VOID AS $$
BEGIN DELETE FROM refresh_tokens WHERE revoked = TRUE AND created_at < NOW() - INTERVAL '3 days'; END; $$ LANGUAGE plpgsql;

-- Schedule pg_cron to delete 3 day old expired tokens every day
SELECT cron.schedule('0 0 * * *', 'SELECT delete_old_revoked_tokens()');



-- +goose Down
-- Remove the cron jobs for token revocation and deletion
SELECT cron.unschedule((SELECT jobid FROM cron.job WHERE job = 'SELECT revoke_expired_tokens()'));
SELECT cron.unschedule((SELECT jobid FROM cron.job WHERE job = 'SELECT delete_old_revoked_tokens()'));

-- Drop the function that deletes 3 day old revoked tokens
DROP FUNCTION IF EXISTS delete_old_revoked_tokens();

-- Drop the function that revokes expired tokens
DROP FUNCTION IF EXISTS revoke_expired_tokens();

-- Drop the refresh_tokens table
DROP TABLE IF EXISTS refresh_tokens CASCADE;

-- Optionally, disable the pg_cron extension if it's not needed anymore
DROP EXTENSION IF EXISTS "pg_cron";