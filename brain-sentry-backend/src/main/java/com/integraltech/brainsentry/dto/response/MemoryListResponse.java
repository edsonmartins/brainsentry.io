package com.integraltech.brainsentry.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * Response for paginated memory list.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MemoryListResponse {

    /**
     * List of memories.
     */
    private List<MemoryResponse> memories;

    /**
     * Current page number (0-indexed).
     */
    private Integer page;

    /**
     * Page size.
     */
    private Integer size;

    /**
     * Total number of elements.
     */
    private Long totalElements;

    /**
     * Total number of pages.
     */
    private Integer totalPages;

    /**
     * Whether there's a next page.
     */
    private Boolean hasNext;

    /**
     * Whether there's a previous page.
     */
    private Boolean hasPrevious;
}
