package com.integraltech.brainsentry.dto.request;

import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;
import java.util.Map;

/**
 * Request to update an existing memory.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class UpdateMemoryRequest {

    /**
     * Updated content.
     */
    @Size(max = 10000, message = "Content must not exceed 10000 characters")
    private String content;

    /**
     * Updated summary.
     */
    @Size(max = 500, message = "Summary must not exceed 500 characters")
    private String summary;

    /**
     * Updated category.
     */
    private MemoryCategory category;

    /**
     * Updated importance level.
     */
    private ImportanceLevel importance;

    /**
     * Updated tags.
     */
    private List<String> tags;

    /**
     * Updated metadata.
     */
    private Map<String, Object> metadata;

    /**
     * Updated code example.
     */
    @Size(max = 5000, message = "Code example must not exceed 5000 characters")
    private String codeExample;

    /**
     * Updated programming language.
     */
    private String programmingLanguage;

    /**
     * Reason for the update.
     */
    private String changeReason;
}
