-- 015_issues.sql
-- Purpose: track operational issues and their resolution history.

CREATE TABLE IF NOT EXISTS issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    priority VARCHAR(16) NOT NULL DEFAULT 'medium',
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    reported_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT issues_system_fk FOREIGN KEY (system_id)
        REFERENCES systems (id) ON DELETE CASCADE,
    CONSTRAINT issues_priority_check CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT issues_status_check CHECK (status IN ('open', 'acknowledged', 'in_progress', 'resolved', 'closed'))
);

CREATE TABLE IF NOT EXISTS issue_updates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL,
    status VARCHAR(32),
    comment TEXT NOT NULL,
    updated_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT issue_updates_issue_fk FOREIGN KEY (issue_id)
        REFERENCES issues (id) ON DELETE CASCADE,
    CONSTRAINT issue_updates_status_check CHECK (status IS NULL OR status IN ('open', 'acknowledged', 'in_progress', 'resolved', 'closed'))
);

CREATE INDEX IF NOT EXISTS idx_issues_system_status ON issues (system_id, status);
CREATE INDEX IF NOT EXISTS idx_issue_updates_issue_created ON issue_updates (issue_id, created_at DESC);