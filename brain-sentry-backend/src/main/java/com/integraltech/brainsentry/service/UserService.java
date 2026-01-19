package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.User;
import com.integraltech.brainsentry.repository.AuditLogJpaRepository;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.repository.UserJpaRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

/**
 * Service for user management operations.
 *
 * Provides CRUD operations for users including password management,
 * role management, and statistics.
 */
@Service
public class UserService {

    private static final Logger log = LoggerFactory.getLogger(UserService.class);

    private final UserJpaRepository userRepo;
    private final MemoryJpaRepository memoryRepo;
    private final AuditLogJpaRepository auditLogRepo;
    private final BCryptPasswordEncoder passwordEncoder;

    public UserService(UserJpaRepository userRepo,
                       MemoryJpaRepository memoryRepo,
                       AuditLogJpaRepository auditLogRepo,
                       BCryptPasswordEncoder passwordEncoder) {
        this.userRepo = userRepo;
        this.memoryRepo = memoryRepo;
        this.auditLogRepo = auditLogRepo;
        this.passwordEncoder = passwordEncoder;
    }

    /**
     * Get all users for a tenant.
     *
     * @param tenantId the tenant ID
     * @param pageable pagination parameters
     * @return page of users
     */
    @Transactional(readOnly = true)
    public Page<User> getUsers(String tenantId, Pageable pageable) {
        log.debug("Getting users for tenant: {}", tenantId);
        return userRepo.findByTenantId(tenantId, pageable);
    }

    /**
     * Get all users for a tenant.
     *
     * @param tenantId the tenant ID
     * @return list of users
     */
    @Transactional(readOnly = true)
    public List<User> getUsersList(String tenantId) {
        log.debug("Getting users list for tenant: {}", tenantId);
        return userRepo.findByTenantId(tenantId);
    }

    /**
     * Get a user by ID.
     *
     * @param userId the user ID
     * @param tenantId the tenant ID (for security)
     * @return the user
     * @throws IllegalArgumentException if user not found or doesn't belong to tenant
     */
    @Transactional(readOnly = true)
    public User getUser(String userId, String tenantId) {
        User user = userRepo.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("User not found: " + userId));

        if (!user.getTenantId().equals(tenantId)) {
            throw new IllegalArgumentException("User does not belong to tenant");
        }

