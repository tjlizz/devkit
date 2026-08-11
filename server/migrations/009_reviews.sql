-- Buyer reviews and ratings for marketplace apps.
-- Only verified buyers with an active entitlement may review an app. Each
-- buyer may leave at most one review per app (upsert on update). Reviews are
-- hidden (soft-delete) rather than physically removed so moderation history is
-- preserved.

CREATE TABLE app_reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    buyer_id INTEGER NOT NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'hidden')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    FOREIGN KEY (buyer_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (app_id, buyer_id)
);

CREATE INDEX idx_app_reviews_app_id ON app_reviews(app_id);
CREATE INDEX idx_app_reviews_buyer_id ON app_reviews(buyer_id);