-- +goose Up
-- Create table with id, user's id as foreign key, body, created_at, and updated_at
CREATE TABLE chirps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create trigger on chirps table to invoke the trigger function before updates
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON chirps
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();




-- +goose Down
-- Drop the trigger
DROP TRIGGER IF EXISTS set_updated_at ON chirps;

-- Drop the table
DROP TABLE IF EXISTS chirps;