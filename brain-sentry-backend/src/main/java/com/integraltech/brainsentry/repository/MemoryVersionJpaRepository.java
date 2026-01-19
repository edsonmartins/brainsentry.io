package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.MemoryVersion;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

/**
 * JPA repository for MemoryVersion entity.
 * Provides version history and rollback capabilities for memories.
 */
@Repository
@ConditionalOnProperty(name = "features.versioning.enabled", havingValue = "true", matchIfMissing = true)
public interface MemoryVersionJpaRepository extends JpaRepository<MemoryVersion, String> {

    /**
     * Find all versions for a specific memory.
     */
    List<MemoryVersion> findByMemoryIdOrderByVersionDesc(String memoryId);

    /**
     * Find a specific version of a memory.
     */
    Optional<MemoryVersion> findByMemoryIdAndVersion(String memoryId, Integer version);

    /**
     * Find the latest version of a memory.
     */
    @Query("SELECT v FROM MemoryVersion v WHERE v.memoryId = :memoryId ORDER BY v.version DESC LIMIT 1")
    Optional<MemoryVersion> findLatestVersion(@Param("memoryId") String memoryId);

    /**
     * Find all versions for a specific memory by tenant.
     */
    @Query("SELECT v FROM MemoryVersion v WHERE v.memoryId = :memoryId AND v.tenantId = :tenantId ORDER BY v.version DESC")
    List<MemoryVersion> findByMemoryIdAndTenantId(@Param("memoryId") String memoryId, @Param("tenantId") String tenantId);

    /**
     * Count versions for a specific memory.
     */
    long countByMemoryId(String memoryId);

    /**
     * Delete all versions for a specific memory.
     */
    void deleteByMemoryId(String memoryId);

    /**
     * Find versions by change type.
     */
    List<MemoryVersion> findByChangeTypeAndTenantId(String changeType, String tenantId);

    /**
     * Find versions created by a specific user.
     */
    List<MemoryVersion> findByChangedByAndTenantId(String changedBy, String tenantId);
}
