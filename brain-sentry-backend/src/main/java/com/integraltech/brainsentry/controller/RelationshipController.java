package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.MemoryRelationship;
import com.integraltech.brainsentry.domain.enums.RelationshipType;
import com.integraltech.brainsentry.service.RelationshipService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.constraints.Min;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;

/**
 * REST controller for memory relationship operations.
 *
 * Provides endpoints for managing connections between memories.
 */
@Slf4j
@RestController
@RequestMapping("/v1/relationships")
@RequiredArgsConstructor
@ConditionalOnProperty(name = "features.relationship.enabled", havingValue = "true", matchIfMissing = false)
@Tag(name = "Relationships", description = "Gerenciamento de relacionamentos entre memórias")
public class RelationshipController {

    private final RelationshipService relationshipService;

    /**
     * Create a relationship between two memories.
     *
     * POST /v1/relationships
     */
    @PostMapping
    @Operation(summary = "Criar relacionamento", description = "Cria um novo relacionamento entre duas memórias")
    public ResponseEntity<MemoryRelationship> createRelationship(
            @Parameter(description = "ID da memória de origem")
            @RequestParam String fromMemoryId,
            @Parameter(description = "ID da memória de destino")
            @RequestParam String toMemoryId,
            @Parameter(description = "Tipo de relacionamento")
            @RequestParam RelationshipType type,
            @Parameter(description = "ID do usuário (opcional, usa contexto se não fornecido)")
            @RequestParam(defaultValue = "system") String userId) {
        String tenant = TenantContext.getTenantId();
        log.info("POST /v1/relationships - from: {}, to: {}, type: {}, tenant: {}",
                fromMemoryId, toMemoryId, type, tenant);

        MemoryRelationship relationship = relationshipService.createRelationship(
                fromMemoryId, toMemoryId, type, tenant, userId);
        return ResponseEntity.ok(relationship);
    }

    /**
     * Create bidirectional relationships.
     *
     * POST /v1/relationships/bidirectional
     */
    @PostMapping("/bidirectional")
    @Operation(summary = "Criar relacionamento bidirecional", description = "Cria relacionamentos em ambas as direções")
    public ResponseEntity<List<MemoryRelationship>> createBidirectionalRelationship(
            @Parameter(description = "ID da primeira memória")
            @RequestParam String memoryId1,
            @Parameter(description = "ID da segunda memória")
            @RequestParam String memoryId2,
            @Parameter(description = "Tipo de relacionamento da memória 1 para 2")
            @RequestParam RelationshipType type1,
            @Parameter(description = "Tipo de relacionamento da memória 2 para 1")
            @RequestParam RelationshipType type2,
            @Parameter(description = "ID do usuário")
            @RequestParam(defaultValue = "system") String userId) {
        String tenant = TenantContext.getTenantId();
        log.info("POST /v1/relationships/bidirectional - mem1: {}, mem2: {}, tenant: {}",
                memoryId1, memoryId2, tenant);

        List<MemoryRelationship> relationships = relationshipService.createBidirectionalRelationship(
                memoryId1, memoryId2, type1, type2, tenant, userId);
        return ResponseEntity.ok(relationships);
    }

    /**
     * Get all relationships from a specific memory.
     *
     * GET /v1/relationships/from/{memoryId}
     */
    @GetMapping("/from/{memoryId}")
    @Operation(summary = "Listar relacionamentos de origem", description = "Retorna todos os relacionamentos onde a memória é a origem")
    public ResponseEntity<List<MemoryRelationship>> getRelationshipsFrom(
            @Parameter(description = "ID da memória")
            @PathVariable String memoryId) {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/relationships/from/{} - tenant: {}", memoryId, tenant);

        List<MemoryRelationship> relationships = relationshipService.getRelationshipsFrom(memoryId, tenant);
        return ResponseEntity.ok(relationships);
    }

    /**
     * Get all relationships to a specific memory.
     *
     * GET /v1/relationships/to/{memoryId}
     */
    @GetMapping("/to/{memoryId}")
    @Operation(summary = "Listar relacionamentos de destino", description = "Retorna todos os relacionamentos onde a memória é o destino")
    public ResponseEntity<List<MemoryRelationship>> getRelationshipsTo(
            @Parameter(description = "ID da memória")
            @PathVariable String memoryId) {
        log.info("GET /v1/relationships/to/{}", memoryId);

        List<MemoryRelationship> relationships = relationshipService.getRelationshipsTo(memoryId);
        return ResponseEntity.ok(relationships);
    }

