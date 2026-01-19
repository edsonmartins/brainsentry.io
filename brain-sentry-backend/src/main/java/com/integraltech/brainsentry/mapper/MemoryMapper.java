package com.integraltech.brainsentry.mapper;

import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.dto.request.UpdateMemoryRequest;
import com.integraltech.brainsentry.dto.response.MemoryResponse;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

/**
 * Mapper for Memory entity and DTOs.
 */
@Component
public class MemoryMapper {

    /**
     * Convert Memory entity to MemoryResponse DTO.
     */
    public MemoryResponse toResponse(Memory memory) {
        if (memory == null) {
            return null;
        }

        return MemoryResponse.builder()
            .id(memory.getId())
            .content(memory.getContent())
            .summary(memory.getSummary())
            .category(memory.getCategory())
            .importance(memory.getImportance())
            .validationStatus(memory.getValidationStatus())
            .metadata(memory.getMetadata())
            .tags(memory.getTags())
            .sourceType(memory.getSourceType())
            .sourceReference(memory.getSourceReference())
            .createdBy(memory.getCreatedBy())
            .tenantId(memory.getTenantId())
            .createdAt(memory.getCreatedAt())
            .updatedAt(memory.getUpdatedAt())
            .lastAccessedAt(memory.getLastAccessedAt())
            .version(memory.getVersion())
            .accessCount(memory.getAccessCount())
            .injectionCount(memory.getInjectionCount())
            .helpfulCount(memory.getHelpfulCount())
            .notHelpfulCount(memory.getNotHelpfulCount())
            .helpfulnessRate(memory.getHelpfulnessRate())
            .relevanceScore(memory.getRelevanceScore())
            .codeExample(memory.getCodeExample())
            .programmingLanguage(memory.getProgrammingLanguage())
            .build();
    }

    /**
     * Convert a list of Memory entities to MemoryResponse DTOs.
     */
    public List<MemoryResponse> toResponseList(List<Memory> memories) {
        if (memories == null) {
            return List.of();
        }
        return memories.stream()
            .map(this::toResponse)
            .collect(Collectors.toList());
    }

    /**
     * Convert CreateMemoryRequest DTO to Memory entity.
     */
    public Memory toEntity(CreateMemoryRequest request) {
        if (request == null) {
            return null;
        }

        return Memory.builder()
            .content(request.getContent())
            .summary(request.getSummary())
            .category(request.getCategory())
            .importance(request.getImportance())
            .tags(request.getTags())
            .metadata(request.getMetadata())
            .sourceType(request.getSourceType())
            .sourceReference(request.getSourceReference())
            .createdBy(request.getCreatedBy())
            .tenantId(request.getTenantId())
            .codeExample(request.getCodeExample())
            .programmingLanguage(request.getProgrammingLanguage())
            .build();
    }

    /**
     * Update Memory entity from UpdateMemoryRequest DTO.
     * Only updates non-null fields from the request.
     */
    public void updateEntityFromRequest(UpdateMemoryRequest request, Memory memory) {
        if (request == null || memory == null) {
            return;
        }

        if (request.getContent() != null) {
            memory.setContent(request.getContent());
        }
        if (request.getSummary() != null) {
            memory.setSummary(request.getSummary());
        }
        if (request.getCategory() != null) {
            memory.setCategory(request.getCategory());
        }
        if (request.getImportance() != null) {
            memory.setImportance(request.getImportance());
        }
        if (request.getTags() != null) {
            memory.setTags(request.getTags());
        }
        if (request.getMetadata() != null) {
            memory.setMetadata(request.getMetadata());
        }
        if (request.getCodeExample() != null) {
            memory.setCodeExample(request.getCodeExample());
        }
        if (request.getProgrammingLanguage() != null) {
            memory.setProgrammingLanguage(request.getProgrammingLanguage());
        }
    }
}
