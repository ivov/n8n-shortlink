CREATE TABLE IF NOT EXISTS shortlinks (
	slug TEXT PRIMARY KEY,
	kind TEXT NOT NULL CHECK (kind IN ('workflow', 'url')),
	content TEXT NOT NULL,
	created_at TEXT DEFAULT CURRENT_TIMESTAMP,
	creator_ip TEXT DEFAULT 'unknown',
	expires_at TEXT, 
	password TEXT CHECK (LENGTH(password) = 0 OR LENGTH(password) >= 8),
	allowed_visits INTEGER DEFAULT -1 CHECK (allowed_visits = -1 OR allowed_visits > 0)
) STRICT;