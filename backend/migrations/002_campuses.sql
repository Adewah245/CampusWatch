-- 002_campuses.sql
-- Purpose: create campus records under each institution so physical locations can be grouped correctly.

CREATE TABLE IF NOT EXISTS campuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT campuses_institution_fk
        FOREIGN KEY (institution_id)
        REFERENCES institutions (id)
        ON DELETE CASCADE,

    CONSTRAINT campuses_name_unique UNIQUE (institution_id, name),
    CONSTRAINT campuses_slug_unique UNIQUE (institution_id, slug),
    CONSTRAINT campuses_status_check CHECK (
        status IN ('active', 'inactive', 'suspended')
    )
);

CREATE INDEX IF NOT EXISTS idx_campuses_institution_id
    ON campuses (institution_id);

CREATE INDEX IF NOT EXISTS idx_campuses_name
    ON campuses (name);

CREATE INDEX IF NOT EXISTS idx_campuses_status
    ON campuses (status);
