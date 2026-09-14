-- 013_user_sessions.sql
-- Purpose: track human login/logout activity on monitored systems.

CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL,
    username VARCHAR(255) NOT NULL,
    state VARCHAR(32) NOT NULL DEFAULT 'logged_in_active',
    login_at TIMESTAMPTZ NOT NULL,
    logout_at TIMESTAMPTZ,
    last_activity_at TIMESTAMPTZ NOT NULL,
    duration_seconds BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT user_sessions_system_fk FOREIGN KEY (system_id)
        REFERENCES systems (id) ON DELETE CASCADE,
    CONSTRAINT user_sessions_state_check CHECK (state IN ('logged_out', 'logged_in_active', 'logged_in_idle')),
    CONSTRAINT user_sessions_duration_check CHECK (duration_seconds IS NULL OR duration_seconds >= 0)
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_system_login
    ON user_sessions (system_id, login_at DESC);