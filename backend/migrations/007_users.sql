-- 007_users.sql
-- Purpose: store institution user accounts and identity information for role-based access.

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255),
    role VARCHAR(32) NOT NULL DEFAULT 'operator',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_institution_fk
        FOREIGN KEY (institution_id)
        REFERENCES institutions (id)
        ON DELETE CASCADE,

    CONSTRAINT users_email_unique UNIQUE (institution_id, email),
    CONSTRAINT users_role_check CHECK (
        role IN ('admin', 'manager', 'operator', 'viewer')
    )
);

CREATE INDEX IF NOT EXISTS idx_users_institution_id
    ON users (institution_id);

CREATE INDEX IF NOT EXISTS idx_users_email
    ON users (email);

CREATE INDEX IF NOT EXISTS idx_users_role
    ON users (role);
