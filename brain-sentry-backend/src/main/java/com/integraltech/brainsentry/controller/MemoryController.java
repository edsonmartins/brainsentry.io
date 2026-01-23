package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.dto.request.SearchRequest;
import com.integraltech.brainsentry.dto.request.UpdateMemoryRequest;
import com.integraltech.brainsentry.dto.response.GraphRelationshipResponse;
import com.integraltech.brainsentry.dto.response.KnowledgeGraphResponse;
import com.integraltech.brainsentry.dto.response.MemoryListResponse;
import com.integraltech.brainsentry.dto.response.MemoryResponse;
import com.integraltech.brainsentry.service.EntityGraphService;
import com.integraltech.brainsentry.service.MemoryService;
import com.integraltech.brainsentry.config.TenantContext;
import jakarta.validation.Valid;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Collections;
import java.util.List;

/**
 * REST controller for memory operations.
 *
 * Provides CRUD endpoints for managing memories in the
 * Brain Sentry system.
 */
@Slf4j
@RestController
@RequestMapping("/v1/memories")
public class MemoryController {

    private final MemoryService memoryService;
    private final EntityGraphService entityGraphService;  // May be null if feature disabled

    public MemoryController(MemoryService memoryService,
                           @Autowired(required = false) EntityGraphService entityGraphService) {
        this.memoryService = memoryService;
        this.entityGraphService = entityGraphService;
    }