    /**
     * Get a specific relationship between two memories.
     *
     * GET /v1/relationships/between
     */
    @GetMapping("/between")
    @Operation(summary = "Obter relacionamento específico", description = "Retorna o relacionamento entre duas memórias")
    public ResponseEntity<MemoryRelationship> getRelationship(
            @Parameter(description = "ID da memória de origem")
            @RequestParam String fromMemoryId,
            @Parameter(description = "ID da memória de destino")
            @RequestParam String toMemoryId) {
        log.info("GET /v1/relationships/between?from={}&to={}", fromMemoryId, toMemoryId);

        return relationshipService.getRelationship(fromMemoryId, toMemoryId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    /**
     * Find related memories based on strength threshold.
     *
     * GET /v1/relationships/{memoryId}/related
     */
    @GetMapping("/{memoryId}/related")
    @Operation(summary = "Encontrar memórias relacionadas", description = "Retorna memórias relacionadas com força mínima especificada")
    public ResponseEntity<List<RelationshipService.RelatedMemory>> findRelatedMemories(
            @Parameter(description = "ID da memória")
            @PathVariable String memoryId,
            @Parameter(description = "Força mínima (0.0 a 1.0)")
            @RequestParam(defaultValue = "0.5") double minStrength) {
        String tenant = TenantContext.getTenantId();
        log.info("GET /v1/relationships/{}/related?minStrength={} - tenant: {}",
                memoryId, minStrength, tenant);

        List<RelationshipService.RelatedMemory> related =
                relationshipService.findRelatedMemories(memoryId, tenant, minStrength);
        return ResponseEntity.ok(related);
    }

    /**
     * Update relationship strength.
     *
     * PUT /v1/relationships/{relationshipId}/strength
     */
    @PutMapping("/{relationshipId}/strength")
    @Operation(summary = "Atualizar força do relacionamento", description = "Atualiza a força (peso) de um relacionamento")
    public ResponseEntity<MemoryRelationship> updateStrength(
            @Parameter(description = "ID do relacionamento")
            @PathVariable String relationshipId,
            @Parameter(description = "Nova força (0.0 a 1.0)")
            @RequestParam double strength) {
        log.info("PUT /v1/relationships/{}/strength?strength={}", relationshipId, strength);

        MemoryRelationship updated = relationshipService.updateStrength(relationshipId, strength);
        return ResponseEntity.ok(updated);
    }

    /**
     * Delete a relationship.
     *
     * DELETE /v1/relationships/between
     */
    @DeleteMapping("/between")
    @Operation(summary = "Deletar relacionamento", description = "Remove o relacionamento entre duas memórias")
    public ResponseEntity<Void> deleteRelationship(
            @Parameter(description = "ID da memória de origem")
            @RequestParam String fromMemoryId,
            @Parameter(description = "ID da memória de destino")
            @RequestParam String toMemoryId) {
        String tenant = TenantContext.getTenantId();
        log.info("DELETE /v1/relationships/between?from={}&to={} - tenant: {}",
                fromMemoryId, toMemoryId, tenant);

        boolean deleted = relationshipService.deleteRelationship(fromMemoryId, toMemoryId, tenant);
        return deleted ? ResponseEntity.noContent().build() : ResponseEntity.notFound().build();
    }

    /**
     * Delete all relationships for a memory.
     *
     * DELETE /v1/relationships/{memoryId}
     */
    @DeleteMapping("/{memoryId}")
    @Operation(summary = "Deletar todos os relacionamentos", description = "Remove todos os relacionamentos de uma memória")
    public ResponseEntity<Void> deleteAllRelationshipsForMemory(
            @Parameter(description = "ID da memória")
            @PathVariable String memoryId) {
        log.info("DELETE /v1/relationships/{}", memoryId);

        relationshipService.deleteAllRelationshipsForMemory(memoryId);
        return ResponseEntity.noContent().build();
    }

    /**
     * Suggest relationships based on semantic similarity.
     *
     * POST /v1/relationships/{memoryId}/suggest
     */
    @PostMapping("/{memoryId}/suggest")
    @Operation(summary = "Sugerir relacionamentos", description = "Sugere relacionamentos com base em similaridade semântica")
    public ResponseEntity<List<MemoryRelationship>> suggestRelationships(
            @Parameter(description = "ID da memória")
            @PathVariable String memoryId,
            @Parameter(description = "Limiar de similaridade")
            @RequestParam(defaultValue = "0.7") double threshold) {
        String tenant = TenantContext.getTenantId();
        log.info("POST /v1/relationships/{}/suggest?threshold={} - tenant: {}",
                memoryId, threshold, tenant);

        List<MemoryRelationship> suggestions =
                relationshipService.suggestRelationships(memoryId, tenant, threshold);
        return ResponseEntity.ok(suggestions);
    }
}
