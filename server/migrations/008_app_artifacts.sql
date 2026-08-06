-- Secure artifact delivery for purchased apps.
-- A developer can attach one or more downloadable artifacts (bundles, source
-- archives) to an app. Downloads are served from disk only after the buyer's
-- entitlement is verified via a short-lived signed delivery token.

CREATE TABLE app_artifacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    file_name TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    size_bytes INTEGER NOT NULL DEFAULT 0,
    content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    checksum_sha256 TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE
);

CREATE INDEX idx_app_artifacts_app_id ON app_artifacts(app_id);