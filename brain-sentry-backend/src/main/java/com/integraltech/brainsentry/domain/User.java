package com.integraltech.brainsentry.domain;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/**
 * User entity representing a user in the Brain Sentry system.
 *
 * Users can belong to tenants and have roles that determine their
 * access level within the system.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "users", indexes = {
    @Index(name = "idx_user_tenant", columnList = "tenantId"),
    @Index(name = "idx_user_email", columnList = "email"),
    @Index(name = "idx_user_active", columnList = "active")
})
public class User {

    /**
     * Unique identifier for this user.
     */
    @Id
    @Column(length = 100)
    private String id;

    /**
     * User's email address (unique).
     */
    @Column(length = 255, nullable = false, unique = true)
    private String email;

    /**
     * User's display name.
     */
    @Column(length = 255)
    private String name;

    /**
     * Hashed password (BCrypt).
     */
    @Column(length = 255, nullable = false)
    private String passwordHash;

    /**
     * Tenant ID for multi-tenancy support.
     */
    @Column(length = 100, nullable = false)
    private String tenantId;

    /**
     * User roles (comma-separated or separate table).
     * Values: USER, ADMIN, MODERATOR
     */
    @ElementCollection
    @CollectionTable(name = "user_roles", joinColumns = @JoinColumn(name = "user_id"))
    @Column(name = "role")
    @Builder.Default
    private List<String> roles = new ArrayList<>();

    /**
     * Whether the user account is active.
     */
    @Column(nullable = false)
    @Builder.Default
    private Boolean active = true;

    /**
     * When the user was created.
     */
    @Column(nullable = false)
    private Instant createdAt;

    /**
     * When the user last logged in.
     */
    private Instant lastLoginAt;

    /**
     * Email verification status.
     */
    @Column(nullable = false)
    @Builder.Default
    private Boolean emailVerified = false;

    /**
     * Additional user metadata stored as JSON.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String metadata;

    @PrePersist
    protected void onCreate() {
        if (createdAt == null) {
            createdAt = Instant.now();
        }
        if (id == null || id.isEmpty()) {
            id = "user-" + System.currentTimeMillis() + "-" + (int)(Math.random() * 1000);
        }
    }

    // Manual builder method
    public static UserBuilder builder() {
        return new UserBuilder();
    }

    public static class UserBuilder {
        private String id;
        private String email;
        private String name;
        private String passwordHash;
        private String tenantId;
        private List<String> roles = new ArrayList<>();
        private Boolean active = true;
        private Instant createdAt;
        private Instant lastLoginAt;
        private Boolean emailVerified = false;
        private String metadata;

        public UserBuilder id(String id) {
            this.id = id;
            return this;
        }

        public UserBuilder email(String email) {
            this.email = email;
            return this;
        }

        public UserBuilder name(String name) {
            this.name = name;
            return this;
        }

        public UserBuilder passwordHash(String passwordHash) {
            this.passwordHash = passwordHash;
            return this;
        }

        public UserBuilder tenantId(String tenantId) {
            this.tenantId = tenantId;
            return this;
        }

        public UserBuilder roles(List<String> roles) {
            this.roles = roles != null ? roles : new ArrayList<>();
            return this;
        }

        public UserBuilder active(Boolean active) {
            this.active = active;
            return this;
        }

        public UserBuilder createdAt(Instant createdAt) {
            this.createdAt = createdAt;
            return this;
        }

        public UserBuilder lastLoginAt(Instant lastLoginAt) {
            this.lastLoginAt = lastLoginAt;
            return this;
        }

        public UserBuilder emailVerified(Boolean emailVerified) {
            this.emailVerified = emailVerified;
            return this;
        }

        public UserBuilder metadata(String metadata) {
            this.metadata = metadata;
            return this;
        }

        public User build() {
            User user = new User();
            user.id = this.id;
            user.email = this.email;
            user.name = this.name;
            user.passwordHash = this.passwordHash;
            user.tenantId = this.tenantId;
            user.roles = this.roles;
            user.active = this.active;
            user.createdAt = this.createdAt;
            user.lastLoginAt = this.lastLoginAt;
            user.emailVerified = this.emailVerified;
            user.metadata = this.metadata;
            return user;
        }
    }
}