        return user;
    }

    /**
     * Find user by email.
     *
     * @param email the email
     * @return optional user
     */
    @Transactional(readOnly = true)
    public Optional<User> findByEmail(String email) {
        return userRepo.findByEmail(email);
    }

    /**
     * Create a new user.
     *
     * @param email the email
     * @param name the name
     * @param password the plain text password
     * @param tenantId the tenant ID
     * @param roles the roles
     * @return the created user
     * @throws IllegalArgumentException if email already exists
     */
    @Transactional
    public User createUser(String email, String name, String password, String tenantId, List<String> roles) {
        log.info("Creating user: {} for tenant: {}", email, tenantId);

        // Check if email already exists
        if (userRepo.findByEmail(email).isPresent()) {
            throw new IllegalArgumentException("Email already exists: " + email);
        }

        // Check tenant user limit
        long currentCount = userRepo.countByTenantId(tenantId);
        // Could check tenant maxUsers limit here if needed

        String passwordHash = passwordEncoder.encode(password);

        User user = User.builder()
                .email(email.toLowerCase())
                .name(name)
                .passwordHash(passwordHash)
                .tenantId(tenantId)
                .roles(roles != null ? roles : List.of("USER"))
                .active(true)
                .emailVerified(false)
                .build();

        User saved = userRepo.save(user);
        log.info("User created: {}", saved.getId());
        return saved;
    }

    /**
     * Update a user.
     *
     * @param userId the user ID
     * @param tenantId the tenant ID
     * @param name the new name (optional)
     * @param email the new email (optional)
     * @param active the active status (optional)
     * @param roles the new roles (optional)
     * @return the updated user
     */
    @Transactional
    public User updateUser(String userId, String tenantId, String name, String email,
                          Boolean active, List<String> roles) {
        log.info("Updating user: {} for tenant: {}", userId, tenantId);

        User user = getUser(userId, tenantId);

        if (name != null) {
            user.setName(name);
        }

        if (email != null && !email.equals(user.getEmail())) {
            // Check if new email already exists
            if (userRepo.existsByEmailAndTenantIdAndIdNot(email, tenantId, userId)) {
                throw new IllegalArgumentException("Email already exists: " + email);
            }
            user.setEmail(email.toLowerCase());
            // Reset email verification when email changes
            user.setEmailVerified(false);
        }

        if (active != null) {
            user.setActive(active);
        }

        if (roles != null) {
            user.setRoles(roles);
        }

        User updated = userRepo.save(user);
        log.info("User updated: {}", userId);
        return updated;
    }

    /**
     * Update user password.
     *
     * @param userId the user ID
     * @param tenantId the tenant ID
     * @param currentPassword the current password
     * @param newPassword the new password
     * @throws IllegalArgumentException if current password is wrong
     */
    @Transactional
    public void updatePassword(String userId, String tenantId, String currentPassword, String newPassword) {
        log.info("Updating password for user: {}", userId);

        User user = getUser(userId, tenantId);

        if (!passwordEncoder.matches(currentPassword, user.getPasswordHash())) {
            throw new IllegalArgumentException("Current password is incorrect");
        }

        user.setPasswordHash(passwordEncoder.encode(newPassword));
        userRepo.save(user);
        log.info("Password updated for user: {}", userId);
    }

    /**
     * Reset user password (admin operation).
     *
     * @param userId the user ID
     * @param tenantId the tenant ID
     * @param newPassword the new password
     */
    @Transactional
    public void resetPassword(String userId, String tenantId, String newPassword) {
        log.info("Resetting password for user: {}", userId);

        User user = getUser(userId, tenantId);
        user.setPasswordHash(passwordEncoder.encode(newPassword));
        userRepo.save(user);
        log.info("Password reset for user: {}", userId);
    }

    /**
     * Delete a user.
     *
     * @param userId the user ID
     * @param tenantId the tenant ID
     */
    @Transactional
    public void deleteUser(String userId, String tenantId) {
        log.info("Deleting user: {} for tenant: {}", userId, tenantId);

        User user = getUser(userId, tenantId);
        userRepo.delete(user);
        log.info("User deleted: {}", userId);
    }

    /**
     * Update last login timestamp.
     *
     * @param userId the user ID
     */
    @Transactional
    public void updateLastLogin(String userId) {
        User user = userRepo.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("User not found: " + userId));
        user.setLastLoginAt(Instant.now());
        userRepo.save(user);
    }

    /**
     * Get user statistics.
     *
     * @param userId the user ID
     * @param tenantId the tenant ID
     * @return user statistics
     */
    @Transactional(readOnly = true)
    public UserStats getUserStats(String userId, String tenantId) {
        User user = getUser(userId, tenantId);

        long memoriesCreated = memoryRepo.countByCreatedBy(userId);
        long totalInteractions = auditLogRepo.countByUserId(userId);

        return new UserStats(
                userId,
                memoriesCreated,
                totalInteractions,
                user.getLastLoginAt()
        );
    }

    /**
     * Search users by email or name.
     *
     * @param tenantId the tenant ID
     * @param search the search term
     * @param pageable pagination parameters
     * @return page of matching users
     */
    @Transactional(readOnly = true)
    public Page<User> searchUsers(String tenantId, String search, Pageable pageable) {
        return userRepo.searchByTenantId(tenantId, search, pageable);
    }

    /**
     * Get active users for a tenant.
     *
     * @param tenantId the tenant ID
     * @return list of active users
     */
    @Transactional(readOnly = true)
    public List<User> getActiveUsers(String tenantId) {
        return userRepo.findByTenantIdAndActive(tenantId, true);
    }

    /**
     * Count users by tenant.
     *
     * @param tenantId the tenant ID
     * @return count of users
     */
    @Transactional(readOnly = true)
    public long countByTenant(String tenantId) {
        return userRepo.countByTenantId(tenantId);
    }

    /**
     * Count active users by tenant.
     *
     * @param tenantId the tenant ID
     * @return count of active users
     */
    @Transactional(readOnly = true)
    public long countActiveByTenant(String tenantId) {
        return userRepo.countByTenantIdAndActive(tenantId, true);
    }

    /**
     * Get recently active users (logged in after date).
     *
     * @param tenantId the tenant ID
     * @param since the date threshold
     * @return list of recently active users
     */
    @Transactional(readOnly = true)
    public List<User> getRecentlyActiveUsers(String tenantId, Instant since) {
        return userRepo.findByTenantIdAndLastLoginAtAfter(tenantId, since);
    }

    /**
     * Verify email.
     *
     * @param userId the user ID
     * @param tenantId the tenant ID
     */
    @Transactional
    public void verifyEmail(String userId, String tenantId) {
        User user = getUser(userId, tenantId);
        user.setEmailVerified(true);
        userRepo.save(user);
        log.info("Email verified for user: {}", userId);
    }

    /**
     * User statistics record.
     */
    public record UserStats(
            String userId,
            Long memoriesCreated,
            Long totalInteractions,
            Instant lastActiveAt
    ) {}
}
