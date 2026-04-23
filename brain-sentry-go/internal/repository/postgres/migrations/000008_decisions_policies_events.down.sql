DROP INDEX IF EXISTS idx_memories_as_of;
ALTER TABLE memories DROP COLUMN IF EXISTS recorded_at;

DROP INDEX IF EXISTS idx_events_source_memory;
DROP INDEX IF EXISTS idx_events_participants;
DROP INDEX IF EXISTS idx_events_tenant_type_time;
DROP TABLE IF EXISTS events;

DROP INDEX IF EXISTS idx_policies_tenant_category;
DROP TABLE IF EXISTS policies;

DROP INDEX IF EXISTS idx_decisions_memory_ids;
DROP INDEX IF EXISTS idx_decisions_entity_ids;
DROP INDEX IF EXISTS idx_decisions_valid;
DROP INDEX IF EXISTS idx_decisions_parent;
DROP INDEX IF EXISTS idx_decisions_tenant_category;
DROP TABLE IF EXISTS decisions;
