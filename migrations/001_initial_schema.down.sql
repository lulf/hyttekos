-- Drop the cleanup function
DROP FUNCTION IF EXISTS cleanup_expired_sessions();

-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_temperature_timestamp;

-- Drop tables
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS temperature_readings;