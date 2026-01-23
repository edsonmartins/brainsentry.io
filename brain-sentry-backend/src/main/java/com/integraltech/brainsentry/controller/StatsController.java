package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.dto.response.StatsResponse;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

/**
 * REST controller for system statistics.
 *
 * Provides metrics and analytics about the Brain Sentry system.
 * All operations automatically filtered by current tenant via Hibernate 6 @TenantId.
 */
@Slf4j
@RestController
@RequestMapping("/v1/stats")
@RequiredArgsConstructor
public class StatsController {

    private final MemoryJpaRepository memoryJpaRepository;

    /**
     * Get system overview statistics.
     * GET /api/v1/stats/overview
     *
     * Note: Tenant is automatically extracted from X-Tenant-ID header.
     */
    @GetMapping("/overview")
    public ResponseEntity<StatsResponse> getOverview() {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/stats/overview - tenant: {}", tenant);

        long totalMemories = memoryJpaRepository.count();

        Map<String, Long> memoriesByCategory = new HashMap<>();
        memoriesByCategory.put("DECISION", memoryJpaRepository.countByCategory(MemoryCategory.DECISION));
        memoriesByCategory.put("PATTERN", memoryJpaRepository.countByCategory(MemoryCategory.PATTERN));
        memoriesByCategory.put("ANTIPATTERN", memoryJpaRepository.countByCategory(MemoryCategory.ANTIPATTERN));
        memoriesByCategory.put("BUG", memoryJpaRepository.countByCategory(MemoryCategory.BUG));
        memoriesByCategory.put("OPTIMIZATION", memoryJpaRepository.countByCategory(MemoryCategory.OPTIMIZATION));
        memoriesByCategory.put("DOMAIN", memoryJpaRepository.countByCategory(MemoryCategory.DOMAIN));
        memoriesByCategory.put("INTEGRATION", memoryJpaRepository.countByCategory(MemoryCategory.INTEGRATION));

        Map<String, Long> memoriesByImportance = Map.of(
            "CRITICAL", memoryJpaRepository.countByImportance(ImportanceLevel.CRITICAL),
            "IMPORTANT", memoryJpaRepository.countByImportance(ImportanceLevel.IMPORTANT),
            "MINOR", memoryJpaRepository.countByImportance(ImportanceLevel.MINOR)
        );

        StatsResponse response = StatsResponse.builder()
            .totalMemories(totalMemories)
            .memoriesByCategory(memoriesByCategory)
            .memoriesByImportance(memoriesByImportance)
            .requestsToday(0L)  // TODO: Implement from audit logs
            .injectionRate(0.0)  // TODO: Implement from audit logs
            .avgLatencyMs(0.0)   // TODO: Implement from audit logs
            .helpfulnessRate(0.0) // TODO: Implement from feedback
            .totalInjections(0L)  // TODO: Implement from audit logs
            .activeMemories24h(0L) // TODO: Implement from audit logs
            .build();

        return ResponseEntity.ok(response);
    }

    /**
     * Health check endpoint.
     * GET /api/v1/stats/health
     */
    @GetMapping("/health")
    public ResponseEntity<Map<String, Object>> health() {
        Map<String, Object> health = new HashMap<>();
        health.put("status", "UP");
        health.put("timestamp", System.currentTimeMillis());
        health.put("service", "brain-sentry");
        health.put("tenant", TenantContext.getTenantId());
        return ResponseEntity.ok(health);
    }
}
