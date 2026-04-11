-- F-11: Emotional weight field (-1 to +1)
ALTER TABLE memories ADD COLUMN IF NOT EXISTS emotional_weight DOUBLE PRECISION NOT NULL DEFAULT 0;

-- F-13: SimHash fingerprint for near-duplicate detection
ALTER TABLE memories ADD COLUMN IF NOT EXISTS sim_hash VARCHAR(16) NOT NULL DEFAULT '';

-- Index for SimHash lookups
CREATE INDEX IF NOT EXISTS idx_memories_sim_hash ON memories (tenant_id, sim_hash) WHERE sim_hash != '' AND deleted_at IS NULL;

-- F-35: Session observations table
CREATE TABLE IF NOT EXISTS session_observations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    context TEXT NOT NULL DEFAULT '',
    related_memory_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    auto_generated BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_session_observations_tenant ON session_observations (tenant_id);
CREATE INDEX IF NOT EXISTS idx_session_observations_session ON session_observations (session_id);
CREATE INDEX IF NOT EXISTS idx_session_observations_type ON session_observations (tenant_id, type);
