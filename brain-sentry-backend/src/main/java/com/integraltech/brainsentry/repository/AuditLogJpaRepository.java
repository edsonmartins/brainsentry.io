package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.AuditLog;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

/**
 * JPA repository for AuditLog entity.
 */
@Repository
public interface AuditLogJpaRepository extends JpaRepository<AuditLog, String> {

    /**
     * Find audit logs by tenant ID.
     */
    List<AuditLog> findByTenantId(String tenantId);

    /**
     * Find audit logs by tenant ID and event type.
     */
    List<AuditLog> findByTenantIdAndEventType(String tenantId, String eventType);

    /**
     * Find audit logs by tenant ID and user ID.
     */
    List<AuditLog> findByTenantIdAndUserId(String tenantId, String userId);

    /**
     * Find audit logs by tenant ID and session ID.
     */
    List<AuditLog> findByTenantIdAndSessionId(String tenantId, String sessionId);

    /**
     * Find audit logs by tenant ID within a date range.
     */
    @Query("SELECT a FROM AuditLog a WHERE a.tenantId = :tenantId AND a.timestamp BETWEEN :startDate AND :endDate ORDER BY a.timestamp DESC")
    List<AuditLog> findByTenantIdAndTimestampBetween(
            @Param("tenantId") String tenantId,
            @Param("startDate") Instant startDate,
            @Param("endDate") Instant endDate
    );

    /**
     * Find recent audit logs by tenant ID.
     */
    @Query("SELECT a FROM AuditLog a WHERE a.tenantId = :tenantId ORDER BY a.timestamp DESC")
    List<AuditLog> findRecentByTenantId(@Param("tenantId") String tenantId);

    /**
     * Count audit logs by tenant ID and event type.
     */
    long countByTenantIdAndEventType(String tenantId, String eventType);

    /**
     * Delete old audit logs older than the specified date.
     */
    @Query("DELETE FROM AuditLog a WHERE a.tenantId = :tenantId AND a.timestamp < :beforeDate")
    int deleteOldAuditLogs(
            @Param("tenantId") String tenantId,
            @Param("beforeDate") Instant beforeDate
    );

    /**
     * Find the most recent audit log for a tenant.
     */
    Optional<AuditLog> findFirstByTenantIdOrderByTimestampDesc(String tenantId);

    /**
     * Count audit logs by user ID.
     */
    @Query("SELECT COUNT(a) FROM AuditLog a WHERE a.userId = :userId")
    long countByUserId(@Param("userId") String userId);
}
