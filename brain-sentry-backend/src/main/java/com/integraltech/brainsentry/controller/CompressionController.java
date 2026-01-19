package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.dto.request.CompressionRequest;
import com.integraltech.brainsentry.dto.response.CompressedContextResponse;
import com.integraltech.brainsentry.service.ArchitectService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

/**
 * REST controller for context compression operations.
 *
 * Inspired by Confucius Code Agent's Architect agent.
 * Provides endpoints for compressing conversation history
 * and extracting structured summaries.
 */
@Slf4j
@RestController
@RequestMapping("/v1/compression")
@RequiredArgsConstructor
public class CompressionController {

    private final ArchitectService architectService;

    /**
     * Compress conversation context.
     * POST /api/v1/compression/compress
     */
    @PostMapping("/compress")
    public ResponseEntity<CompressedContextResponse> compressContext(
        @Valid @RequestBody CompressionRequest request
    ) {
        log.info("POST /v1/compression/compress - messages: {}", request.getMessages().size());

        CompressedContextResponse response = architectService.compressContext(
            request.getMessages(),
            request.getTokenThreshold()
        );

        return ResponseEntity.ok(response);
    }

    /**
     * Extract structured summary from messages.
     * POST /api/v1/compression/summary
     */
    @PostMapping("/summary")
    public ResponseEntity<CompressedContextResponse.StructuredSummary> extractSummary(
        @Valid @RequestBody CompressionRequest request
    ) {
        log.info("POST /v1/compression/summary - messages: {}", request.getMessages().size());

        CompressedContextResponse.StructuredSummary summary =
            architectService.extractSummary(request.getMessages());

        return ResponseEntity.ok(summary);
    }

    /**
     * Check if compression is needed.
     * POST /api/v1/compression/check
     */
    @PostMapping("/check")
    public ResponseEntity<Boolean> shouldCompress(
        @Valid @RequestBody CompressionRequest request
    ) {
        log.info("POST /v1/compression/check");

        boolean needed = architectService.shouldCompress(
            request.getMessages(),
            request.getTokenThreshold()
        );

        return ResponseEntity.ok(needed);
    }

    /**
     * Identify critical messages that should be preserved.
     * POST /api/v1/compression/critical
     */
    @PostMapping("/critical")
    public ResponseEntity<?> identifyCriticalMessages(
        @Valid @RequestBody CompressionRequest request
    ) {
        log.info("POST /v1/compression/critical");

        var critical = architectService.identifyCriticalMessages(
            request.getMessages(),
            request.getPreserveKeywords()
        );

        return ResponseEntity.ok(critical);
    }

    /**
     * Get compression status for a tenant.
     * GET /api/v1/compression/status
     */
    @GetMapping("/status")
    public ResponseEntity<CompressionStatus> getCompressionStatus(
        @RequestParam(defaultValue = "default") String tenantId
    ) {
        log.info("GET /v1/compression/status - tenant: {}", tenantId);

        TenantContext.setTenantId(tenantId);

        // Return current compression settings and statistics
        CompressionStatus status = CompressionStatus.builder()
            .tenantId(tenantId)
            .compressionEnabled(true)
            .defaultTokenThreshold(100000)
            .defaultPreserveRecent(10)
            .defaultTargetRatio(0.5)
            .totalCompressions(0)
            .totalTokensSaved(0L)
            .build();

        return ResponseEntity.ok(status);
    }

    /**
     * Compression status information.
     */
    @lombok.Data
    @lombok.Builder
    public static class CompressionStatus {
        private String tenantId;
        private Boolean compressionEnabled;
        private Integer defaultTokenThreshold;
        private Integer defaultPreserveRecent;
        private Double defaultTargetRatio;
        private Integer totalCompressions;
        private Long totalTokensSaved;
    }
}
