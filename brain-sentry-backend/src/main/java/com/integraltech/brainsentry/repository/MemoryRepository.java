package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.Memory;

import java.util.List;
import java.util.Optional;

/**
 * Repository interface for Memory entities.
 *
 * Memories are stored in FalkorDB with vector embeddings
 * for semantic search capabilities.
 */
public interface MemoryRepository {

    /**
     * Save a new memory or update an existing one.
     *
     * @param memory the memory to save
     * @return the saved memory
     */
    Memory save(Memory memory);

    /**
     * Find a memory by its ID.
     *
     * @param id the memory ID
     * @return Optional containing the memory if found
     */
    Optional<Memory> findById(String id);

    /**
     * Find all memories for a given tenant.
     *
     * @param tenantId the tenant ID
     * @return list of memories
     */
    List<Memory> findByTenantId(String tenantId);

    /**
     * Find memories by category.
     *
     * @param category the category
     * @param tenantId the tenant ID
     * @return list of memories
     */
    List<Memory> findByCategory(String category, String tenantId);

    /**
     * Find memories by importance level.
     *
     * @param importance the importance level
     * @param tenantId the tenant ID
     * @return list of memories
     */
    List<Memory> findByImportance(String importance, String tenantId);

    /**
     * Search memories by tags.
     *
     * @param tags list of tags to search
     * @param tenantId the tenant ID
     * @return list of matching memories
     */
    List<Memory> findByTags(List<String> tags, String tenantId);

    /**
     * Semantic vector search for similar memories.
     *
     * @param embedding the query embedding
     * @param limit maximum number of results
     * @param tenantId the tenant ID
     * @return list of similar memories with scores
     */
    List<Memory> vectorSearch(float[] embedding, int limit, String tenantId);

    /**
     * Find related memories through graph relationships.
     *
     * @param memoryId the starting memory ID
     * @param depth depth of graph traversal
     * @param tenantId the tenant ID
     * @return list of related memories
     */
    List<Memory> findRelated(String memoryId, int depth, String tenantId);

    /**
     * Execute a raw Cypher query and return the ResultSet.
     * Used for custom graph queries.
     *
     * @param query the Cypher query string
     * @return ResultSet from FalkorDB
     */
    com.falkordb.ResultSet query(String query);

    /**
     * Delete a memory by ID (soft delete).
     *
     * @param id the memory ID
     * @return true if deleted
     */
    boolean deleteById(String id);

    /**
     * Count total memories for a tenant.
     *
     * @param tenantId the tenant ID
     * @return count of memories
     */
    long countByTenantId(String tenantId);

    /**
     * Archive current version before update.
     *
     * @param memory the memory to archive
     */
    void archiveVersion(Memory memory);

    /**
     * Get version history for a memory.
     *
     * @param memoryId the memory ID
     * @param tenantId the tenant ID
     * @return list of version numbers
     */
    List<Integer> getVersionHistory(String memoryId, String tenantId);

    /**
     * Create all graph relationships for a tenant's memories.
     * This is typically called after reprocessing memories to the graph.
     *
     * @param tenantId the tenant ID
     */
    void createAllRelationships(String tenantId);
}
