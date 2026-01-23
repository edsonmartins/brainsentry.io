package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.domain.enums.ValidationStatus;
import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.dto.request.SearchRequest;
import com.integraltech.brainsentry.dto.request.UpdateMemoryRequest;
import com.integraltech.brainsentry.dto.response.GraphRelationshipResponse;
import com.integraltech.brainsentry.dto.response.MemoryListResponse;
import com.integraltech.brainsentry.dto.response.MemoryResponse;
import com.integraltech.brainsentry.mapper.MemoryMapper;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.repository.MemoryRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.PageRequest;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * Service for memory CRUD operations.
 *
 * Uses dual storage strategy:
 * - PostgreSQL (via JPA): Relational persistence with automatic tenant filtering via @TenantId
 * - FalkorDB (via Jedis): Graph operations and vector search
 *
 * Multi-tenancy is handled automatically by Hibernate 6 @TenantId annotation.
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class MemoryService {

    private final MemoryJpaRepository memoryJpaRepo;  // PostgreSQL (automatic tenant filtering)
    private final MemoryRepository memoryGraphRepo;    // FalkorDB (vector search + graph)
    private final EmbeddingService embeddingService;
    private final OpenRouterService openRouterService;
    private final MemoryMapper memoryMapper;

    /**
     * Create a new memory.
     *
     * @param request the create request
     * @return the created memory
     */
    @Transactional
    public MemoryResponse createMemory(CreateMemoryRequest request) {
        log.info("Creating memory with category: {}, tenant: {}",
            request.getCategory(), TenantContext.getTenantId());

        // Auto-analyze if category/importance not provided
        MemoryCategory category = request.getCategory();
        ImportanceLevel importance = request.getImportance();

        if (category == null || importance == null) {
            var analysis = openRouterService.analyzeImportance(request.getContent());
            if (category == null && analysis.getCategory() != null) {
                category = MemoryCategory.valueOf(analysis.getCategory());
            }
            if (importance == null && analysis.getImportance() != null) {
                importance = ImportanceLevel.valueOf(analysis.getImportance());
            }
        }

        // Generate embedding
        float[] embedding = embeddingService.embed(request.getContent());

        // Build memory
        Memory memory = Memory.builder()
            .id(generateMemoryId())
            .content(request.getContent())
            .summary(request.getSummary())
            .category(category != null ? category : MemoryCategory.PATTERN)
            .importance(importance != null ? importance : ImportanceLevel.MINOR)
            .tags(request.getTags())
            .metadata(request.getMetadata())
            .sourceType(request.getSourceType())
            .sourceReference(request.getSourceReference())
            .createdBy(request.getCreatedBy())
            .tenantId(request.getTenantId() != null ? request.getTenantId() : TenantContext.getTenantId())
            .embedding(embedding)
            .codeExample(request.getCodeExample())
            .programmingLanguage(request.getProgrammingLanguage())
            .validationStatus(ValidationStatus.PENDING)
            .version(1)
            .accessCount(0)
            .injectionCount(0)
            .helpfulCount(0)
            .notHelpfulCount(0)
            .createdAt(Instant.now())
            .build();

        // Save to PostgreSQL (Hibernate 6 filters by tenant automatically via @TenantId)
        Memory saved = memoryJpaRepo.save(memory);

        // Also store in FalkorDB for vector search and graph operations
        memoryGraphRepo.save(saved);

        log.info("Created memory: {} for tenant: {}", saved.getId(), saved.getTenantId());

        return memoryMapper.toResponse(saved);
    }

    /**
     * Get a memory by ID.
     *
     * @param id the memory ID
     * @return the memory
     * @throws jakarta.persistence.EntityNotFoundException if not found
     */
    @Transactional(readOnly = true)
    public MemoryResponse getMemory(String id) {
        Memory memory = memoryJpaRepo.findById(id)
            .orElseThrow(() -> new RuntimeException("Memory not found: " + id));

        // Update access count
        memory.setLastAccessedAt(Instant.now());
        memory.setAccessCount(memory.getAccessCount() + 1);
        memoryJpaRepo.save(memory);

        return memoryMapper.toResponse(memory);
    }

    /**
     * List memories with pagination.
     *
     * @param page page number (0-indexed)
     * @param size page size
     * @return paginated memory list (automatically filtered by current tenant)
     */
    @Transactional(readOnly = true)
    public MemoryListResponse listMemories(int page, int size) {
        // JPA with @TenantId automatically filters by current tenant
        var pageResult = memoryJpaRepo.findAll(PageRequest.of(page, size));

        List<MemoryResponse> responses = pageResult.getContent().stream()
            .map(memoryMapper::toResponse)
            .collect(Collectors.toList());

        return MemoryListResponse.builder()
            .memories(responses)
            .page(page)
            .size(size)
            .totalElements(pageResult.getTotalElements())
            .totalPages(pageResult.getTotalPages())
            .hasNext(pageResult.hasNext())
            .hasPrevious(pageResult.hasPrevious())
            .build();
    }

    /**
     * Update an existing memory.
     *
     * @param id the memory ID
     * @param request the update request
     * @return the updated memory
     */
    @Transactional
    public MemoryResponse updateMemory(String id, UpdateMemoryRequest request) {
        Memory existing = memoryJpaRepo.findById(id)
            .orElseThrow(() -> new RuntimeException("Memory not found: " + id));

        // Archive current version
        memoryGraphRepo.archiveVersion(existing);

        // Update fields
        if (request.getContent() != null) {
            existing.setContent(request.getContent());
            // Re-generate embedding
            existing.setEmbedding(embeddingService.embed(request.getContent()));
        }
        if (request.getSummary() != null) {
            existing.setSummary(request.getSummary());
        }
        if (request.getCategory() != null) {
            existing.setCategory(request.getCategory());
        }
        if (request.getImportance() != null) {
            existing.setImportance(request.getImportance());
        }
        if (request.getTags() != null) {
            existing.setTags(request.getTags());
        }
        if (request.getMetadata() != null) {
            existing.setMetadata(request.getMetadata());
        }
        if (request.getCodeExample() != null) {
            existing.setCodeExample(request.getCodeExample());
        }
        if (request.getProgrammingLanguage() != null) {
            existing.setProgrammingLanguage(request.getProgrammingLanguage());
        }

        existing.setVersion(existing.getVersion() + 1);
        existing.setUpdatedAt(Instant.now());

        // Save to both storages
        Memory saved = memoryJpaRepo.save(existing);
        memoryGraphRepo.save(existing);

        log.info("Updated memory: {} to version {}", saved.getId(), saved.getVersion());

        return memoryMapper.toResponse(saved);
    }

    /**
     * Delete a memory (soft delete).
     *
     * @param id the memory ID
     * @return true if deleted
     */
    @Transactional
    public boolean deleteMemory(String id) {
        if (!memoryJpaRepo.existsById(id)) {
            return false;
        }

        memoryJpaRepo.deleteById(id);
        memoryGraphRepo.deleteById(id);

        log.info("Deleted memory: {}", id);
        return true;
    }

    /**
     * Search memories using semantic vector search (FalkorDB).
     * Falls back to full-text search if vector search is unavailable.
     *
     * @param request the search request
     * @return list of matching memories
     */
    @Transactional(readOnly = true)
    public List<MemoryResponse> search(SearchRequest request) {
        List<Memory> results;

        try {
            // Try vector search first (FalkorDB)
            float[] embedding = embeddingService.embed(request.getQuery());
            results = memoryGraphRepo.vectorSearch(
                embedding,
                request.getLimit() != null ? request.getLimit() : 10,
                TenantContext.getTenantId()
            );

            // If vector search returns no results, fall back to full-text search
            if (results == null || results.isEmpty()) {
                log.debug("Vector search returned no results, falling back to full-text search");
                results = memoryJpaRepo.fullTextSearch(request.getQuery());

                // Apply limit
                int limit = request.getLimit() != null ? request.getLimit() : 10;
                if (results.size() > limit) {
                    results = results.subList(0, limit);
                }
            }
        } catch (Exception e) {
            log.warn("Vector search failed, falling back to full-text search: {}", e.getMessage());
            // Fallback to full-text search
            results = memoryJpaRepo.fullTextSearch(request.getQuery());

            // Apply limit
            int limit = request.getLimit() != null ? request.getLimit() : 10;
            if (results.size() > limit) {
                results = results.subList(0, limit);
            }
        }

        return results.stream()
            .map(memoryMapper::toResponse)
            .collect(Collectors.toList());
    }

    /**
     * Record feedback for a memory.
     *
     * @param id the memory ID
     * @param helpful whether the memory was helpful
     */
    @Transactional
    public void recordFeedback(String id, boolean helpful) {
        Memory memory = memoryJpaRepo.findById(id)
            .orElseThrow(() -> new RuntimeException("Memory not found: " + id));

        if (helpful) {
            memory.setHelpfulCount(memory.getHelpfulCount() + 1);
        } else {
            memory.setNotHelpfulCount(memory.getNotHelpfulCount() + 1);
        }

        memoryJpaRepo.save(memory);
        log.info("Recorded feedback for memory: {} helpful={}", id, helpful);
    }

    /**
     * Get memories by category (automatically filtered by tenant).
     *
     * @param category the category
     * @return list of memories
     */
    @Transactional(readOnly = true)
    public List<MemoryResponse> getByCategory(String category) {
        List<Memory> memories = memoryJpaRepo.findByCategory(category);
        return memories.stream()
            .map(memoryMapper::toResponse)
            .collect(Collectors.toList());
    }

    /**
     * Get memories by importance level (automatically filtered by tenant).
     *
     * @param importance the importance level
     * @return list of memories
     */
    @Transactional(readOnly = true)
    public List<MemoryResponse> getByImportance(String importance) {
        List<Memory> memories = memoryJpaRepo.findByImportance(importance);
        return memories.stream()
            .map(memoryMapper::toResponse)
            .collect(Collectors.toList());
    }

    /**
     * Find related memories through graph traversal.
     *
     * @param memoryId the starting memory ID
     * @param depth graph traversal depth
     * @return list of related memories
     */
    @Transactional(readOnly = true)
    public List<MemoryResponse> getRelated(String memoryId, int depth) {
        List<Memory> memories = memoryGraphRepo.findRelated(
            memoryId,
            depth,
            TenantContext.getTenantId()
        );
        return memories.stream()
            .map(memoryMapper::toResponse)
            .collect(Collectors.toList());
    }

    /**
     * Generate a unique memory ID.
     *
     * @return unique ID in format "mem_" + 12 hex characters
     */
    private String generateMemoryId() {
        return "mem_" + UUID.randomUUID().toString().replace("-", "").substring(0, 12);
    }

    /**
     * Reprocess all existing memories to FalkorDB graph.
     * This is useful when:
     * 1. FalkorDB integration was recently enabled
     * 2. Graph data was lost and needs to be rebuilt
     * 3. Relationships need to be recalculated
     *
     * @return number of memories reprocessed
     */
    @Transactional
    public int reprocessToGraph() {
        log.info("Starting reprocess of all memories to FalkorDB graph");

        // Get all memories for current tenant
        String tenantId = TenantContext.getTenantId();
        List<Memory> memories = memoryJpaRepo.findAll();

        int count = 0;
        for (Memory memory : memories) {
            try {
                // Save to FalkorDB graph
                memoryGraphRepo.save(memory);
                count++;
                log.debug("Reprocessed memory {} to graph", memory.getId());
            } catch (Exception e) {
                log.warn("Failed to reprocess memory {} to graph: {}", memory.getId(), e.getMessage());
            }
        }

        // Create all relationships between memories
        try {
            memoryGraphRepo.createAllRelationships(tenantId);
            log.info("Created graph relationships for {} memories", count);
        } catch (Exception e) {
            log.warn("Failed to create graph relationships: {}", e.getMessage());
        }

        log.info("Reprocessing complete: {} memories processed to FalkorDB graph", count);
        return count;
    }

    /**
     * Get all graph relationships from FalkorDB.
     * Returns all RELATED_TO relationships between memories.
     *
     * @return list of graph relationships
     */
    @Transactional(readOnly = true)
    public List<GraphRelationshipResponse> getGraphRelationships() {
        String tenantId = TenantContext.getTenantId();
        log.info("Fetching graph relationships for tenant: {}", tenantId);

        // Get all memories first to build a map
        List<Memory> allMemories = memoryJpaRepo.findAll();
        Map<String, Memory> memoryMap = allMemories.stream()
            .collect(Collectors.toMap(Memory::getId, m -> m));

        // Query all relationships from FalkorDB
        List<GraphRelationshipResponse> relationships = new ArrayList<>();
        try {
            // Query for all RELATED_TO relationships in the graph
            String query = String.format(
                "MATCH (m1:Memory)-[r:RELATED_TO]->(m2:Memory) " +
                "WHERE m1.tenantId = '%s' AND m2.tenantId = '%s' " +
                "RETURN m1.id as fromId, m2.id as toId, r.type as type, " +
                "r.strength as strength, r.tag as tag",
                tenantId, tenantId
            );

            log.debug("Executing graph query: {}", query);

            // First, let's check how many Memory nodes exist in the graph
            String countQuery = String.format(
                "MATCH (m:Memory) WHERE m.tenantId = '%s' RETURN count(m) as count",
                tenantId
            );
            var countResult = memoryGraphRepo.query(countQuery);
            for (var countRecord : countResult) {
                Object countValue = countRecord.getValue("count");
                log.debug("Memory count in graph: {}", String.valueOf(countValue));
            }

            // Also check relationship count
            String relCountQuery = String.format(
                "MATCH (m1:Memory)-[r:RELATED_TO]->(m2:Memory) WHERE m1.tenantId = '%s' RETURN count(r) as count",
                tenantId
            );
            var relCountResult = memoryGraphRepo.query(relCountQuery);
            for (var relCountRecord : relCountResult) {
                Object relCountValue = relCountRecord.getValue("count");
                log.debug("Relationship count in graph: {}", String.valueOf(relCountValue));
            }

            var resultSet = memoryGraphRepo.query(query);
            log.debug("Query result size: {}", resultSet.size());

            for (var record : resultSet) {
                log.debug("Processing record: {}", record);
                String fromId = record.getValue("fromId");
                String toId = record.getValue("toId");
                String type = record.getValue("type");
                Double strength = record.getValue("strength");
                String tag = record.getValue("tag");

                log.debug("Relationship: {} -> {} (type: {}, tag: {})", fromId, toId, type, tag);

                Memory fromMem = memoryMap.get(fromId);
                Memory toMem = memoryMap.get(toId);

                if (fromMem != null && toMem != null) {
                    relationships.add(new GraphRelationshipResponse(
                        fromId + "-" + toId,
                        fromId,
                        toId,
                        fromMem.getSummary(),
                        toMem.getSummary(),
                        type,
                        strength,
                        tag
                    ));
                } else {
                    log.warn("Memory not found in map: fromId={}, toId={}", fromId, toId);
                }
            }
        } catch (Exception e) {
            log.error("Failed to fetch graph relationships", e);
        }

        log.info("Found {} graph relationships", relationships.size());
        return relationships;
    }
}
