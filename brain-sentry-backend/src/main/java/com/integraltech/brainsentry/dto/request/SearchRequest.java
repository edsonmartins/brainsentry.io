package com.integraltech.brainsentry.dto.request;

import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * Request to search memories.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SearchRequest {

    /**
     * Search query text.
     */
    private String query;

    /**
     * Filter by categories.
     */
    private List<MemoryCategory> categories;

    /**
     * Filter by minimum importance level.
     */
    private ImportanceLevel minImportance;

    /**
     * Filter by tags.
     */
    private List<String> tags;

    /**
     * Maximum number of results.
     */
    @Builder.Default
    private Integer limit = 10;

    /**
     * Include related memories in results.
     */
    private Boolean includeRelated;

    /**
     * Tenant ID.
     */
    private String tenantId;
}
