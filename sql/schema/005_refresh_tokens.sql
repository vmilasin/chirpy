-- +goose Up

ALTER TABLE refresh_tokens
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Create trigger on refresh_tokens table to invoke the trigger function before updates
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON refresh_tokens
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- -goose Down

-- Drop the trigger
DROP TRIGGER IF EXISTS set_updated_at ON refresh_tokens;

ALTER TABLE refresh_tokens
DROP COLUMN updated_at;