DROP TABLE IF EXISTS session_observations;
DROP INDEX IF EXISTS idx_memories_sim_hash;
ALTER TABLE memories DROP COLUMN IF EXISTS sim_hash;
ALTER TABLE memories DROP COLUMN IF EXISTS emotional_weight;
