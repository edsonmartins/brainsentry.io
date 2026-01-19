package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.Tenant;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

/**
 * JPA repository for Tenant entity.
 */
@Repository
public interface TenantJpaRepository extends JpaRepository<Tenant, String> {

    /**
     * Find tenant by slug.
     */
    Optional<Tenant> findBySlug(String slug);

    /**
     * Find tenants by active status.
     */
    List<Tenant> findByActive(Boolean active);

    /**
     * Find tenants by active status with pagination.
     */
    Page<Tenant> findByActive(Boolean active, Pageable pageable);

    /**
     * Find tenants by name containing (case-insensitive).
     */
    List<Tenant> findByNameContainingIgnoreCase(String name);

    /**
     * Find tenants by slug containing (case-insensitive).
     */
    List<Tenant> findBySlugContainingIgnoreCase(String slug);

    /**
     * Find tenants by name or slug containing (case-insensitive).
     */
    @Query("SELECT t FROM Tenant t WHERE LOWER(t.name) LIKE LOWER(CONCAT('%', :search, '%')) OR " +
           "LOWER(t.slug) LIKE LOWER(CONCAT('%', :search, '%'))")
    Page<Tenant> search(@Param("search") String search, Pageable pageable);

    /**
     * Count active tenants.
     */
    long countByActive(Boolean active);

    /**
     * Check if slug exists (excluding a specific tenant ID).
     */
    @Query("SELECT COUNT(t) > 0 FROM Tenant t WHERE t.slug = :slug AND t.id != :excludeId")
    boolean existsBySlugAndIdNot(@Param("slug") String slug, @Param("excludeId") String excludeId);

    /**
     * Check if slug exists.
     */
    boolean existsBySlug(String slug);
}
