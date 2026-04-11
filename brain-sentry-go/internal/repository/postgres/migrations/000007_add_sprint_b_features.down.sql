DROP INDEX IF EXISTS idx_memories_active;
DROP INDEX IF EXISTS idx_memories_superseded_by;
DROP INDEX IF EXISTS idx_memories_valid_to;
ALTER TABLE memories DROP COLUMN IF EXISTS superseded_by;
ALTER TABLE memories DROP COLUMN IF EXISTS decay_rate;
ALTER TABLE memories DROP COLUMN IF EXISTS valid_to;
ALTER TABLE memories DROP COLUMN IF EXISTS valid_from;
