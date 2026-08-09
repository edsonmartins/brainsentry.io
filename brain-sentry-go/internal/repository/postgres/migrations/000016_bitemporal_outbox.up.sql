CREATE TABLE IF NOT EXISTS memory_history (
    id BIGSERIAL PRIMARY KEY,
    memory_id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    system_from TIMESTAMPTZ NOT NULL,
    system_to TIMESTAMPTZ,
    valid_from TIMESTAMPTZ,
    valid_to TIMESTAMPTZ,
    operation VARCHAR(20) NOT NULL,
    snapshot JSONB NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_memory_history_current
    ON memory_history(memory_id, tenant_id) WHERE system_to IS NULL;
CREATE INDEX IF NOT EXISTS idx_memory_history_as_of
    ON memory_history(tenant_id, system_from, system_to);

INSERT INTO memory_history (
    memory_id, tenant_id, system_from, system_to, valid_from, valid_to,
    operation, snapshot, tags
)
SELECT m.id, m.tenant_id, COALESCE(m.recorded_at, m.created_at), NULL,
       m.valid_from, m.valid_to,
       CASE WHEN m.deleted_at IS NULL THEN 'create' ELSE 'delete' END,
       jsonb_build_object(
           'id', m.id,
           'content', m.content,
           'summary', m.summary,
           'category', m.category,
           'importance', m.importance,
           'validationStatus', m.validation_status,
           'metadata', m.metadata,
           'tags', to_jsonb(COALESCE(array_agg(mt.tag) FILTER (WHERE mt.tag IS NOT NULL), '{}')),
           'sourceType', m.source_type,
           'sourceReference', m.source_reference,
           'createdBy', m.created_by,
           'tenantId', m.tenant_id,
           'createdAt', m.created_at,
           'updatedAt', m.updated_at,
           'lastAccessedAt', m.last_accessed_at,
           'version', m.version,
           'accessCount', m.access_count,
           'injectionCount', m.injection_count,
           'helpfulCount', m.helpful_count,
           'notHelpfulCount', m.not_helpful_count,
           'codeExample', m.code_example,
           'programmingLanguage', m.programming_language,
           'memoryType', m.memory_type,
           'deletedAt', m.deleted_at,
           'emotionalWeight', m.emotional_weight,
           'simHash', m.sim_hash,
           'validFrom', m.valid_from,
           'validTo', m.valid_to,
           'decayRate', m.decay_rate,
           'supersededBy', m.superseded_by,
           'recordedAt', m.recorded_at,
           'provenance', m.provenance
       ),
       COALESCE(array_agg(mt.tag) FILTER (WHERE mt.tag IS NOT NULL), '{}')
FROM memories m
LEFT JOIN memory_tags mt ON mt.memory_id = m.id
WHERE NOT EXISTS (
    SELECT 1 FROM memory_history h
    WHERE h.memory_id = m.id AND h.tenant_id = m.tenant_id
)
GROUP BY m.id;

CREATE TABLE IF NOT EXISTS projection_outbox (
    id VARCHAR(100) PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    operation VARCHAR(20) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    last_error TEXT
);

CREATE INDEX IF NOT EXISTS idx_projection_outbox_pending
    ON projection_outbox(status, available_at, created_at);
CREATE INDEX IF NOT EXISTS idx_projection_outbox_aggregate
    ON projection_outbox(tenant_id, aggregate_type, aggregate_id);
