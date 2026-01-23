package com.integraltech.brainsentry.repository;

import com.integraltech.brainsentry.domain.Note;
import com.integraltech.brainsentry.domain.enums.NoteCategory;
import com.integraltech.brainsentry.domain.enums.NoteSeverity;
import com.integraltech.brainsentry.domain.enums.NoteType;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;

/**
 * JPA repository for Note entity.
 *
 * Manages generic notes for insights, patterns, and architectural decisions.
 */
@Repository
public interface NoteJpaRepository extends JpaRepository<Note, String> {

    /**
     * Find all notes for a tenant.
     */
    List<Note> findByTenantId(String tenantId);

    /**
     * Find notes by session ID.
     */
    List<Note> findBySessionId(String sessionId);

    /**
     * Find notes by project ID.
     */
    List<Note> findByProjectId(String projectId);

    /**
     * Find notes by type.
     */
    List<Note> findByTenantIdAndType(String tenantId, NoteType type);

    /**
     * Find notes by category.
     */
    List<Note> findByTenantIdAndCategory(String tenantId, NoteCategory category);

    /**
     * Find notes by severity.
     */
    List<Note> findByTenantIdAndSeverity(String tenantId, NoteSeverity severity);

    /**
     * Find notes by type and category.
     */
    @Query("SELECT n FROM Note n WHERE n.tenantId = :tenantId AND n.type = :type AND n.category = :category")
    List<Note> findByTenantIdAndTypeAndCategory(
        @Param("tenantId") String tenantId,
        @Param("type") NoteType type,
        @Param("category") NoteCategory category
    );

    /**
     * Find shared or generic notes (applicable to multiple projects).
     */
    @Query("SELECT n FROM Note n WHERE n.tenantId = :tenantId AND n.category IN ('SHARED', 'GENERIC')")
    List<Note> findSharedByTenantId(@Param("tenantId") String tenantId);

    /**
     * Find notes with specific keyword.
     */
    @Query("SELECT n FROM Note n JOIN n.keywords k WHERE k = :keyword AND n.tenantId = :tenantId")
    List<Note> findByKeyword(@Param("tenantId") String tenantId, @Param("keyword") String keyword);

    /**
     * Find recently created notes.
     */
    List<Note> findByTenantIdOrderByCreatedAtDesc(String tenantId);

    /**
     * Find most accessed notes.
     */
    @Query("SELECT n FROM Note n WHERE n.tenantId = :tenantId ORDER BY n.accessCount DESC, n.lastAccessedAt DESC")
    List<Note> findMostAccessedByTenantId(@Param("tenantId") String tenantId, Pageable pageable);

    /**
     * Find notes created after a certain date.
     */
    List<Note> findByTenantIdAndCreatedAtAfter(String tenantId, Instant date);

    /**
     * Find notes related to a specific memory.
     */
    @Query("SELECT n FROM Note n JOIN n.relatedMemoryIds m WHERE m = :memoryId AND n.tenantId = :tenantId")
    List<Note> findByRelatedMemory(@Param("memoryId") String memoryId, @Param("tenantId") String tenantId);

    /**
     * Find notes related to another note.
     */
    @Query("SELECT n FROM Note n JOIN n.relatedNoteIds nn WHERE nn = :noteId AND n.tenantId = :tenantId")
    List<Note> findByRelatedNote(@Param("noteId") String noteId, @Param("tenantId") String tenantId);

    /**
     * Full-text search in note title and content.
     */
    @Query(value = "SELECT * FROM notes n WHERE n.tenant_id = :tenantId AND (LOWER(n.title) LIKE LOWER('%' || :query || '%') OR LOWER(n.content) LIKE LOWER('%' || :query || '%'))", nativeQuery = true)
    List<Note> searchByTitleOrContent(@Param("tenantId") String tenantId, @Param("query") String query);

    /**
     * Count notes by type for a tenant.
     */
    Long countByTenantIdAndType(String tenantId, NoteType type);

    /**
     * Count notes by category for a tenant.
     */
    Long countByTenantIdAndCategory(String tenantId, NoteCategory category);

    /**
     * Count notes by severity for a tenant.
     */
    Long countByTenantIdAndSeverity(String tenantId, NoteSeverity severity);
}
