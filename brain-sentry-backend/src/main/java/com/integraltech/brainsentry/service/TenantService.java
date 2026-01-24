package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.Tenant;
import com.integraltech.brainsentry.repository.TenantJpaRepository;
import com.integraltech.brainsentry.repository.UserJpaRepository;
import com.integraltech.brainsentry.repository.AuditLogJpaRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

/**
 * Service for tenant management operations.
 *
 * Provides CRUD operations for tenants including configuration
 * management, statistics, and multi-tenancy support.
 */
@Service
public class TenantService {

    private static final Logger log = LoggerFactory.getLogger(TenantService.class);

    private final TenantJpaRepository tenantRepo;
    private final UserJpaRepository userRepo;
    private final AuditLogJpaRepository auditLogRepo;

    public TenantService(TenantJpaRepository tenantRepo,
                        UserJpaRepository userRepo,
                        AuditLogJpaRepository auditLogRepo) {
        this.tenantRepo = tenantRepo;
        this.userRepo = userRepo;
        this.auditLogRepo = auditLogRepo;
    }

    /**
     * Get all tenants.
     *
     * @param pageable pagination parameters
     * @return page of tenants
     */
    @Transactional(readOnly = true)
    public Page<Tenant> getTenants(Pageable pageable) {
        log.debug("Getting all tenants");
        return tenantRepo.findAll(pageable);
    }

    /**
     * Get all tenants as list.
     *
     * @return list of all tenants
     */
    @Transactional(readOnly = true)
    public List<Tenant> getTenantsList() {
        return tenantRepo.findAll();
    }

    /**
     * Get active tenants.
     *
     * @return list of active tenants
     */
    @Transactional(readOnly = true)
    public List<Tenant> getActiveTenants() {
        return tenantRepo.findByActive(true);
    }

    /**
     * Get a tenant by ID.
     *
     * @param tenantId the tenant ID
     * @return the tenant
     * @throws IllegalArgumentException if tenant not found
     */
    @Transactional(readOnly = true)
    public Tenant getTenant(String tenantId) {
        return tenantRepo.findById(tenantId)
                .orElseThrow(() -> new IllegalArgumentException("Tenant not found: " + tenantId));
    }

    /**
     * Get a tenant by slug.
     *
     * @param slug the tenant slug
     * @return optional tenant
     */
    @Transactional(readOnly = true)
    public Optional<Tenant> findBySlug(String slug) {
        return tenantRepo.findBySlug(slug);
    }

    /**
     * Create a new tenant.
     *
     * @param name the tenant name
     * @param slug the tenant slug
     * @param description the description
     * @param maxMemories max memories limit (0 = unlimited)
     * @param maxUsers max users limit (0 = unlimited)
     * @return the created tenant
     * @throws IllegalArgumentException if slug already exists
     */
    @Transactional
    public Tenant createTenant(String name, String slug, String description,
                               Integer maxMemories, Integer maxUsers) {
        log.info("Creating tenant: {} ({})", name, slug);

        // Check if slug already exists
        if (tenantRepo.existsBySlug(slug)) {
            throw new IllegalArgumentException("Slug already exists: " + slug);
        }

        Tenant tenant = Tenant.builder()
                .name(name)
                .slug(slug.toLowerCase())
                .description(description)
                .active(true)
                .maxMemories(maxMemories != null ? maxMemories : 0)
                .maxUsers(maxUsers != null ? maxUsers : 0)
                .settings(createDefaultSettings())
                .build();

        Tenant saved = tenantRepo.save(tenant);
        log.info("Tenant created: {}", saved.getId());
        return saved;
    }

    /**
     * Update a tenant.
     *
     * @param tenantId the tenant ID
     * @param name the new name (optional)
     * @param description the new description (optional)
     * @param active the active status (optional)
     * @param maxMemories new max memories (optional)
     * @param maxUsers new max users (optional)
     * @return the updated tenant
     */
    @Transactional
    public Tenant updateTenant(String tenantId, String name, String description,
                               Boolean active, Integer maxMemories, Integer maxUsers) {
        log.info("Updating tenant: {}", tenantId);

        Tenant tenant = getTenant(tenantId);

        if (name != null) {
            tenant.setName(name);
        }

        if (description != null) {
            tenant.setDescription(description);
        }

        if (active != null) {
            tenant.setActive(active);
        }

        if (maxMemories != null) {
            tenant.setMaxMemories(maxMemories);
        }

        if (maxUsers != null) {
            tenant.setMaxUsers(maxUsers);
        }

        Tenant updated = tenantRepo.save(tenant);
        log.info("Tenant updated: {}", tenantId);
        return updated;
    }

    /**
     * Delete a tenant.
     *
     * @param tenantId the tenant ID
     */
    @Transactional
    public void deleteTenant(String tenantId) {
        log.info("Deleting tenant: {}", tenantId);

        Tenant tenant = getTenant(tenantId);

        // Check if there are users in this tenant
        long userCount = userRepo.countByTenantId(tenantId);
        if (userCount > 0) {
            throw new IllegalStateException("Cannot delete tenant with " + userCount + " users");
        }

        tenantRepo.delete(tenant);
        log.info("Tenant deleted: {}", tenantId);
    }

