package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.domain.Tenant;
import com.integraltech.brainsentry.service.TenantService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

/**
 * REST controller for tenant management operations.
 *
 * Provides endpoints for managing tenants in the multi-tenant Brain Sentry system.
 */
@Slf4j
@RestController
@RequestMapping("/v1/tenants")
@Tag(name = "Tenants", description = "Gerenciamento de tenants (multi-tenancy)")
public class TenantController {

    private final TenantService tenantService;

    public TenantController(TenantService tenantService) {
        this.tenantService = tenantService;
    }

    /**
     * Get all tenants.
     *
     * GET /v1/tenants
     */
    @GetMapping
    @Operation(summary = "Listar tenants", description = "Retorna todos os tenants do sistema")
    public ResponseEntity<List<TenantResponse>> getTenants(
            @Parameter(description = "Número da página")
            @RequestParam(defaultValue = "0") int page,
            @Parameter(description = "Tamanho da página")
            @RequestParam(defaultValue = "20") int size) {
        log.info("GET /v1/tenants - page: {}, size: {}", page, size);

        Pageable pageable = PageRequest.of(page, size);
        Page<Tenant> tenantsPage = tenantService.getTenants(pageable);

        List<TenantResponse> responses = tenantsPage.getContent().stream()
                .map(this::toResponse)
                .toList();

        return ResponseEntity.ok()
                .header("X-Total-Count", String.valueOf(tenantsPage.getTotalElements()))
                .header("X-Total-Pages", String.valueOf(tenantsPage.getTotalPages()))
                .body(responses);
    }

    /**
     * Get all tenants stats.
     *
     * GET /v1/tenants/stats
     */
    @GetMapping("/stats")
    @Operation(summary = "Estatísticas de todos tenants", description = "Retorna estatísticas de todos os tenants")
    public ResponseEntity<List<TenantStatsResponse>> getAllTenantsStats() {
        log.info("GET /v1/tenants/stats");

        List<Tenant> tenants = tenantService.getTenantsList();
        List<TenantStatsResponse> responses = tenants.stream()
                .map(tenant -> {
                    TenantService.TenantStats stats = tenantService.getTenantStats(tenant.getId());
                    return new TenantStatsResponse(
                            stats.tenantId(),
                            stats.totalMemories(),
                            stats.totalUsers(),
                            stats.totalInjections(),
                            stats.totalRequests(),
                            stats.lastActivityAt()
                    );
                })
                .toList();

        return ResponseEntity.ok(responses);
    }

    /**
     * Get a tenant by ID.
     *
     * GET /v1/tenants/{tenantId}
     */
    @GetMapping("/{tenantId}")
    @Operation(summary = "Obter tenant", description = "Retorna um tenant específico por ID")
    public ResponseEntity<TenantResponse> getTenant(
            @Parameter(description = "ID do tenant")
            @PathVariable String tenantId) {
        log.info("GET /v1/tenants/{}", tenantId);

        Tenant tenant = tenantService.getTenant(tenantId);
        return ResponseEntity.ok(toResponse(tenant));
    }

    /**
     * Create a new tenant.
     *
     * POST /v1/tenants
     */
    @PostMapping
    @Operation(summary = "Criar tenant", description = "Cria um novo tenant no sistema")
    public ResponseEntity<TenantResponse> createTenant(
            @RequestBody CreateTenantRequest request) {
        log.info("POST /v1/tenants - name: {}", request.name());

        Tenant created = tenantService.createTenant(
                request.name(),
                request.slug(),
                request.description(),
                request.maxMemories(),
                request.maxUsers()
        );

        return ResponseEntity.ok(toResponse(created));
    }

    /**
     * Update a tenant.
     *
     * PATCH /v1/tenants/{tenantId}
     */
    @PatchMapping("/{tenantId}")
    @Operation(summary = "Atualizar tenant", description = "Atualiza dados de um tenant")
    public ResponseEntity<TenantResponse> updateTenant(
            @Parameter(description = "ID do tenant")
            @PathVariable String tenantId,
            @RequestBody UpdateTenantRequest request) {
        log.info("PATCH /v1/tenants/{}", tenantId);

        Tenant updated = tenantService.updateTenant(
                tenantId,
                request.name(),
                request.description(),
                request.active(),
                request.maxMemories(),
                request.maxUsers()
        );

        return ResponseEntity.ok(toResponse(updated));
    }

