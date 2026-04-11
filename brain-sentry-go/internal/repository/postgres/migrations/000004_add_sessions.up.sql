CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL DEFAULT '',
    tenant_id VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    memory_count INT NOT NULL DEFAULT 0,
    interception_count INT NOT NULL DEFAULT 0,
    note_count INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_sessions_tenant_status ON sessions (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions (expires_at) WHERE status = 'ACTIVE';
