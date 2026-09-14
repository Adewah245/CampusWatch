-- 012_system_health.sql
-- Purpose: store point-in-time health reports from monitored systems.

CREATE TABLE IF NOT EXISTS system_health_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    cpu_percent DOUBLE PRECISION NOT NULL,
    memory_percent DOUBLE PRECISION NOT NULL,
    disk_percent DOUBLE PRECISION NOT NULL,
    battery_percent DOUBLE PRECISION,
    network_connected BOOLEAN NOT NULL,
    operating_system VARCHAR(100),
    agent_health VARCHAR(32) NOT NULL,
    uptime_seconds BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT system_health_system_fk FOREIGN KEY (system_id)
        REFERENCES systems (id) ON DELETE CASCADE,
    CONSTRAINT system_health_cpu_check CHECK (cpu_percent >= 0 AND cpu_percent <= 100),
    CONSTRAINT system_health_memory_check CHECK (memory_percent >= 0 AND memory_percent <= 100),
    CONSTRAINT system_health_disk_check CHECK (disk_percent >= 0 AND disk_percent <= 100),
    CONSTRAINT system_health_battery_check CHECK (battery_percent IS NULL OR (battery_percent >= 0 AND battery_percent <= 100)),
    CONSTRAINT system_health_agent_check CHECK (agent_health IN ('healthy', 'warning', 'unhealthy')),
    CONSTRAINT system_health_uptime_check CHECK (uptime_seconds >= 0)
);

CREATE INDEX IF NOT EXISTS idx_system_health_system_recorded
    ON system_health_snapshots (system_id, recorded_at DESC);