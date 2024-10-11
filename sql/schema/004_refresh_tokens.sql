-- +goose Up
-- Create table with id, user's id as foreign key, refresh_token string, created_at, expires_at and revoked flag
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    refresh_token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Enable pg_cron extension for refresh token lifecycle management
CREATE EXTENSION IF NOT EXISTS "pg_cron";

-- Create a function that revokes refresh tokens after they expire
CREATE OR REPLACE FUNCTION revoke_expired_tokens() 
RETURNS VOID AS $$
BEGIN UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE expires_at < NOW() AND revoked_at IS NULL; END; $$ LANGUAGE plpgsql;

-- Schedule pg_cron to invalidate expired tokens every day
SELECT cron.schedule('0 0 * * *', 'SELECT revoke_expired_tokens()');

-- Create a function that deletes 3 day old expired tokens
CREATE OR REPLACE FUNCTION delete_old_revoked_tokens() 
RETURNS VOID AS $$
BEGIN DELETE FROM refresh_tokens WHERE revoked_at IS NOT NULL AND created_at < NOW() - INTERVAL '3 days'; END; $$ LANGUAGE plpgsql;

-- Schedule pg_cron to delete 3 day old expired tokens every day
SELECT cron.schedule('0 0 * * *', 'SELECT delete_old_revoked_tokens()');



-- +goose Down

-- Remove the cron jobs for token revocation and deletion
-- Fetch and unschedule the job for revoking expired tokens
-- +goose StatementBegin
DO $$
DECLARE
    revoke_jobid INT;
BEGIN
    -- Cast the job column to text for comparison using LIKE
    SELECT jobid INTO revoke_jobid FROM cron.job
        WHERE job::TEXT LIKE '%revoke_expired_tokens%';  
    IF revoke_jobid IS NOT NULL THEN
        PERFORM cron.unschedule(revoke_jobid);
    END IF;
END $$;
-- +goose StatementEnd

-- Fetch and unschedule the job for deleting old revoked tokens
-- +goose StatementBegin
DO $$
DECLARE
    delete_jobid INT;
BEGIN
    -- Cast the job column to text for comparison using LIKE
    SELECT jobid INTO delete_jobid FROM cron.job
        WHERE job::TEXT LIKE '%delete_old_revoked_tokens%';  
    IF delete_jobid IS NOT NULL THEN
        PERFORM cron.unschedule(delete_jobid);
    END IF;
END $$;
-- +goose StatementEnd

-- Drop the function that deletes 3 day old revoked tokens
DROP FUNCTION IF EXISTS delete_old_revoked_tokens();

-- Drop the function that revokes expired tokens
DROP FUNCTION IF EXISTS revoke_expired_tokens();

-- Drop the refresh_tokens table
DROP TABLE IF EXISTS refresh_tokens CASCADE;

-- Optionally, disable the pg_cron extension if it's not needed anymore
DROP EXTENSION IF EXISTS "pg_cron";