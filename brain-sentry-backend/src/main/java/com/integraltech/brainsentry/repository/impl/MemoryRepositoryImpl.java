package com.integraltech.brainsentry.repository.impl;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.repository.MemoryRepository;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.autoconfigure.condition.ConditionalOnBean;
import org.springframework.stereotype.Repository;
import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;
import redis.clients.jedis.exceptions.JedisException;

import java.time.Instant;
import java.util.*;

/**
 * Jedis implementation of MemoryRepository.
 *
 * Uses Redis for storage with placeholder for graph operations.
 * Full graph support will be added when FalkorDB is properly configured.
 */
@Repository
@ConditionalOnBean(JedisPool.class)
@Slf4j
public class MemoryRepositoryImpl implements MemoryRepository {

    private final JedisPool jedisPool;
    private final ObjectMapper objectMapper;

    private static final String MEMORY_PREFIX = "memory:";
    private static final String TAG_INDEX = "tag_idx:";
    private static final String TENANT_MEMORIES = "tenant_memories:";

    public MemoryRepositoryImpl(JedisPool jedisPool, ObjectMapper objectMapper) {
        this.jedisPool = jedisPool;
        this.objectMapper = objectMapper;
    }

    @Override
    public Memory save(Memory memory) {
        try (Jedis jedis = jedisPool.getResource()) {
            // Generate ID if not present
            if (memory.getId() == null) {
                memory.setId(generateId());
            }

            // Set timestamps
            if (memory.getCreatedAt() == null) {
                memory.setCreatedAt(Instant.now());
            }
            memory.setUpdatedAt(Instant.now());

            // Serialize and store memory
            String key = MEMORY_PREFIX + memory.getId();
            String json = objectMapper.writeValueAsString(memory);
            jedis.set(key, json);

            // Add to tenant index
            jedis.sadd(TENANT_MEMORIES + memory.getTenantId(), memory.getId());

            // Index tags
            if (memory.getTags() != null) {
                for (String tag : memory.getTags()) {
                    jedis.sadd(TAG_INDEX + tag, memory.getId());
                }
            }

            log.debug("Saved memory: {}", memory.getId());
            return memory;
        } catch (Exception e) {
            log.error("Error saving memory: {}", memory.getId(), e);
            throw new RuntimeException("Failed to save memory", e);
        }
    }

    @Override
    public Optional<Memory> findById(String id) {
        try (Jedis jedis = jedisPool.getResource()) {
            String key = MEMORY_PREFIX + id;
            String json = jedis.get(key);

            if (json == null) {
                return Optional.empty();
            }

            Memory memory = objectMapper.readValue(json, Memory.class);
            return Optional.of(memory);
        } catch (Exception e) {
            log.error("Error finding memory by id: {}", id, e);
            return Optional.empty();
        }
    }

    @Override
    public List<Memory> findByTenantId(String tenantId) {
        try (Jedis jedis = jedisPool.getResource()) {
            Set<String> memoryIds = jedis.smembers(TENANT_MEMORIES + tenantId);
            List<Memory> memories = new ArrayList<>();

            for (String id : memoryIds) {
                findById(id).ifPresent(memories::add);
            }

            return memories;
        } catch (Exception e) {
            log.error("Error finding memories by tenant: {}", tenantId, e);
            return List.of();
        }
    }

    @Override
    public List<Memory> findByCategory(String category, String tenantId) {
        try (Jedis jedis = jedisPool.getResource()) {
            // Get all memories for tenant and filter by category
            return findByTenantId(tenantId).stream()
                .filter(m -> category.equals(m.getCategory().name()))
                .toList();
        } catch (Exception e) {
            log.error("Error finding memories by category: {}", category, e);
            return List.of();
        }
    }

    @Override
    public List<Memory> findByImportance(String importance, String tenantId) {
        try (Jedis jedis = jedisPool.getResource()) {
            // Get all memories for tenant and filter by importance
            return findByTenantId(tenantId).stream()
                .filter(m -> importance.equals(m.getImportance().name()))
                .sorted((a, b) -> Integer.compare(b.getAccessCount(), a.getAccessCount()))
                .toList();
        } catch (Exception e) {
            log.error("Error finding memories by importance: {}", importance, e);
            return List.of();
        }
    }

