-- Semantica-inspired features: Decisions, Policies, Events, bi-temporal provenance

-- Decisions: first-class auditable reasoning records
CREATE TABLE IF NOT EXISTS decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(36) NOT NULL,
    category VARCHAR(120) NOT NULL,
    scenario TEXT NOT NULL,
    reasoning TEXT NOT NULL,
    outcome VARCHAR(40) NOT NULL DEFAULT 'pending',
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    agent_id VARCHAR(120) NOT NULL DEFAULT '',
    session_id VARCHAR(120) NOT NULL DEFAULT '',
    parent_decision_id UUID,
    entity_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    memory_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    policy_violations JSONB NOT NULL DEFAULT '[]'::jsonb,
    embedding vector(1536),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    superseded_by UUID
);

CREATE INDEX IF NOT EXISTS idx_decisions_tenant_category
    ON decisions (tenant_id, category, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_decisions_parent
    ON decisions (parent_decision_id)
    WHERE parent_decision_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_decisions_valid
    ON decisions (tenant_id, valid_from, valid_until);
CREATE INDEX IF NOT EXISTS idx_decisions_entity_ids
    ON decisions USING GIN (entity_ids);
CREATE INDEX IF NOT EXISTS idx_decisions_memory_ids
    ON decisions USING GIN (memory_ids);

-- Policies: rules that can be enforced against decisions
CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(120) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    rule_type VARCHAR(40) NOT NULL,
    rule_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_policies_tenant_category
    ON policies (tenant_id, category, enabled);

-- Events: typed occurrences with participants (distinct from Memory/Decision)
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(120) NOT NULL,
    title VARCHAR(500) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    participants JSONB NOT NULL DEFAULT '[]'::jsonb,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_memory_id VARCHAR(36),
    embedding vector(1536),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_tenant_type_time
    ON events (tenant_id, event_type, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_participants
    ON events USING GIN (participants);
CREATE INDEX IF NOT EXISTS idx_events_source_memory
    ON events (source_memory_id)
    WHERE source_memory_id IS NOT NULL;

-- Bi-temporal provenance on Memory: add recorded_at (when system learned the fact)
ALTER TABLE memories ADD COLUMN IF NOT EXISTS recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_memories_as_of
    ON memories (tenant_id, valid_from, valid_to, recorded_at);
