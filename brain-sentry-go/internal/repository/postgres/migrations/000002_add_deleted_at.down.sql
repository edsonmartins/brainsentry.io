DROP INDEX IF EXISTS idx_memories_deleted_at;
ALTER TABLE memories DROP COLUMN IF EXISTS deleted_at;