    /**
     * Create a new memory.
     * POST /api/v1/memories
     */
    @PostMapping
    public ResponseEntity<MemoryResponse> createMemory(
        @Valid @RequestBody CreateMemoryRequest request
    ) {
        log.info("POST /v1/memories - category: {}", request.getCategory());
        MemoryResponse response = memoryService.createMemory(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    /**
     * Get a memory by ID.
     * GET /api/v1/memories/{id}
     */
    @GetMapping("/{id}")
    public ResponseEntity<MemoryResponse> getMemory(@PathVariable String id) {
        log.info("GET /v1/memories/{}", id);
        MemoryResponse response = memoryService.getMemory(id);
        return ResponseEntity.ok(response);
    }

    /**
     * List memories with pagination.
     * GET /api/v1/memories?page=0&size=20
     *
     * Note: Tenant is automatically extracted from X-Tenant-ID header by TenantFilter.
     */
    @GetMapping
    public ResponseEntity<MemoryListResponse> listMemories(
        @RequestParam(defaultValue = "0") int page,
        @RequestParam(defaultValue = "20") int size,
        @RequestParam(required = false) String tenantId  // Optional override for testing
    ) {
        log.info("GET /v1/memories - page: {}, size: {}, tenant override: {}", page, size, tenantId);
        // If tenantId is provided, temporarily set it in context
        if (tenantId != null && !tenantId.isEmpty()) {
            com.integraltech.brainsentry.config.TenantContext.setTenantId(tenantId);
        }
        MemoryListResponse response = memoryService.listMemories(page, size);
        return ResponseEntity.ok(response);
    }

    /**
     * Update a memory.
     * PUT /api/v1/memories/{id}
     */
    @PutMapping("/{id}")
    public ResponseEntity<MemoryResponse> updateMemory(
        @PathVariable String id,
        @Valid @RequestBody UpdateMemoryRequest request
    ) {
        log.info("PUT /v1/memories/{}", id);
        MemoryResponse response = memoryService.updateMemory(id, request);
        return ResponseEntity.ok(response);
    }

    /**
     * Delete a memory.
     * DELETE /api/v1/memories/{id}
     */
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteMemory(@PathVariable String id) {
        log.info("DELETE /v1/memories/{}", id);
        memoryService.deleteMemory(id);
        return ResponseEntity.noContent().build();
    }

    /**
     * Search memories using semantic search.
     * POST /api/v1/memories/search
     */
    @PostMapping("/search")
    public ResponseEntity<List<MemoryResponse>> search(@Valid @RequestBody SearchRequest request) {
        log.info("POST /v1/memories/search - query: {}", request.getQuery());
        List<MemoryResponse> results = memoryService.search(request);
        return ResponseEntity.ok(results);
    }

    /**
     * Get memories by category.
     * GET /api/v1/memories/by-category/{category}
     *
     * Note: Tenant is automatically extracted from X-Tenant-ID header.
     */
    @GetMapping("/by-category/{category}")
    public ResponseEntity<List<MemoryResponse>> getByCategory(
        @PathVariable String category
    ) {
        log.info("GET /v1/memories/by-category/{}", category);
        List<MemoryResponse> results = memoryService.getByCategory(category);
        return ResponseEntity.ok(results);
    }

    /**
     * Get memories by importance level.
     * GET /api/v1/memories/by-importance/{importance}
     *
     * Note: Tenant is automatically extracted from X-Tenant-ID header.
     */
    @GetMapping("/by-importance/{importance}")
    public ResponseEntity<List<MemoryResponse>> getByImportance(
        @PathVariable String importance
    ) {
        log.info("GET /v1/memories/by-importance/{}", importance);
        List<MemoryResponse> results = memoryService.getByImportance(importance);
        return ResponseEntity.ok(results);
    }

    /**
     * Find related memories.
     * GET /api/v1/memories/{id}/related?depth=2
     *
     * Note: Tenant is automatically extracted from X-Tenant-ID header.
     */
    @GetMapping("/{id}/related")
    public ResponseEntity<List<MemoryResponse>> getRelated(
        @PathVariable String id,
        @RequestParam(defaultValue = "2") int depth
    ) {
        log.info("GET /v1/memories/{}/related - depth: {}", id, depth);
        List<MemoryResponse> results = memoryService.getRelated(id, depth);
        return ResponseEntity.ok(results);
    }

    /**
     * Record feedback for a memory.
     * POST /api/v1/memories/{id}/feedback?helpful=true
     */
    @PostMapping("/{id}/feedback")
    public ResponseEntity<Void> recordFeedback(
        @PathVariable String id,
        @RequestParam boolean helpful
    ) {
        log.info("POST /v1/memories/{}/feedback - helpful: {}", id, helpful);
        memoryService.recordFeedback(id, helpful);
        return ResponseEntity.ok().build();
    }

    /**
     * Reprocess all memories to FalkorDB graph.
     * POST /api/v1/memories/reprocess-graph
     *
     * This endpoint reprocesses all existing memories and stores them in FalkorDB
     * with proper embeddings and relationships.
     */
    @PostMapping("/reprocess-graph")
    public ResponseEntity<String> reprocessGraph() {
        log.info("POST /v1/memories/reprocess-graph - Starting graph reprocessing");
        int count = memoryService.reprocessToGraph();
        return ResponseEntity.ok("Reprocessed " + count + " memories to FalkorDB graph");
    }

    /**
     * Get all graph relationships from FalkorDB.
     * GET /api/v1/memories/relationships
     *
     * Returns all relationships between memories stored in the FalkorDB graph.
     */
    @GetMapping("/relationships")
    public ResponseEntity<List<GraphRelationshipResponse>> getGraphRelationships() {
        log.info("GET /v1/memories/relationships - Fetching all graph relationships");
        List<GraphRelationshipResponse> relationships = memoryService.getGraphRelationships();
        return ResponseEntity.ok(relationships);
    }

    /**
     * Get the knowledge graph with extracted entities and their relationships.
     * GET /api/v1/memories/knowledge-graph?limit=100
     *
     * Returns entities (nodes) and relationships (edges) extracted from memories.
     * This is different from memory-to-memory relationships - these are entities
     * extracted FROM the content of messages (like CLIENTE, PRODUTO, PEDIDO, etc.).
     */
    @GetMapping("/knowledge-graph")
    public ResponseEntity<KnowledgeGraphResponse> getKnowledgeGraph(
            @RequestParam(defaultValue = "100") int limit
    ) {
        log.info("GET /v1/memories/knowledge-graph - limit: {}", limit);

        if (entityGraphService == null) {
            log.warn("EntityGraphService not available (feature disabled)");
            return ResponseEntity.ok(KnowledgeGraphResponse.builder()
                    .nodes(Collections.emptyList())
                    .edges(Collections.emptyList())
                    .totalNodes(0)
                    .totalEdges(0)
                    .build());
        }

        String tenantId = TenantContext.getTenantId();
        KnowledgeGraphResponse response = entityGraphService.getKnowledgeGraphResponse(tenantId, limit);
        return ResponseEntity.ok(response);
    }

    /**
     * Extract entities from a specific memory.
     * POST /api/v1/memories/{id}/extract-entities
     *
     * Manually triggers entity extraction for a specific memory.
     * Useful for reprocessing existing memories that were created before
     * the entity extraction feature was enabled.
     */
    @PostMapping("/{id}/extract-entities")
    public ResponseEntity<String> extractEntitiesFromMemory(@PathVariable String id) {
        log.info("POST /v1/memories/{}/extract-entities", id);

        if (entityGraphService == null) {
            log.warn("EntityGraphService not available (feature disabled)");
            return ResponseEntity.badRequest().body("Entity extraction feature is disabled");
        }

        try {
            MemoryResponse memoryResponse = memoryService.getMemory(id);
            if (memoryResponse == null) {
                return ResponseEntity.notFound().build();
            }

            // Create a Memory object from the response for extraction
            com.integraltech.brainsentry.domain.Memory memory = new com.integraltech.brainsentry.domain.Memory();
            memory.setId(memoryResponse.getId());
            memory.setContent(memoryResponse.getContent());
            memory.setSummary(memoryResponse.getSummary());
            memory.setTenantId(TenantContext.getTenantId());

            // Trigger extraction synchronously for this endpoint
            entityGraphService.extractAndStoreEntitiesSync(memory, TenantContext.getTenantId());

            return ResponseEntity.ok("Entity extraction triggered for memory: " + id);
        } catch (Exception e) {
            log.error("Error extracting entities from memory {}: {}", id, e.getMessage());
            return ResponseEntity.internalServerError().body("Error: " + e.getMessage());
        }
    }

    /**
     * Extract entities from all existing memories.
     * POST /api/v1/memories/extract-all-entities
     *
     * Reprocesses all memories for the current tenant to extract entities.
     * This can be a long-running operation for tenants with many memories.
     */
    @PostMapping("/extract-all-entities")
    public ResponseEntity<String> extractEntitiesFromAllMemories() {
        log.info("POST /v1/memories/extract-all-entities");

        if (entityGraphService == null) {
            log.warn("EntityGraphService not available (feature disabled)");
            return ResponseEntity.badRequest().body("Entity extraction feature is disabled");
        }

        try {
            String tenantId = TenantContext.getTenantId();
            MemoryListResponse memories = memoryService.listMemories(0, 100);

            int count = 0;
            for (MemoryResponse memoryResponse : memories.getMemories()) {
                try {
                    com.integraltech.brainsentry.domain.Memory memory = new com.integraltech.brainsentry.domain.Memory();
                    memory.setId(memoryResponse.getId());
                    memory.setContent(memoryResponse.getContent());
                    memory.setSummary(memoryResponse.getSummary());
                    memory.setTenantId(tenantId);

                    entityGraphService.extractAndStoreEntitiesSync(memory, tenantId);
                    count++;
                } catch (Exception e) {
                    log.warn("Failed to extract entities from memory {}: {}", memoryResponse.getId(), e.getMessage());
                }
            }

            return ResponseEntity.ok("Entity extraction completed for " + count + " memories");
        } catch (Exception e) {
            log.error("Error extracting entities from all memories: {}", e.getMessage());
            return ResponseEntity.internalServerError().body("Error: " + e.getMessage());
        }
    }
}
