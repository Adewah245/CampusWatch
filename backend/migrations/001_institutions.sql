-- 001_institutions.sql
-- Purpose: create the foundational institution table used to isolate tenant data.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT institutions_name_unique UNIQUE (name),
    CONSTRAINT institutions_slug_unique UNIQUE (slug),
    CONSTRAINT institutions_status_check CHECK (
        status IN ('active', 'inactive', 'suspended')
    )
);

CREATE INDEX IF NOT EXISTS idx_institutions_name
    ON institutions (name);

CREATE INDEX IF NOT EXISTS idx_institutions_slug
    ON institutions (slug);

CREATE INDEX IF NOT EXISTS idx_institutions_status
    ON institutions (status);
