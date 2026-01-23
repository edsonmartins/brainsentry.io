package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.dto.request.SearchRequest;
import com.integraltech.brainsentry.dto.request.UpdateMemoryRequest;
import com.integraltech.brainsentry.dto.response.GraphRelationshipResponse;
import com.integraltech.brainsentry.dto.response.MemoryListResponse;
import com.integraltech.brainsentry.dto.response.MemoryResponse;
import com.integraltech.brainsentry.service.MemoryService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

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
@RequiredArgsConstructor
public class MemoryController {

    private final MemoryService memoryService;

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
}
