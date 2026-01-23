package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.MemoryRelationship;
import com.integraltech.brainsentry.domain.enums.RelationshipType;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.repository.MemoryRelationshipJpaRepository;
import com.integraltech.brainsentry.repository.MemoryRepository;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * Service for managing relationships between memories.
 *
 * Relationships enable the graph-based memory retrieval system,
 * allowing context expansion through related concepts.
 */
@Slf4j
@Service
@ConditionalOnProperty(name = "features.relationship.enabled", havingValue = "true", matchIfMissing = false)
public class RelationshipService {

    private final MemoryRelationshipJpaRepository relationshipRepo;
    private final MemoryJpaRepository memoryJpaRepo;
    private final AuditService auditService;
    private final MemoryRepository memoryRepository;
    private final OpenRouterService openRouterService;

    private static final double RELATIONSHIP_CONFIDENCE_THRESHOLD = 0.7;
    private static final int MAX_SIMILAR_MEMORIES_TO_ANALYZE = 10;

    public RelationshipService(MemoryRelationshipJpaRepository relationshipRepo,
                               MemoryJpaRepository memoryJpaRepo,
                               AuditService auditService,
                               @Autowired(required = false) MemoryRepository memoryRepository,
                               @Autowired(required = false) OpenRouterService openRouterService) {
        this.relationshipRepo = relationshipRepo;
        this.memoryJpaRepo = memoryJpaRepo;
        this.auditService = auditService;
        this.memoryRepository = memoryRepository;
        this.openRouterService = openRouterService;
    }

    /**
     * Create a relationship between two memories.
     *
     * @param fromMemoryId source memory ID
     * @param toMemoryId target memory ID
     * @param type relationship type
     * @param tenantId tenant ID
     * @param userId user creating the relationship
     * @return created relationship
     */
    @Transactional
    public MemoryRelationship createRelationship(String fromMemoryId, String toMemoryId,
                                                 RelationshipType type, String tenantId, String userId) {
        // Verify both memories exist
        if (!memoryJpaRepo.existsById(fromMemoryId)) {
            throw new IllegalArgumentException("Source memory not found: " + fromMemoryId);
        }
        if (!memoryJpaRepo.existsById(toMemoryId)) {
            throw new IllegalArgumentException("Target memory not found: " + toMemoryId);
        }

        // Check if relationship already exists
        Optional<MemoryRelationship> existing = relationshipRepo.findByFromAndTo(fromMemoryId, toMemoryId);
        if (existing.isPresent()) {
            // Update frequency and lastUsedAt
            MemoryRelationship rel = existing.get();
            rel.setFrequency(rel.getFrequency() + 1);
            rel.setLastUsedAt(Instant.now());
            rel.setTenantId(tenantId);
            return relationshipRepo.save(rel);
        }

        // Create new relationship
        MemoryRelationship relationship = MemoryRelationship.builder()
                .id(UUID.randomUUID().toString())
                .fromMemoryId(fromMemoryId)
                .toMemoryId(toMemoryId)
                .type(type)
                .frequency(1)
                .strength(0.5)
                .createdAt(Instant.now())
                .lastUsedAt(Instant.now())
                .tenantId(tenantId)
                .build();

        relationship = relationshipRepo.save(relationship);

        // Log the relationship creation
        auditService.logRelationshipCreated(fromMemoryId, toMemoryId, type.getDisplayName(), userId, tenantId);

        log.debug("Relationship created: {} -> {} ({})", fromMemoryId, toMemoryId, type);
        return relationship;
    }

    /**
     * Create a relationship bidirectionally (both directions).
     *
     * @param memoryId1 first memory ID
     * @param memoryId2 second memory ID
     * @param type1 relationship type from memory1 to memory2
     * @param type2 relationship type from memory2 to memory1
     * @param tenantId tenant ID
     * @param userId user creating the relationships
     * @return list of created relationships
     */
    @Transactional
    public List<MemoryRelationship> createBidirectionalRelationship(String memoryId1, String memoryId2,
                                                                      RelationshipType type1, RelationshipType type2,
                                                                      String tenantId, String userId) {
        MemoryRelationship rel1 = createRelationship(memoryId1, memoryId2, type1, tenantId, userId);
        MemoryRelationship rel2 = createRelationship(memoryId2, memoryId1, type2, tenantId, userId);
        return List.of(rel1, rel2);
    }

