ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

CREATE TABLE developer_applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    display_name TEXT NOT NULL,
    profile_url TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    review_note TEXT NOT NULL DEFAULT '',
    reviewed_by INTEGER,
    reviewed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_developer_applications_pending_user
    ON developer_applications(user_id)
    WHERE status = 'pending';
CREATE INDEX idx_developer_applications_status_created
    ON developer_applications(status, created_at);