    /**
     * Get tenant statistics.
     *
     * @param tenantId the tenant ID
     * @return tenant statistics
     */
    @Transactional(readOnly = true)
    public TenantStats getTenantStats(String tenantId) {
        Tenant tenant = getTenant(tenantId);

        long totalUsers = userRepo.countByTenantId(tenantId);
        long activeUsers = userRepo.countByTenantIdAndActive(tenantId, true);

        // Get recent activity timestamp from audit logs
        Instant lastActivityAt = auditLogRepo.findFirstByTenantIdOrderByTimestampDesc(tenantId)
                .map(log -> log.getTimestamp())
                .orElse(tenant.getCreatedAt());

        // For totalMemories and totalRequests, we'd need to query with tenant filter
        // Since Hibernate filters by tenant automatically, we can get these from context
        // For now, returning 0 as these would require bypassing the tenant filter

        return new TenantStats(
                tenantId,
                0L, // totalMemories - would need native query
                totalUsers,
                0L, // totalInjections - would need native query
                0L, // totalRequests - would need native query
                lastActivityAt
        );
    }


    /**
     * Get tenant configuration.
     *
     * @param tenantId the tenant ID
     * @return configuration map
     */
    @Transactional(readOnly = true)
    public Map<String, Object> getTenantConfig(String tenantId) {
        Tenant tenant = getTenant(tenantId);

        Map<String, Object> config = new HashMap<>(tenant.getSettings());
        config.put("tenantId", tenantId);
        config.put("name", tenant.getName());
        config.put("slug", tenant.getSlug());
        config.put("maxMemories", tenant.getMaxMemories());
        config.put("maxUsers", tenant.getMaxUsers());
        config.put("active", tenant.getActive());

        return config;
    }

    /**
     * Update tenant configuration.
     *
     * @param tenantId the tenant ID
     * @param config the configuration updates
     * @return updated configuration map
     */
    @Transactional
    public Map<String, Object> updateTenantConfig(String tenantId, Map<String, Object> config) {
        log.info("Updating config for tenant: {}", tenantId);

        Tenant tenant = getTenant(tenantId);

        // Merge settings
        Map<String, Object> currentSettings = new HashMap<>(tenant.getSettings());
        currentSettings.putAll(config);
        tenant.setSettings(currentSettings);

        tenantRepo.save(tenant);

        return getTenantConfig(tenantId);
    }

    /**
     * Search tenants by name or slug.
     *
     * @param search the search term
     * @param pageable pagination parameters
     * @return page of matching tenants
     */
    @Transactional(readOnly = true)
    public Page<Tenant> searchTenants(String search, Pageable pageable) {
        return tenantRepo.search(search, pageable);
    }

    /**
     * Activate a tenant.
     *
     * @param tenantId the tenant ID
     */
    @Transactional
    public void activateTenant(String tenantId) {
        Tenant tenant = getTenant(tenantId);
        tenant.setActive(true);
        tenantRepo.save(tenant);
        log.info("Tenant activated: {}", tenantId);
    }

    /**
     * Deactivate a tenant.
     *
     * @param tenantId the tenant ID
     */
    @Transactional
    public void deactivateTenant(String tenantId) {
        Tenant tenant = getTenant(tenantId);
        tenant.setActive(false);
        tenantRepo.save(tenant);
        log.info("Tenant deactivated: {}", tenantId);
    }

    /**
     * Count active tenants.
     *
     * @return count of active tenants
     */
    @Transactional(readOnly = true)
    public long countActive() {
        return tenantRepo.countByActive(true);
    }

    /**
     * Check if tenant can create more users.
     *
     * @param tenantId the tenant ID
     * @return true if under limit or unlimited
     */
    @Transactional(readOnly = true)
    public boolean canCreateUser(String tenantId) {
        Tenant tenant = getTenant(tenantId);
        if (!tenant.getActive()) {
            return false;
        }
        if (tenant.getMaxUsers() == 0) {
            return true; // unlimited
        }
        long currentCount = userRepo.countByTenantId(tenantId);
        return currentCount < tenant.getMaxUsers();
    }

    /**
     * Create default settings for a new tenant.
     *
     * @return default settings map
     */
    private Map<String, Object> createDefaultSettings() {
        Map<String, Object> settings = new HashMap<>();
        settings.put("features", List.of("vector_search", "relationship_tracking", "audit_logging"));
        settings.put("embeddingModel", "all-MiniLM-L6-v2");
        settings.put("maxContextMemories", 10);
        settings.put("enableAutoCleanup", false);
        settings.put("cleanupAfterDays", 90);
        return settings;
    }

    /**
     * Tenant statistics record.
     */
    public record TenantStats(
            String tenantId,
            Long totalMemories,
            Long totalUsers,
            Long totalInjections,
            Long totalRequests,
            Instant lastActivityAt
    ) {}
}