    /**
     * Get all relationships for a specific memory (as source).
     *
     * @param memoryId the memory ID
     * @param tenantId tenant ID
     * @return list of relationships
     */
    @Transactional(readOnly = true)
    public List<MemoryRelationship> getRelationshipsFrom(String memoryId, String tenantId) {
        return relationshipRepo.findByFromMemoryIdAndTenantId(memoryId, tenantId);
    }

    /**
     * Get all relationships pointing to a specific memory (as target).
     *
     * @param memoryId the memory ID
     * @return list of relationships
     */
    @Transactional(readOnly = true)
    public List<MemoryRelationship> getRelationshipsTo(String memoryId) {
        return relationshipRepo.findByToMemoryId(memoryId);
    }

    /**
     * Get all relationships for a tenant.
     *
     * @param tenantId tenant ID
     * @return list of all relationships
     */
    @Transactional(readOnly = true)
    public List<MemoryRelationship> getAllRelationships(String tenantId) {
        return relationshipRepo.findByTenantId(tenantId);
    }

    /**
     * Get relationship between two specific memories.
     *
     * @param fromMemoryId source memory ID
     * @param toMemoryId target memory ID
     * @return optional relationship
     */
    @Transactional(readOnly = true)
    public Optional<MemoryRelationship> getRelationship(String fromMemoryId, String toMemoryId) {
        return relationshipRepo.findByFromAndTo(fromMemoryId, toMemoryId);
    }

    /**
     * Delete a relationship between two memories.
     *
     * @param fromMemoryId source memory ID
     * @param toMemoryId target memory ID
     * @param tenantId tenant ID
     * @return true if deleted, false if not found
     */
    @Transactional
    public boolean deleteRelationship(String fromMemoryId, String toMemoryId, String tenantId) {
        Optional<MemoryRelationship> existing = relationshipRepo.findByFromAndTo(fromMemoryId, toMemoryId);
        if (existing.isPresent() && existing.get().getTenantId().equals(tenantId)) {
            relationshipRepo.delete(existing.get());
            log.debug("Relationship deleted: {} -> {}", fromMemoryId, toMemoryId);
            return true;
        }
        return false;
    }

    /**
     * Delete all relationships for a specific memory.
     *
     * @param memoryId the memory ID
     */
    @Transactional
    public void deleteAllRelationshipsForMemory(String memoryId) {
        relationshipRepo.deleteByFromMemoryId(memoryId);
        log.debug("All relationships deleted for memory: {}", memoryId);
    }

    /**
     * Update relationship strength.
     *
     * @param relationshipId the relationship ID
     * @param strength new strength value (0.0 to 1.0)
     * @return updated relationship
     */
    @Transactional
    public MemoryRelationship updateStrength(String relationshipId, double strength) {
        MemoryRelationship relationship = relationshipRepo.findById(relationshipId)
                .orElseThrow(() -> new IllegalArgumentException("Relationship not found: " + relationshipId));

        if (strength < 0.0 || strength > 1.0) {
            throw new IllegalArgumentException("Strength must be between 0.0 and 1.0");
        }

        relationship.setStrength(strength);
        relationship.setLastUsedAt(Instant.now());
        return relationshipRepo.save(relationship);
    }

    /**
     * Find related memories based on relationship type and strength.
     *
     * @param memoryId the memory ID
     * @param tenantId tenant ID
     * @param minStrength minimum strength threshold
     * @return list of related memory IDs with their strength
     */
    @Transactional(readOnly = true)
    public List<RelatedMemory> findRelatedMemories(String memoryId, String tenantId, double minStrength) {
        List<MemoryRelationship> relationships = relationshipRepo.findByFromMemoryIdAndTenantId(memoryId, tenantId);

        return relationships.stream()
                .filter(rel -> rel.getStrength() != null && rel.getStrength() >= minStrength)
                .map(rel -> new RelatedMemory(rel.getToMemoryId(), rel.getType(), rel.getStrength()))
                .toList();
    }

