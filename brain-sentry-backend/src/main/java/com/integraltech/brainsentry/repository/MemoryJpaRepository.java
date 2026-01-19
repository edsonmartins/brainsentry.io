package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

/**
 * JPA Repository for Memory entity.
 *
 * Works with PostgreSQL for persistent storage.
 * Multi-tenancy is handled automatically by Hibernate 6 @TenantId.
 *
 * Note: Vector search operations use MemoryRepositoryImpl with FalkorDB.
 * This repository handles CRUD and metadata operations.
 */
@Repository
public interface MemoryJpaRepository extends JpaRepository<Memory, String>, JpaSpecificationExecutor<Memory> {

    /**
     * Find memories by category (automatically filtered by current tenant).
     */
    List<Memory> findByCategory(String category);

    /**
     * Find memories by tenant ID and category enum.
     * Note: tenantId is included for explicit queries when needed.
     */
    List<Memory> findByTenantIdAndCategory(String tenantId, MemoryCategory category);

    /**
     * Find memories by importance level (automatically filtered by current tenant).
     */
    List<Memory> findByImportance(String importance);

    /**
     * Find memories by validation status.
     */
    List<Memory> findByValidationStatus(String validationStatus);

    /**
     * Find memories created after a specific date.
     */
    List<Memory> findByCreatedAtAfter(Instant date);

    /**
     * Find memories by a tag.
     * Note: Requires join with memory_tags table
     */
    @Query("SELECT m FROM Memory m JOIN m.tags t WHERE t = :tag")
    List<Memory> findByTag(@Param("tag") String tag);

    /**
     * Find memories with low usage for potential cleanup.
     */
    @Query("SELECT m FROM Memory m WHERE m.lastAccessedAt < :date OR m.lastAccessedAt IS NULL")
    List<Memory> findStaleMemories(@Param("date") Instant date);

    /**
     * Count memories by category.
     */
    long countByCategory(String category);

    /**
     * Count memories by importance level.
     */
    long countByImportance(String importance);

    /**
     * Find recently accessed memories (top 10).
     * Uses Pageable to limit results.
     */
    @Query("SELECT m FROM Memory m ORDER BY m.lastAccessedAt DESC")
    List<Memory> findTop10ByLastAccessedAtOrderByLastAccessedAtDesc(org.springframework.data.domain.Pageable pageable);

    /**
     * Find recently accessed memories without pagination.
     * Returns all memories ordered by last access time.
     */
    @Query("SELECT m FROM Memory m ORDER BY m.lastAccessedAt DESC")
    List<Memory> findAllByOrderByLastAccessedAtDesc();

    /**
     * Count memories created by a specific user.
     */
    long countByCreatedBy(String createdBy);
}
