package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.domain.User;
import com.integraltech.brainsentry.service.UserService;
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

/**
 * REST controller for user management operations.
 *
 * Provides endpoints for managing users in the Brain Sentry system.
 */
@Slf4j
@RestController
@RequestMapping("/v1/users")
@Tag(name = "Users", description = "Gerenciamento de usuários")
public class UserController {

    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    /**
     * Get all users.
     *
     * GET /v1/users
     */
    @GetMapping
    @Operation(summary = "Listar usuários", description = "Retorna todos os usuários do sistema")
    public ResponseEntity<List<UserResponse>> getUsers(
            @Parameter(description = "ID do tenant para filtrar")
            @RequestParam(required = false) String tenantId,
            @Parameter(description = "Número da página")
            @RequestParam(defaultValue = "0") int page,
            @Parameter(description = "Tamanho da página")
            @RequestParam(defaultValue = "20") int size) {
        log.info("GET /v1/users - tenant: {}, page: {}, size: {}", tenantId, page, size);

        Pageable pageable = PageRequest.of(page, size);
        Page<User> usersPage;

        if (tenantId != null) {
            usersPage = userService.getUsers(tenantId, pageable);
        } else {
            // If no tenant specified, return first page
            usersPage = userService.getUsers("default", pageable);
        }

        List<UserResponse> responses = usersPage.getContent().stream()
                .map(this::toResponse)
                .toList();

        return ResponseEntity.ok()
                .header("X-Total-Count", String.valueOf(usersPage.getTotalElements()))
                .header("X-Total-Pages", String.valueOf(usersPage.getTotalPages()))
                .body(responses);
    }

    /**
     * Get a user by ID.
     *
     * GET /v1/users/{userId}
     */
    @GetMapping("/{userId}")
    @Operation(summary = "Obter usuário", description = "Retorna um usuário específico por ID")
    public ResponseEntity<UserResponse> getUser(
            @Parameter(description = "ID do usuário")
            @PathVariable String userId,
            @Parameter(description = "ID do tenant")
            @RequestParam(defaultValue = "default") String tenantId) {
        log.info("GET /v1/users/{} - tenant: {}", userId, tenantId);

        User user = userService.getUser(userId, tenantId);
        return ResponseEntity.ok(toResponse(user));
    }

    /**
     * Create a new user.
     *
     * POST /v1/users
     */
    @PostMapping
    @Operation(summary = "Criar usuário", description = "Cria um novo usuário no sistema")
    public ResponseEntity<UserResponse> createUser(
            @RequestBody CreateUserRequest request,
            @Parameter(description = "ID do tenant")
            @RequestParam(defaultValue = "default") String tenantId) {
        log.info("POST /v1/users - email: {}, tenant: {}", request.email(), tenantId);

        User created = userService.createUser(
                request.email(),
                request.name(),
                request.password(),
                tenantId,
                request.roles() != null ? request.roles() : List.of("USER")
        );

        return ResponseEntity.ok(toResponse(created));
    }

    /**
     * Update a user.
     *
     * PATCH /v1/users/{userId}
     */
    @PatchMapping("/{userId}")
    @Operation(summary = "Atualizar usuário", description = "Atualiza dados de um usuário")
    public ResponseEntity<UserResponse> updateUser(
            @Parameter(description = "ID do usuário")
            @PathVariable String userId,
            @RequestBody UpdateUserRequest request,
            @Parameter(description = "ID do tenant")
            @RequestParam(defaultValue = "default") String tenantId) {
        log.info("PATCH /v1/users/{} - tenant: {}", userId, tenantId);

        User updated = userService.updateUser(
                userId,
                tenantId,
                request.name(),
                request.email(),
                request.active(),
                request.roles()
        );

        return ResponseEntity.ok(toResponse(updated));
    }

    /**
     * Delete a user.
     *
     * DELETE /v1/users/{userId}
     */
    @DeleteMapping("/{userId}")
    @Operation(summary = "Deletar usuário", description = "Remove um usuário do sistema")
    public ResponseEntity<Void> deleteUser(
            @Parameter(description = "ID do usuário")
            @PathVariable String userId,
            @Parameter(description = "ID do tenant")
            @RequestParam(defaultValue = "default") String tenantId) {
        log.info("DELETE /v1/users/{} - tenant: {}", userId, tenantId);

        userService.deleteUser(userId, tenantId);
        return ResponseEntity.noContent().build();
    }

    /**
     * Get user statistics.
     *
     * GET /v1/users/{userId}/stats
     */
    @GetMapping("/{userId}/stats")
    @Operation(summary = "Estatísticas do usuário", description = "Retorna estatísticas de atividade do usuário")
    public ResponseEntity<UserStatsResponse> getUserStats(
            @Parameter(description = "ID do usuário")
            @PathVariable String userId,
            @Parameter(description = "ID do tenant")
            @RequestParam(defaultValue = "default") String tenantId) {
        log.info("GET /v1/users/{}/stats - tenant: {}", userId, tenantId);

        UserService.UserStats stats = userService.getUserStats(userId, tenantId);
        return ResponseEntity.ok(new UserStatsResponse(
                stats.userId(),
                stats.memoriesCreated(),
                stats.totalInteractions(),
                stats.lastActiveAt()
        ));
    }

    /**
     * Search users.
     *
     * GET /v1/users/search
     */
    @GetMapping("/search")
    @Operation(summary = "Buscar usuários", description = "Busca usuários por email ou nome")
    public ResponseEntity<List<UserResponse>> searchUsers(
            @Parameter(description = "Termo de busca")
            @RequestParam String query,
            @Parameter(description = "ID do tenant")
            @RequestParam(defaultValue = "default") String tenantId,
            @Parameter(description = "Número da página")
            @RequestParam(defaultValue = "0") int page,
            @Parameter(description = "Tamanho da página")
            @RequestParam(defaultValue = "20") int size) {
        log.info("GET /v1/users/search - query: {}, tenant: {}", query, tenantId);

        Pageable pageable = PageRequest.of(page, size);
        Page<User> results = userService.searchUsers(tenantId, query, pageable);

        List<UserResponse> responses = results.getContent().stream()
                .map(this::toResponse)
                .toList();

        return ResponseEntity.ok()
                .header("X-Total-Count", String.valueOf(results.getTotalElements()))
                .body(responses);
    }

    private UserResponse toResponse(User user) {
        return new UserResponse(
                user.getId(),
                user.getName(),
                user.getEmail(),
                user.getActive(),
                user.getRoles(),
                user.getCreatedAt(),
                user.getLastLoginAt()
        );
    }

    // ==================== DTOs ====================

    public record UserResponse(
            String id,
            String name,
            String email,
            Boolean active,
            List<String> roles,
            java.time.Instant createdAt,
            java.time.Instant lastLoginAt
    ) {}

    public record CreateUserRequest(
            String name,
            String email,
            String password,
            List<String> roles
    ) {}

    public record UpdateUserRequest(
            String name,
            String email,
            Boolean active,
            List<String> roles
    ) {}

    public record UserStatsResponse(
            String userId,
            Long memoriesCreated,
            Long totalInteractions,
            java.time.Instant lastActiveAt
    ) {}
}
