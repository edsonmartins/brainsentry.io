package com.integraltech.brainsentry.domain;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

/**
 * Tenant entity for multi-tenancy support.
 *
 * Each tenant represents an organization or customer
 * with isolated data and configuration.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "tenants", indexes = {
    @Index(name = "idx_tenant_slug", columnList = "slug"),
    @Index(name = "idx_tenant_active", columnList = "active")
})
public class Tenant {

    /**
     * Unique identifier for this tenant.
     */
    @Id
    @Column(length = 100)
    private String id;

    /**
     * Human-readable name of the tenant.
     */
    @Column(length = 255, nullable = false)
    private String name;

    /**
     * URL-friendly identifier (unique).
     */
    @Column(length = 100, nullable = false, unique = true)
    private String slug;

    /**
     * Description of the tenant/organization.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String description;

    /**
     * Whether the tenant is active.
     */
    @Column(nullable = false)
    @Builder.Default
    private Boolean active = true;

    /**
     * Maximum number of memories allowed (0 = unlimited).
     */
    @Builder.Default
    private Integer maxMemories = 0;

    /**
     * Maximum number of users allowed (0 = unlimited).
     */
    @Builder.Default
    private Integer maxUsers = 0;

    /**
     * Tenant configuration stored as JSON.
     * May include feature flags, custom settings, etc.
     */
    @JdbcTypeCode(SqlTypes.JSON)
    @Builder.Default
    private Map<String, Object> settings = new HashMap<>();

    /**
     * When the tenant was created.
     */
    @Column(nullable = false)
    private Instant createdAt;

    /**
     * When the tenant was last updated.
     */
    private Instant updatedAt;

    @PrePersist
    protected void onCreate() {
        if (createdAt == null) {
            createdAt = Instant.now();
        }
        if (updatedAt == null) {
            updatedAt = Instant.now();
        }
        if (id == null || id.isEmpty()) {
            id = "tenant-" + System.currentTimeMillis();
        }
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = Instant.now();
    }

    // Manual builder method
    public static TenantBuilder builder() {
        return new TenantBuilder();
    }

    public static class TenantBuilder {
        private String id;
        private String name;
        private String slug;
        private String description;
        private Boolean active = true;
        private Integer maxMemories = 0;
        private Integer maxUsers = 0;
        private Map<String, Object> settings = new HashMap<>();
        private Instant createdAt;
        private Instant updatedAt;

        public TenantBuilder id(String id) {
            this.id = id;
            return this;
        }

        public TenantBuilder name(String name) {
            this.name = name;
            return this;
        }

        public TenantBuilder slug(String slug) {
            this.slug = slug;
            return this;
        }

        public TenantBuilder description(String description) {
            this.description = description;
            return this;
        }

        public TenantBuilder active(Boolean active) {
            this.active = active;
            return this;
        }

        public TenantBuilder maxMemories(Integer maxMemories) {
            this.maxMemories = maxMemories;
            return this;
        }

        public TenantBuilder maxUsers(Integer maxUsers) {
            this.maxUsers = maxUsers;
            return this;
        }

        public TenantBuilder settings(Map<String, Object> settings) {
            this.settings = settings != null ? settings : new HashMap<>();
            return this;
        }

        public TenantBuilder createdAt(Instant createdAt) {
            this.createdAt = createdAt;
            return this;
        }

        public TenantBuilder updatedAt(Instant updatedAt) {
            this.updatedAt = updatedAt;
            return this;
        }

        public Tenant build() {
            Tenant tenant = new Tenant();
            tenant.id = this.id;
            tenant.name = this.name;
            tenant.slug = this.slug;
            tenant.description = this.description;
            tenant.active = this.active;
            tenant.maxMemories = this.maxMemories;
            tenant.maxUsers = this.maxUsers;
            tenant.settings = this.settings;
            tenant.createdAt = this.createdAt;
            tenant.updatedAt = this.updatedAt;
            return tenant;
        }
    }
}
