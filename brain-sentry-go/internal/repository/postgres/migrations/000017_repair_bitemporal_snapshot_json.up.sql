UPDATE memory_history
SET snapshot = (
    snapshot - ARRAY[
        'validation_status', 'embedding', 'source_type', 'source_reference',
        'created_by', 'tenant_id', 'created_at', 'updated_at',
        'last_accessed_at', 'access_count', 'injection_count', 'helpful_count',
        'not_helpful_count', 'code_example', 'programming_language',
        'memory_type', 'deleted_at', 'emotional_weight', 'sim_hash',
        'valid_from', 'valid_to', 'decay_rate', 'superseded_by', 'recorded_at'
    ]::TEXT[]
) || jsonb_build_object(
    'validationStatus', snapshot->'validation_status',
    'sourceType', snapshot->'source_type',
    'sourceReference', snapshot->'source_reference',
    'createdBy', snapshot->'created_by',
    'tenantId', snapshot->'tenant_id',
    'createdAt', snapshot->'created_at',
    'updatedAt', snapshot->'updated_at',
    'lastAccessedAt', snapshot->'last_accessed_at',
    'accessCount', snapshot->'access_count',
    'injectionCount', snapshot->'injection_count',
    'helpfulCount', snapshot->'helpful_count',
    'notHelpfulCount', snapshot->'not_helpful_count',
    'codeExample', snapshot->'code_example',
    'programmingLanguage', snapshot->'programming_language',
    'memoryType', snapshot->'memory_type',
    'deletedAt', snapshot->'deleted_at',
    'emotionalWeight', snapshot->'emotional_weight',
    'simHash', snapshot->'sim_hash',
    'validFrom', snapshot->'valid_from',
    'validTo', snapshot->'valid_to',
    'decayRate', snapshot->'decay_rate',
    'supersededBy', snapshot->'superseded_by',
    'recordedAt', snapshot->'recorded_at',
    'tags', to_jsonb(tags)
)
WHERE snapshot ? 'tenant_id';
