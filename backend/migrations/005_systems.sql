-- 005_systems.sql
-- Purpose: create monitored computer records that belong to a physical location.

CREATE TABLE IF NOT EXISTS systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    device_name VARCHAR(255) NOT NULL,
    operating_system VARCHAR(100) NOT NULL,
    os_version VARCHAR(100),
    system_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    monitor_status VARCHAR(32) NOT NULL DEFAULT 'offline',
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT systems_location_fk
        FOREIGN KEY (location_id)
        REFERENCES locations (id)
        ON DELETE CASCADE,

    CONSTRAINT systems_hostname_unique UNIQUE (hostname),
    CONSTRAINT systems_device_name_unique UNIQUE (device_name),
    CONSTRAINT systems_system_status_check CHECK (
        system_status IN ('pending', 'active', 'suspended', 'retired')
    ),
    CONSTRAINT systems_monitor_status_check CHECK (
        monitor_status IN ('online', 'offline', 'inactive', 'maintenance', 'suspended', 'retired')
    )
);

CREATE INDEX IF NOT EXISTS idx_systems_location_id
    ON systems (location_id);

CREATE INDEX IF NOT EXISTS idx_systems_hostname
    ON systems (hostname);

CREATE INDEX IF NOT EXISTS idx_systems_system_status
    ON systems (system_status);

CREATE INDEX IF NOT EXISTS idx_systems_monitor_status
    ON systems (monitor_status);
