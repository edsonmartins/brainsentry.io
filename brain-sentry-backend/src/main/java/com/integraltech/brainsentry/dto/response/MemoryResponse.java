package com.integraltech.brainsentry.dto.response;

import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.domain.enums.ValidationStatus;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.List;
import java.util.Map;

/**
 * Response containing memory details.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MemoryResponse {

    private String id;
    private String content;
    private String summary;
    private MemoryCategory category;
    private ImportanceLevel importance;
    private ValidationStatus validationStatus;
    private Map<String, Object> metadata;
    private List<String> tags;
    private String sourceType;
    private String sourceReference;
    private String createdBy;
    private String tenantId;
    private Instant createdAt;
    private Instant updatedAt;
    private Instant lastAccessedAt;
    private Integer version;
    private Integer accessCount;
    private Integer injectionCount;
    private Integer helpfulCount;
    private Integer notHelpfulCount;
    private Double helpfulnessRate;
    private Double relevanceScore;
    private String codeExample;
    private String programmingLanguage;
}
