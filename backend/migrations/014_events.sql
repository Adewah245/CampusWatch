-- 014_events.sql
-- Purpose: record operational events emitted by agents and backend workflows.

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL,
    agent_id UUID,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT events_system_fk FOREIGN KEY (system_id)
        REFERENCES systems (id) ON DELETE CASCADE,
    CONSTRAINT events_agent_fk FOREIGN KEY (agent_id)
        REFERENCES agents (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_events_system_occurred
    ON events (system_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_type
    ON events (event_type);