ALTER TABLE users ADD COLUMN verified_at DATETIME;
ALTER TABLE users ADD COLUMN avatar_url TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN display_name TEXT NOT NULL DEFAULT '';

CREATE TABLE user_verifications (
    token TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_verifications_user_id ON user_verifications(user_id);
