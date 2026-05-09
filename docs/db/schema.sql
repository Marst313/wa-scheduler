CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    recipient_numbers TEXT NOT NULL,
    scheduled_sending_at INTEGER,
    sent_at INTEGER,
    retried_count INTEGER DEFAULT 0,
    status TEXT,
    reason TEXT DEFAULT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);