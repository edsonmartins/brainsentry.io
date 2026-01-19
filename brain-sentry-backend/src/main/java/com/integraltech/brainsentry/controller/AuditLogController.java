package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.AuditLog;
import com.integraltech.brainsentry.dto.response.AuditLogResponse;
import com.integraltech.brainsentry.mapper.AuditLogMapper;
import com.integraltech.brainsentry.service.AuditService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.constraints.Min;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.Instant;
import java.util.List;

/**
 * REST controller for audit log operations.
 *
 * Provides endpoints for querying and analyzing system audit logs.
 */
@RestController
@RequestMapping("/v1/audit-logs")
@RequiredArgsConstructor
@Tag(name = "Audit Logs", description = "Consultas e análises de logs de auditoria")
public class AuditLogController {

    private static final Logger log = LoggerFactory.getLogger(AuditLogController.class);

    private final AuditService auditService;

    /**
     * Get all audit logs for the current tenant.
     *
     * GET /v1/audit-logs
     */
    @GetMapping
    @Operation(summary = "Listar logs de auditoria", description = "Retorna todos os logs de auditoria do tenant atual")
    public ResponseEntity<List<AuditLogResponse>> getAuditLogs() {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/audit-logs - tenant: {}", tenant);

        List<AuditLog> logs = auditService.getAuditLogsByTenant(tenant);
        return ResponseEntity.ok(AuditLogMapper.toResponseList(logs));
    }

    /**
     * Get audit logs by event type.
     *
     * GET /v1/audit-logs/by-event/{eventType}
     */
    @GetMapping("/by-event/{eventType}")
    @Operation(summary = "Listar logs por tipo de evento", description = "Retorna logs filtrados por tipo de evento")
    public ResponseEntity<List<AuditLogResponse>> getAuditLogsByEventType(
            @Parameter(description = "Tipo do evento (ex: context_injection, memory_created)")
            @PathVariable String eventType) {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/audit-logs/by-event/{} - tenant: {}", eventType, tenant);

        List<AuditLog> logs = auditService.getAuditLogsByTenantAndEventType(tenant, eventType);
        return ResponseEntity.ok(AuditLogMapper.toResponseList(logs));
    }

    /**
     * Get audit logs by user.
     *
     * GET /v1/audit-logs/by-user/{userId}
     */
    @GetMapping("/by-user/{userId}")
    @Operation(summary = "Listar logs por usuário", description = "Retorna logs filtrados por usuário")
    public ResponseEntity<List<AuditLogResponse>> getAuditLogsByUser(
            @Parameter(description = "ID do usuário")
            @PathVariable String userId) {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/audit-logs/by-user/{} - tenant: {}", userId, tenant);

        List<AuditLog> logs = auditService.getAuditLogsByTenantAndUser(tenant, userId);
        return ResponseEntity.ok(AuditLogMapper.toResponseList(logs));
    }

    /**
     * Get audit logs by session.
     *
     * GET /v1/audit-logs/by-session/{sessionId}
     */
    @GetMapping("/by-session/{sessionId}")
    @Operation(summary = "Listar logs por sessão", description = "Retorna logs de uma sessão específica")
    public ResponseEntity<List<AuditLogResponse>> getAuditLogsBySession(
            @Parameter(description = "ID da sessão")
            @PathVariable String sessionId) {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/audit-logs/by-session/{} - tenant: {}", sessionId, tenant);

        List<AuditLog> logs = auditService.getAuditLogsByTenantAndSession(tenant, sessionId);
        return ResponseEntity.ok(AuditLogMapper.toResponseList(logs));
    }

    /**
     * Get recent audit logs.
     *
     * GET /v1/audit-logs/recent?limit=10
     */
    @GetMapping("/recent")
    @Operation(summary = "Listar logs recentes", description = "Retorna os logs mais recentes, limitados pela quantidade especificada")
    public ResponseEntity<List<AuditLogResponse>> getRecentAuditLogs(
            @Parameter(description = "Número máximo de logs a retornar")
            @RequestParam(defaultValue = "10") @Min(1) int limit) {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/audit-logs/recent?limit={} - tenant: {}", limit, tenant);

        List<AuditLog> logs = auditService.getRecentAuditLogs(tenant, limit);
        return ResponseEntity.ok(AuditLogMapper.toResponseList(logs));
    }

