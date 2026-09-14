-- 006_agents.sql
-- Purpose: create agent identities that authenticate with the backend and are linked to a system.

CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL,
    agent_code VARCHAR(255) NOT NULL,
    agent_version VARCHAR(50) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    last_heartbeat_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT agents_system_fk
        FOREIGN KEY (system_id)
        REFERENCES systems (id)
        ON DELETE CASCADE,

    CONSTRAINT agents_agent_code_unique UNIQUE (agent_code),
    CONSTRAINT agents_status_check CHECK (
        status IN ('pending', 'approved', 'rejected', 'disabled')
    )
);

CREATE INDEX IF NOT EXISTS idx_agents_system_id
    ON agents (system_id);

CREATE INDEX IF NOT EXISTS idx_agents_status
    ON agents (status);

CREATE INDEX IF NOT EXISTS idx_agents_last_heartbeat_at
    ON agents (last_heartbeat_at);
