-- 008_roles.sql
-- Purpose: represent roles that can be granted to users across the platform.

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT roles_name_unique UNIQUE (name)
);

CREATE INDEX IF NOT EXISTS idx_roles_name
    ON roles (name);
