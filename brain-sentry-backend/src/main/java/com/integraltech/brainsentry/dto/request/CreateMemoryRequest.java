package com.integraltech.brainsentry.dto.request;

import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;
import java.util.Map;

/**
 * Request to create a new memory.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CreateMemoryRequest {

    /**
     * The full content of the memory.
     */
    @NotBlank(message = "Content is required")
    @Size(max = 10000, message = "Content must not exceed 10000 characters")
    private String content;

    /**
     * Brief summary (1-2 sentences).
     */
    @Size(max = 500, message = "Summary must not exceed 500 characters")
    private String summary;

    /**
     * Memory category.
     */
    private MemoryCategory category;

    /**
     * Importance level.
     */
    private ImportanceLevel importance;

    /**
     * Tags for filtering.
     */
    private List<String> tags;

    /**
     * Additional metadata.
     */
    private Map<String, Object> metadata;

    /**
     * Where this memory came from.
     */
    private String sourceType;

    /**
     * Reference to the source.
     */
    private String sourceReference;

    /**
     * Code example (optional).
     */
    @Size(max = 5000, message = "Code example must not exceed 5000 characters")
    private String codeExample;

    /**
     * Programming language of the code example.
     */
    private String programmingLanguage;

    /**
     * User creating this memory.
     */
    private String createdBy;

    /**
     * Tenant ID.
     */
    private String tenantId;

    // Manual builder method in case Lombok doesn't generate it
    public static CreateMemoryRequestBuilder builder() {
        return new CreateMemoryRequestBuilder();
    }

    public static class CreateMemoryRequestBuilder {
        private String content;
        private String summary;
        private MemoryCategory category;
        private ImportanceLevel importance;
        private List<String> tags;
        private Map<String, Object> metadata;
        private String sourceType;
        private String sourceReference;
        private String codeExample;
        private String programmingLanguage;
        private String createdBy;
        private String tenantId;

        public CreateMemoryRequestBuilder content(String content) {
            this.content = content;
            return this;
        }

        public CreateMemoryRequestBuilder summary(String summary) {
            this.summary = summary;
            return this;
        }

        public CreateMemoryRequestBuilder category(MemoryCategory category) {
            this.category = category;
            return this;
        }

        public CreateMemoryRequestBuilder importance(ImportanceLevel importance) {
            this.importance = importance;
            return this;
        }

        public CreateMemoryRequestBuilder tags(List<String> tags) {
            this.tags = tags;
            return this;
        }

        public CreateMemoryRequestBuilder metadata(Map<String, Object> metadata) {
            this.metadata = metadata;
            return this;
        }

        public CreateMemoryRequestBuilder sourceType(String sourceType) {
            this.sourceType = sourceType;
            return this;
        }

        public CreateMemoryRequestBuilder sourceReference(String sourceReference) {
            this.sourceReference = sourceReference;
            return this;
        }

        public CreateMemoryRequestBuilder codeExample(String codeExample) {
            this.codeExample = codeExample;
            return this;
        }

        public CreateMemoryRequestBuilder programmingLanguage(String programmingLanguage) {
            this.programmingLanguage = programmingLanguage;
            return this;
        }

        public CreateMemoryRequestBuilder createdBy(String createdBy) {
            this.createdBy = createdBy;
            return this;
        }

        public CreateMemoryRequestBuilder tenantId(String tenantId) {
            this.tenantId = tenantId;
            return this;
        }

        public CreateMemoryRequest build() {
            return new CreateMemoryRequest(this);
        }
    }

    // Private constructor for builder
    private CreateMemoryRequest(CreateMemoryRequestBuilder builder) {
        this.content = builder.content;
        this.summary = builder.summary;
        this.category = builder.category;
        this.importance = builder.importance;
        this.tags = builder.tags;
        this.metadata = builder.metadata;
        this.sourceType = builder.sourceType;
        this.sourceReference = builder.sourceReference;
        this.codeExample = builder.codeExample;
        this.programmingLanguage = builder.programmingLanguage;
        this.createdBy = builder.createdBy;
        this.tenantId = builder.tenantId;
    }
}
