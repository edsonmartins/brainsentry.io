-- Sprint B: Temporal decay, supersession, and hybrid scoring fields

-- F-10: Memory type classification column
ALTER TABLE memories ADD COLUMN IF NOT EXISTS memory_type VARCHAR(50) NOT NULL DEFAULT 'semantic';

-- F-03/F-04: Temporal decay and supersession fields
ALTER TABLE memories ADD COLUMN IF NOT EXISTS valid_from TIMESTAMPTZ;
ALTER TABLE memories ADD COLUMN IF NOT EXISTS valid_to TIMESTAMPTZ;
ALTER TABLE memories ADD COLUMN IF NOT EXISTS decay_rate DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE memories ADD COLUMN IF NOT EXISTS superseded_by VARCHAR(36) NOT NULL DEFAULT '';

-- Index for efficient expiration queries
CREATE INDEX IF NOT EXISTS idx_memories_valid_to ON memories (tenant_id, valid_to)
    WHERE valid_to IS NOT NULL AND deleted_at IS NULL;

-- Index for supersession chain lookups
CREATE INDEX IF NOT EXISTS idx_memories_superseded_by ON memories (tenant_id, superseded_by)
    WHERE superseded_by != '' AND deleted_at IS NULL;

-- Index for active memories (not superseded, not expired)
CREATE INDEX IF NOT EXISTS idx_memories_active ON memories (tenant_id, memory_type, created_at DESC)
    WHERE deleted_at IS NULL AND superseded_by = '';
