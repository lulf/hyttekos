-- Temperature readings table with time-series optimization
CREATE TABLE temperature_readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    temperature DECIMAL(5,2) NOT NULL,
    timestamp DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient time-based queries
CREATE INDEX idx_temperature_timestamp ON temperature_readings(timestamp DESC);

-- Sessions table for OAuth authentication
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for session cleanup
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);