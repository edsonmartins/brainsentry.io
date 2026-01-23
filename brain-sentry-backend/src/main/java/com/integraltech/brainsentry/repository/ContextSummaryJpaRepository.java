package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.ContextSummary;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;

/**
 * JPA repository for ContextSummary entity.
 *
 * Manages persistence of context compression events.
 */
@Repository
public interface ContextSummaryJpaRepository extends JpaRepository<ContextSummary, String> {

    /**
     * Find all summaries for a tenant.
     */
    List<ContextSummary> findByTenantId(String tenantId);

    /**
     * Find all summaries for a tenant with pagination.
     */
    Page<ContextSummary> findByTenantId(String tenantId, Pageable pageable);

    /**
     * Find summaries by session ID.
     */
    List<ContextSummary> findBySessionId(String sessionId);

    /**
     * Find most recent summary for a session.
     */
    @Query("SELECT s FROM ContextSummary s WHERE s.sessionId = :sessionId ORDER BY s.createdAt DESC")
    List<ContextSummary> findLatestBySessionId(@Param("sessionId") String sessionId);

    /**
     * Find effective compressions (>25% reduction).
     */
    @Query("SELECT s FROM ContextSummary s WHERE s.tenantId = :tenantId AND s.originalTokenCount - s.compressedTokenCount > (s.originalTokenCount * 0.25)")
    List<ContextSummary> findEffectiveByTenantId(@Param("tenantId") String tenantId);

    /**
     * Find summaries created after a certain date.
     */
    List<ContextSummary> findByTenantIdAndCreatedAtAfter(String tenantId, Instant date);

    /**
     * Find summaries by compression method.
     */
    List<ContextSummary> findByTenantIdAndCompressionMethod(String tenantId, String method);

    /**
     * Calculate average compression ratio for a tenant.
     */
    @Query("SELECT AVG(s.compressionRatio) FROM ContextSummary s WHERE s.tenantId = :tenantId")
    Double getAverageCompressionRatio(@Param("tenantId") String tenantId);

    /**
     * Get total token savings for a tenant.
     */
    @Query("SELECT SUM(s.originalTokenCount - s.compressedTokenCount) FROM ContextSummary s WHERE s.tenantId = :tenantId")
    Long getTotalTokenSavings(@Param("tenantId") String tenantId);

    /**
     * Count compressions per session.
     */
    @Query("SELECT s.sessionId, COUNT(s) as compressionCount FROM ContextSummary s WHERE s.tenantId = :tenantId GROUP BY s.sessionId")
    List<Object[]> countBySessionId(@Param("tenantId") String tenantId);

    /**
     * Find summaries with poor compression (>0.7 ratio).
     */
    @Query("SELECT s FROM ContextSummary s WHERE s.tenantId = :tenantId AND s.compressionRatio > 0.7")
    List<ContextSummary> findPoorCompressionsByTenantId(@Param("tenantId") String tenantId);

    /**
     * Find all summaries created in a date range.
     */
    List<ContextSummary> findByTenantIdAndCreatedAtBetween(String tenantId, Instant start, Instant end);
}
