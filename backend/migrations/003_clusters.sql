-- 003_clusters.sql
-- Purpose: create cluster records to group systems within a campus.

CREATE TABLE IF NOT EXISTS clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campus_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT clusters_campus_fk
        FOREIGN KEY (campus_id)
        REFERENCES campuses (id)
        ON DELETE CASCADE,

    CONSTRAINT clusters_name_unique UNIQUE (campus_id, name),
    CONSTRAINT clusters_slug_unique UNIQUE (campus_id, slug),
    CONSTRAINT clusters_status_check CHECK (
        status IN ('active', 'inactive', 'suspended')
    )
);

CREATE INDEX IF NOT EXISTS idx_clusters_campus_id
    ON clusters (campus_id);

CREATE INDEX IF NOT EXISTS idx_clusters_name
    ON clusters (name);

CREATE INDEX IF NOT EXISTS idx_clusters_status
    ON clusters (status);
