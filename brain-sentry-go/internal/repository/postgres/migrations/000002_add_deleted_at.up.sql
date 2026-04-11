-- Add soft delete support to memories table
ALTER TABLE memories ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Index for efficient filtering of non-deleted records
CREATE INDEX IF NOT EXISTS idx_memories_deleted_at ON memories (deleted_at) WHERE deleted_at IS NULL;
