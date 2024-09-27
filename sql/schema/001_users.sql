-- +goose Up

-- Enable pgcrypto extension to use gen_random_uuid
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create table with id, created_at, updated_at, and email
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email TEXT NOT NULL UNIQUE
);

-- Create trigger function to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = CURRENT_TIMESTAMP; RETURN NEW; END; $$ LANGUAGE plpgsql;

-- Create trigger on users table to invoke the trigger function before updates
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();






-- +goose Down

-- Drop the trigger
DROP TRIGGER IF EXISTS set_updated_at ON users;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Optionally drop the extension (if not used elsewhere)
DROP EXTENSION IF EXISTS pgcrypto;

-- Drop the table
DROP TABLE IF EXISTS users;
