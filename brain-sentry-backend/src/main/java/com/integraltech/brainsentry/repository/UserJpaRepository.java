package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.User;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

/**
 * JPA repository for User entity.
 */
@Repository
public interface UserJpaRepository extends JpaRepository<User, String> {

    /**
     * Find user by email.
     */
    Optional<User> findByEmail(String email);

    /**
     * Find users by tenant ID.
     */
    List<User> findByTenantId(String tenantId);

    /**
     * Find users by tenant ID with pagination.
     */
    Page<User> findByTenantId(String tenantId, Pageable pageable);

    /**
     * Find users by tenant ID and active status.
     */
    List<User> findByTenantIdAndActive(String tenantId, Boolean active);

    /**
     * Find users by tenant ID with pagination and active status.
     */
    Page<User> findByTenantIdAndActive(String tenantId, Boolean active, Pageable pageable);

    /**
     * Count users by tenant ID.
     */
    long countByTenantId(String tenantId);

    /**
     * Count active users by tenant ID.
     */
    long countByTenantIdAndActive(String tenantId, Boolean active);

    /**
     * Find users who logged in after a specific date.
     */
    List<User> findByTenantIdAndLastLoginAtAfter(String tenantId, Instant date);

    /**
     * Find users by email containing (case-insensitive search).
     */
    List<User> findByEmailContainingIgnoreCase(String email);

    /**
     * Find users by name containing (case-insensitive search).
     */
    List<User> findByNameContainingIgnoreCase(String name);

    /**
     * Find users by tenant ID and email/name containing.
     */
    @Query("SELECT u FROM User u WHERE u.tenantId = :tenantId AND " +
           "(LOWER(u.email) LIKE LOWER(CONCAT('%', :search, '%')) OR " +
           "LOWER(u.name) LIKE LOWER(CONCAT('%', :search, '%')))")
    Page<User> searchByTenantId(@Param("tenantId") String tenantId,
                                 @Param("search") String search,
                                 Pageable pageable);

    /**
     * Check if email exists for a tenant (excluding a specific user ID).
     */
    @Query("SELECT COUNT(u) > 0 FROM User u WHERE u.tenantId = :tenantId AND " +
           "u.email = :email AND u.id != :excludeId")
    boolean existsByEmailAndTenantIdAndIdNot(@Param("email") String email,
                                              @Param("tenantId") String tenantId,
                                              @Param("excludeId") String excludeId);
}
