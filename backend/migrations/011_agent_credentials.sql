-- 011_agent_credentials.sql
-- Purpose: store one-way hashes for agent authentication credentials.

ALTER TABLE agents
    ADD COLUMN IF NOT EXISTS credential_hash VARCHAR(255);