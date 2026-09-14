-- 004_locations.sql
-- Purpose: create table/location records that represent physical spaces inside a cluster.

CREATE TABLE IF NOT EXISTS locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    location_type VARCHAR(50) NOT NULL DEFAULT 'table',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT locations_cluster_fk
        FOREIGN KEY (cluster_id)
        REFERENCES clusters (id)
        ON DELETE CASCADE,

    CONSTRAINT locations_name_unique UNIQUE (cluster_id, name),
    CONSTRAINT locations_slug_unique UNIQUE (cluster_id, slug),
    CONSTRAINT locations_status_check CHECK (
        status IN ('active', 'inactive', 'suspended')
    ),
    CONSTRAINT locations_type_check CHECK (
        location_type IN ('table', 'room', 'lab', 'office')
    )
);

CREATE INDEX IF NOT EXISTS idx_locations_cluster_id
    ON locations (cluster_id);

CREATE INDEX IF NOT EXISTS idx_locations_name
    ON locations (name);

CREATE INDEX IF NOT EXISTS idx_locations_status
    ON locations (status);