    /**
     * Delete a tenant.
     *
     * DELETE /v1/tenants/{tenantId}
     */
    @DeleteMapping("/{tenantId}")
    @Operation(summary = "Deletar tenant", description = "Remove um tenant do sistema")
    public ResponseEntity<Void> deleteTenant(
            @Parameter(description = "ID do tenant")
            @PathVariable String tenantId) {
        log.info("DELETE /v1/tenants/{}", tenantId);

        tenantService.deleteTenant(tenantId);
        return ResponseEntity.noContent().build();
    }

    /**
     * Get tenant statistics.
     *
     * GET /v1/tenants/{tenantId}/stats
     */
    @GetMapping("/{tenantId}/stats")
    @Operation(summary = "Estatísticas do tenant", description = "Retorna estatísticas de uso do tenant")
    public ResponseEntity<TenantStatsResponse> getTenantStats(
            @Parameter(description = "ID do tenant")
            @PathVariable String tenantId) {
        log.info("GET /v1/tenants/{}/stats", tenantId);

        TenantService.TenantStats stats = tenantService.getTenantStats(tenantId);
        return ResponseEntity.ok(new TenantStatsResponse(
                stats.tenantId(),
                stats.totalMemories(),
                stats.totalUsers(),
                stats.totalInjections(),
                stats.totalRequests(),
                stats.lastActivityAt()
        ));
    }

    /**
     * Get tenant configuration.
     *
     * GET /v1/tenants/{tenantId}/config
     */
    @GetMapping("/{tenantId}/config")
    @Operation(summary = "Configuração do tenant", description = "Retorna a configuração de um tenant")
    public ResponseEntity<Map<String, Object>> getTenantConfig(
            @Parameter(description = "ID do tenant")
            @PathVariable String tenantId) {
        log.info("GET /v1/tenants/{}/config", tenantId);

        Map<String, Object> config = tenantService.getTenantConfig(tenantId);
        return ResponseEntity.ok(config);
    }

    /**
     * Update tenant configuration.
     *
     * PUT /v1/tenants/{tenantId}/config
     */
    @PutMapping("/{tenantId}/config")
    @Operation(summary = "Atualizar configuração", description = "Atualiza a configuração de um tenant")
    public ResponseEntity<Map<String, Object>> updateTenantConfig(
            @Parameter(description = "ID do tenant")
            @PathVariable String tenantId,
            @RequestBody Map<String, Object> config) {
        log.info("PUT /v1/tenants/{}/config", tenantId);

        Map<String, Object> updated = tenantService.updateTenantConfig(tenantId, config);
        return ResponseEntity.ok(updated);
    }

    /**
     * Search tenants.
     *
     * GET /v1/tenants/search
     */
    @GetMapping("/search")
    @Operation(summary = "Buscar tenants", description = "Busca tenants por nome ou slug")
    public ResponseEntity<List<TenantResponse>> searchTenants(
            @Parameter(description = "Termo de busca")
            @RequestParam String query,
            @Parameter(description = "Número da página")
            @RequestParam(defaultValue = "0") int page,
            @Parameter(description = "Tamanho da página")
            @RequestParam(defaultValue = "20") int size) {
        log.info("GET /v1/tenants/search - query: {}", query);

        Pageable pageable = PageRequest.of(page, size);
        Page<Tenant> results = tenantService.searchTenants(query, pageable);

        List<TenantResponse> responses = results.getContent().stream()
                .map(this::toResponse)
                .toList();

        return ResponseEntity.ok()
                .header("X-Total-Count", String.valueOf(results.getTotalElements()))
                .body(responses);
    }

    private TenantResponse toResponse(Tenant tenant) {
        return new TenantResponse(
                tenant.getId(),
                tenant.getName(),
                tenant.getSlug(),
                tenant.getDescription(),
                tenant.getActive(),
                tenant.getMaxMemories(),
                tenant.getMaxUsers(),
                tenant.getCreatedAt(),
                tenant.getUpdatedAt()
        );
    }

    // ==================== DTOs ====================

    public record TenantResponse(
            String id,
            String name,
            String slug,
            String description,
            Boolean active,
            Integer maxMemories,
            Integer maxUsers,
            java.time.Instant createdAt,
            java.time.Instant updatedAt
    ) {}

    public record CreateTenantRequest(
            String name,
            String slug,
            String description,
            Integer maxMemories,
            Integer maxUsers
    ) {}

    public record UpdateTenantRequest(
            String name,
            String description,
            Boolean active,
            Integer maxMemories,
            Integer maxUsers
    ) {}

    public record TenantStatsResponse(
            String tenantId,
            Long totalMemories,
            Long totalUsers,
            Long totalInjections,
            Long totalRequests,
            java.time.Instant lastActivityAt
    ) {}
}