    @Override
    public List<Memory> findByTags(List<String> tags, String tenantId) {
        try (Jedis jedis = jedisPool.getResource()) {
            // Find memory IDs that have ALL the specified tags
            Set<String> memoryIds = null;

            for (String tag : tags) {
                Set<String> tagMembers = jedis.smembers(TAG_INDEX + tag);
                if (memoryIds == null) {
                    memoryIds = new HashSet<>(tagMembers);
                } else {
                    memoryIds.retainAll(tagMembers);  // Intersection
                }
            }

            if (memoryIds == null || memoryIds.isEmpty()) {
                return List.of();
            }

            // Fetch full memories and filter by tenant
            List<Memory> memories = new ArrayList<>();
            for (String id : memoryIds) {
                findById(id).ifPresent(mem -> {
                    if (tenantId.equals(mem.getTenantId())) {
                        memories.add(mem);
                    }
                });
            }

            return memories;
        } catch (Exception e) {
            log.error("Error finding memories by tags: {}", tags, e);
            return List.of();
        }
    }

    @Override
    public List<Memory> vectorSearch(float[] embedding, int limit, String tenantId) {
        // Placeholder: return most recently accessed memories for tenant
        // Real vector search will be implemented with proper FalkorDB integration
        try {
            return findByTenantId(tenantId).stream()
                .sorted((a, b) -> Integer.compare(b.getAccessCount(), a.getAccessCount()))
                .limit(limit)
                .toList();
        } catch (Exception e) {
            log.error("Error in vector search (placeholder)", e);
            return List.of();
        }
    }

    @Override
    public List<Memory> findRelated(String memoryId, int depth, String tenantId) {
        // Placeholder: return memories with same tags
        return findById(memoryId)
                .map(memory -> {
                    if (memory.getTags() == null || memory.getTags().isEmpty()) {
                        return List.<Memory>of();
                    }
                    return findByTags(memory.getTags(), tenantId).stream()
                        .filter(m -> !m.getId().equals(memoryId))
                        .limit(depth * 5)
                        .toList();
                })
                .orElse(List.of());
    }

    @Override
    public boolean deleteById(String id) {
        try (Jedis jedis = jedisPool.getResource()) {
            String key = MEMORY_PREFIX + id;

            // Get memory before deletion for cleanup
            Optional<Memory> memoryOpt = findById(id);
            if (memoryOpt.isEmpty()) {
                return false;
            }

            Memory memory = memoryOpt.get();

            // Remove from KV store
            jedis.del(key);

            // Remove from tenant index
            String tenantId = memory.getTenantId();
            if (tenantId != null) {
                jedis.srem(TENANT_MEMORIES + tenantId, id);
            }

            // Remove tag indexes
            if (memory.getTags() != null) {
                for (String tag : memory.getTags()) {
                    jedis.srem(TAG_INDEX + tag, id);
                }
            }

            log.debug("Deleted memory: {}", id);
            return true;
        } catch (Exception e) {
            log.error("Error deleting memory: {}", id, e);
            return false;
        }
    }

    @Override
    public long countByTenantId(String tenantId) {
        try (Jedis jedis = jedisPool.getResource()) {
            Long count = jedis.scard(TENANT_MEMORIES + tenantId);
            return count != null ? count : 0L;
        } catch (Exception e) {
            log.error("Error counting memories for tenant: {}", tenantId, e);
            return 0;
        }
    }

    @Override
    public void archiveVersion(Memory memory) {
        // Version archiving is handled by PostgreSQL
        // This is a placeholder for potential Redis-based archiving
        log.debug("Archive version for memory: {}", memory.getId());
    }

    @Override
    public List<Integer> getVersionHistory(String memoryId, String tenantId) {
        // Version history is stored in PostgreSQL
        // This returns current version from Redis
        return findById(memoryId)
            .map(m -> List.of(m.getVersion()))
            .orElse(List.of());
    }

    // ==================== Private Methods ====================

    private String generateId() {
        // Simple ULID-like ID generation
        return "mem_" + UUID.randomUUID().toString().replace("-", "");
    }

    private List<Double> embeddingArrayToList(float[] array) {
        List<Double> list = new ArrayList<>(array.length);
        for (float v : array) {
            list.add((double) v);
        }
        return list;
    }
}