    /**
     * Auto-create relationships based on semantic similarity.
     *
     * @param memoryId the memory to find relationships for
     * @param tenantId tenant ID
     * @param threshold similarity threshold
     * @return list of created relationships
     */
    @Transactional
    public List<MemoryRelationship> suggestRelationships(String memoryId, String tenantId, double threshold) {
        // This would integrate with the embedding service to find semantically similar memories
        // and automatically create RELATED_TO relationships
        // For now, return empty list as this requires more complex logic
        log.debug("Relationship suggestion not yet implemented for memory: {}", memoryId);
        return List.of();
    }

    /**
     * Automatically detect and create relationships for a new memory.
     *
     * This method:
     * 1. Searches for similar memories using vector search
     * 2. Uses LLM (OpenRouter/Grok) to analyze the relationship type
     * 3. Creates relationships with high confidence scores
     *
     * @param newMemory the newly created memory
     * @param tenantId tenant ID
     * @return list of created relationships
     */
    @Async
    @Transactional
    public List<MemoryRelationship> detectAndCreateRelationships(Memory newMemory, String tenantId) {
        List<MemoryRelationship> createdRelationships = new ArrayList<>();

        // Check if required services are available
        if (memoryRepository == null) {
            log.warn("MemoryRepository not available, skipping automatic relationship detection");
            return createdRelationships;
        }

        if (openRouterService == null || !openRouterService.isConfigured()) {
            log.warn("OpenRouterService not configured, skipping LLM-based relationship detection");
            return createdRelationships;
        }

        if (newMemory.getEmbedding() == null || newMemory.getEmbedding().length == 0) {
            log.warn("Memory {} has no embedding, skipping relationship detection", newMemory.getId());
            return createdRelationships;
        }

        try {
            // 1. Search for similar memories via vector search
            log.debug("Searching for similar memories to {} for relationship detection", newMemory.getId());
            List<Memory> similarMemories = memoryRepository.vectorSearch(
                    newMemory.getEmbedding(),
                    MAX_SIMILAR_MEMORIES_TO_ANALYZE,
                    tenantId
            );

            if (similarMemories.isEmpty()) {
                log.debug("No similar memories found for {}", newMemory.getId());
                return createdRelationships;
            }

            log.info("Found {} similar memories for relationship analysis", similarMemories.size());

            // 2. For each similar memory, use LLM to determine relationship type
            for (Memory similarMemory : similarMemories) {
                // Skip if it's the same memory
                if (similarMemory.getId().equals(newMemory.getId())) {
                    continue;
                }

                try {
                    // Analyze relationship using LLM
                    OpenRouterService.RelationshipAnalysis analysis = openRouterService.analyzeRelationship(
                            newMemory.getContent(),
                            similarMemory.getContent()
                    );

                    // 3. Create relationship if confidence is high enough
                    if (analysis.isHasRelationship() && analysis.getConfidence() >= RELATIONSHIP_CONFIDENCE_THRESHOLD) {
                        RelationshipType relType = parseRelationshipType(analysis.getType());

                        MemoryRelationship relationship = createRelationship(
                                newMemory.getId(),
                                similarMemory.getId(),
                                relType,
                                tenantId,
                                "system" // Created automatically by the system
                        );

                        createdRelationships.add(relationship);

                        log.info("Auto-created relationship: {} --[{}]--> {} (confidence: {})",
                                newMemory.getId(), relType, similarMemory.getId(), analysis.getConfidence());
                    }
                } catch (Exception e) {
                    log.warn("Error analyzing relationship between {} and {}: {}",
                            newMemory.getId(), similarMemory.getId(), e.getMessage());
                }
            }

            log.info("Created {} automatic relationships for memory {}", createdRelationships.size(), newMemory.getId());

        } catch (Exception e) {
            log.error("Error during automatic relationship detection for memory {}", newMemory.getId(), e);
        }

        return createdRelationships;
    }

    /**
     * Parse relationship type from string to enum.
     * Falls back to RELATED_TO if unknown type.
     */
    private RelationshipType parseRelationshipType(String type) {
        if (type == null) {
            return RelationshipType.RELATED_TO;
        }

        try {
            return RelationshipType.valueOf(type.toUpperCase());
        } catch (IllegalArgumentException e) {
            log.debug("Unknown relationship type '{}', using RELATED_TO", type);
            return RelationshipType.RELATED_TO;
        }
    }

    /**
     * Record value: DTO for related memory with relationship info.
     */
    public record RelatedMemory(
            String memoryId,
            RelationshipType type,
            Double strength
    ) {}
}
