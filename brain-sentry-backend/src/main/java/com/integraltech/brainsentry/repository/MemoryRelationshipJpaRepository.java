package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.MemoryRelationship;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

/**
 * JPA repository for MemoryRelationship entity.
 * Note: This repository is only enabled when the feature flag is set.
 */
@Repository
@ConditionalOnProperty(name = "features.relationship.enabled", havingValue = "true", matchIfMissing = false)
public interface MemoryRelationshipJpaRepository extends JpaRepository<MemoryRelationship, String> {

    /**
     * Find all relationships for a specific memory (as source).
     */
    List<MemoryRelationship> findByFromMemoryId(String fromMemoryId);

    /**
     * Find all relationships pointing to a specific memory (as target).
     */
    List<MemoryRelationship> findByToMemoryId(String toMemoryId);

    /**
     * Find relationships between two specific memories.
     */
    @Query("SELECT r FROM MemoryRelationship r WHERE r.fromMemoryId = :fromId AND r.toMemoryId = :toId")
    Optional<MemoryRelationship> findByFromAndTo(@Param("fromId") String fromMemoryId, @Param("toId") String toMemoryId);

    /**
     * Find all relationships for a specific memory by tenant.
     */
    @Query("SELECT r FROM MemoryRelationship r WHERE r.fromMemoryId = :memoryId AND r.tenantId = :tenantId")
    List<MemoryRelationship> findByFromMemoryIdAndTenantId(@Param("memoryId") String memoryId, @Param("tenantId") String tenantId);

    /**
     * Find all relationships by tenant ID.
     */
    List<MemoryRelationship> findByTenantId(String tenantId);

    /**
     * Find relationships by type and tenant.
     */
    List<MemoryRelationship> findByTypeAndTenantId(com.integraltech.brainsentry.domain.enums.RelationshipType type, String tenantId);

    /**
     * Delete all relationships for a specific memory.
     */
    void deleteByFromMemoryId(String fromMemoryId);

    /**
     * Delete relationships between two specific memories.
     */
    void deleteByFromMemoryIdAndToMemoryId(String fromMemoryId, String toMemoryId);
}
