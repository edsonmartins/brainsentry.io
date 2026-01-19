package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.MemoryVersion;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.repository.MemoryVersionJpaRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * Service for managing memory versions.
 *
 * Provides version history, comparison, and rollback capabilities
 * for memory entries.
 */
@Slf4j
@Service
@RequiredArgsConstructor
@ConditionalOnProperty(name = "features.versioning.enabled", havingValue = "true", matchIfMissing = true)
public class VersionService {

    private final MemoryVersionJpaRepository versionRepository;
    private final MemoryJpaRepository memoryRepository;
    private final AuditService auditService;

    /**
     * Create a new version snapshot of a memory.
     *
     * @param memory the memory to version
     * @param changeType the type of change (create, update, auto_learned, import)
     * @param changedBy user who made the change
     * @param changeReason reason for the change
     * @return created version
     */
    @Transactional
    public MemoryVersion createVersion(Memory memory, String changeType, String changedBy, String changeReason) {
        MemoryVersion version = MemoryVersion.builder()
                .id(UUID.randomUUID().toString())
                .memoryId(memory.getId())
                .version(memory.getVersion())
                .content(memory.getContent())
                .summary(memory.getSummary())
                .category(memory.getCategory())
                .importance(memory.getImportance())
                .metadata(memory.getMetadata())
                .tags(memory.getTags())
                .codeExample(memory.getCodeExample())
                .changedBy(changedBy)
                .changeReason(changeReason)
                .changeType(changeType)
                .createdAt(Instant.now())
                .tenantId(memory.getTenantId())
                .build();

        versionRepository.save(version);
        log.debug("Version {} created for memory: {}", version.getVersion(), memory.getId());

        return version;
    }

    /**
     * Create a version when a memory is created.
     */
    @Transactional
    public MemoryVersion createInitialVersion(Memory memory, String createdBy) {
        return createVersion(memory, "create", createdBy, "Initial version");
    }

    /**
     * Create a version when a memory is updated.
     */
    @Transactional
    public MemoryVersion createUpdateVersion(Memory memory, String updatedBy, String changeReason) {
        return createVersion(memory, "update", updatedBy, changeReason);
    }

    /**
     * Get all versions for a specific memory.
     *
     * @param memoryId the memory ID
     * @return list of versions ordered by version descending
     */
    @Transactional(readOnly = true)
    public List<MemoryVersion> getVersionsForMemory(String memoryId) {
        return versionRepository.findByMemoryIdOrderByVersionDesc(memoryId);
    }

    /**
     * Get all versions for a specific memory by tenant.
     *
     * @param memoryId the memory ID
     * @param tenantId the tenant ID
     * @return list of versions ordered by version descending
     */
    @Transactional(readOnly = true)
    public List<MemoryVersion> getVersionsForMemory(String memoryId, String tenantId) {
        return versionRepository.findByMemoryIdAndTenantId(memoryId, tenantId);
    }

    /**
     * Get a specific version of a memory.
     *
     * @param memoryId the memory ID
     * @param version the version number
     * @return optional version
     */
    @Transactional(readOnly = true)
    public Optional<MemoryVersion> getVersion(String memoryId, Integer version) {
        return versionRepository.findByMemoryIdAndVersion(memoryId, version);
    }

    /**
     * Get the latest version of a memory.
     *
     * @param memoryId the memory ID
     * @return optional latest version
     */
    @Transactional(readOnly = true)
    public Optional<MemoryVersion> getLatestVersion(String memoryId) {
        return versionRepository.findLatestVersion(memoryId);
    }

    /**
     * Compare two versions of a memory.
     *
     * @param memoryId the memory ID
     * @param fromVersion starting version
     * @param toVersion ending version
     * @return version comparison result
     */
    @Transactional(readOnly = true)
    public VersionComparison compareVersions(String memoryId, Integer fromVersion, Integer toVersion) {
        Optional<MemoryVersion> from = getVersion(memoryId, fromVersion);
        Optional<MemoryVersion> to = getVersion(memoryId, toVersion);

        if (from.isEmpty() || to.isEmpty()) {
            throw new IllegalArgumentException("One or both versions not found");
        }

        MemoryVersion fromVer = from.get();
        MemoryVersion toVer = to.get();

        return new VersionComparison(
                fromVer,
                toVer,
                compareFields(fromVer, toVer)
        );
    }

