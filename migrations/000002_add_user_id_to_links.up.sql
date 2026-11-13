ALTER TABLE links ADD COLUMN user_id VARCHAR(32);
CREATE INDEX idx_user_id ON links(user_id);