    /**
     * Get audit logs within a date range.
     *
     * GET /v1/audit-logs/by-date-range?start={start}&end={end}
     */
    @GetMapping("/by-date-range")
    @Operation(summary = "Listar logs por período", description = "Retorna logs dentro de um intervalo de datas")
    public ResponseEntity<List<AuditLogResponse>> getAuditLogsByDateRange(
            @Parameter(description = "Data/hora inicial (ISO-8601)")
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) Instant start,
            @Parameter(description = "Data/hora final (ISO-8601)")
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) Instant end) {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/audit-logs/by-date-range?start={}end={} - tenant: {}", start, end, tenant);

        List<AuditLog> logs = auditService.getAuditLogsByDateRange(tenant, start, end);
        return ResponseEntity.ok(AuditLogMapper.toResponseList(logs));
    }

    /**
     * Get audit log statistics.
     *
     * GET /v1/audit-logs/stats
     */
    @GetMapping("/stats")
    @Operation(summary = "Estatísticas de auditoria", description = "Retorna estatísticas agregadas dos logs de auditoria")
    public ResponseEntity<AuditLogStatsResponse> getAuditLogStats() {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/audit-logs/stats - tenant: {}", tenant);

        // Get counts by event type
        long injectionCount = auditService.countByTenantAndEventType(tenant, "context_injection");
        long createdCount = auditService.countByTenantAndEventType(tenant, "memory_created");
        long updatedCount = auditService.countByTenantAndEventType(tenant, "memory_updated");
        long deletedCount = auditService.countByTenantAndEventType(tenant, "memory_deleted");
        long errorCount = auditService.countByTenantAndEventType(tenant, "error");

        AuditLogStatsResponse stats = AuditLogStatsResponse.builder()
                .totalEvents(injectionCount + createdCount + updatedCount + deletedCount + errorCount)
                .contextInjections(injectionCount)
                .memoriesCreated(createdCount)
                .memoriesUpdated(updatedCount)
                .memoriesDeleted(deletedCount)
                .errors(errorCount)
                .build();

        return ResponseEntity.ok(stats);
    }

    /**
     * Response containing audit log statistics.
     */
    public record AuditLogStatsResponse(
            Long totalEvents,
            Long contextInjections,
            Long memoriesCreated,
            Long memoriesUpdated,
            Long memoriesDeleted,
            Long errors
    ) {
        public static AuditLogStatsResponseBuilder builder() {
            return new AuditLogStatsResponseBuilder();
        }

        public static class AuditLogStatsResponseBuilder {
            private Long totalEvents;
            private Long contextInjections;
            private Long memoriesCreated;
            private Long memoriesUpdated;
            private Long memoriesDeleted;
            private Long errors;

            public AuditLogStatsResponseBuilder totalEvents(Long totalEvents) {
                this.totalEvents = totalEvents;
                return this;
            }

            public AuditLogStatsResponseBuilder contextInjections(Long contextInjections) {
                this.contextInjections = contextInjections;
                return this;
            }

            public AuditLogStatsResponseBuilder memoriesCreated(Long memoriesCreated) {
                this.memoriesCreated = memoriesCreated;
                return this;
            }

            public AuditLogStatsResponseBuilder memoriesUpdated(Long memoriesUpdated) {
                this.memoriesUpdated = memoriesUpdated;
                return this;
            }

            public AuditLogStatsResponseBuilder memoriesDeleted(Long memoriesDeleted) {
                this.memoriesDeleted = memoriesDeleted;
                return this;
            }

            public AuditLogStatsResponseBuilder errors(Long errors) {
                this.errors = errors;
                return this;
            }

            public AuditLogStatsResponse build() {
                return new AuditLogStatsResponse(
                        totalEvents != null ? totalEvents : 0L,
                        contextInjections != null ? contextInjections : 0L,
                        memoriesCreated != null ? memoriesCreated : 0L,
                        memoriesUpdated != null ? memoriesUpdated : 0L,
                        memoriesDeleted != null ? memoriesDeleted : 0L,
                        errors != null ? errors : 0L
                );
            }
        }
    }
}