    /**
     * Rollback a memory to a specific version.
     *
     * @param memoryId the memory ID
     * @param version the version to rollback to
     * @param userId user performing the rollback
     * @return updated memory
     */
    @Transactional
    public Memory rollbackToVersion(String memoryId, Integer version, String userId) {
        Memory memory = memoryRepository.findById(memoryId)
                .orElseThrow(() -> new IllegalArgumentException("Memory not found: " + memoryId));

        MemoryVersion targetVersion = versionRepository.findByMemoryIdAndVersion(memoryId, version)
                .orElseThrow(() -> new IllegalArgumentException("Version not found: " + version));

        // Create version of current state before rollback
        createVersion(memory, "rollback", userId,
                "Rollback from version " + memory.getVersion() + " to " + version);

        // Restore content from target version
        memory.setContent(targetVersion.getContent());
        memory.setSummary(targetVersion.getSummary());
        memory.setCategory(targetVersion.getCategory());
        memory.setImportance(targetVersion.getImportance());
        memory.setMetadata(targetVersion.getMetadata());
        memory.setTags(targetVersion.getTags());
        memory.setCodeExample(targetVersion.getCodeExample());
        memory.setVersion(version + 1); // New version after rollback
        memory.setUpdatedAt(Instant.now());

        Memory saved = memoryRepository.save(memory);
        log.info("Rolled back memory: {} to version: {}", memoryId, version);

        // Log the rollback
        auditService.logMemoryUpdated(memoryId, userId, memory.getTenantId());

        return saved;
    }

    /**
     * Delete all versions for a memory.
     *
     * @param memoryId the memory ID
     */
    @Transactional
    public void deleteAllVersionsForMemory(String memoryId) {
        versionRepository.deleteByMemoryId(memoryId);
        log.debug("All versions deleted for memory: {}", memoryId);
    }

    /**
     * Get version count for a memory.
     *
     * @param memoryId the memory ID
     * @return number of versions
     */
    @Transactional(readOnly = true)
    public long getVersionCount(String memoryId) {
        return versionRepository.countByMemoryId(memoryId);
    }

    /**
     * Get versions by change type for a tenant.
     *
     * @param changeType the change type
     * @param tenantId the tenant ID
     * @return list of versions
     */
    @Transactional(readOnly = true)
    public List<MemoryVersion> getVersionsByChangeType(String changeType, String tenantId) {
        return versionRepository.findByChangeTypeAndTenantId(changeType, tenantId);
    }

    /**
     * Get versions created by a specific user.
     *
     * @param changedBy the user who made changes
     * @param tenantId the tenant ID
     * @return list of versions
     */
    @Transactional(readOnly = true)
    public List<MemoryVersion> getVersionsByUser(String changedBy, String tenantId) {
        return versionRepository.findByChangedByAndTenantId(changedBy, tenantId);
    }

    /**
     * Compare fields between two versions.
     */
    private List<FieldChange> compareFields(MemoryVersion from, MemoryVersion to) {
        return List.of(
                compareField("content", from.getContent(), to.getContent()),
                compareField("summary", from.getSummary(), to.getSummary()),
                compareField("category", from.getCategory(), to.getCategory()),
                compareField("importance", from.getImportance(), to.getImportance()),
                compareField("tags", from.getTags(), to.getTags()),
                compareField("codeExample", from.getCodeExample(), to.getCodeExample())
        ).stream()
                .filter(fc -> fc.changed())
                .collect(Collectors.toList());
    }

    private <T> FieldChange compareField(String fieldName, T fromValue, T toValue) {
        boolean changed = !java.util.Objects.equals(fromValue, toValue);
        return new FieldChange(fieldName, fromValue, toValue, changed);
    }

    /**
     * Record value: Result of comparing two versions.
     */
    public record VersionComparison(
            MemoryVersion fromVersion,
            MemoryVersion toVersion,
            List<FieldChange> changes
    ) {
        public boolean hasChanges() {
            return !changes.isEmpty();
        }
    }

    /**
     * Record value: A field change between versions.
     */
    public record FieldChange(
            String fieldName,
            Object oldValue,
            Object newValue,
            boolean changed
    ) {}
}
