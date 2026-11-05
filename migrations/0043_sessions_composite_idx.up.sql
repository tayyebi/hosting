-- 0043: composite index on sessions(user_id, expires_at DESC) for efficient
-- per-user active-session lookups and cleanup queries.
CREATE INDEX IF NOT EXISTS sessions_user_expires_idx
    ON sessions (user_id, expires_at DESC)
    WHERE revoked_at IS NULL;
