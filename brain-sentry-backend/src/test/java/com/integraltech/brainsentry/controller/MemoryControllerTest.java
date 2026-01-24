package com.integraltech.brainsentry.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.dto.request.SearchRequest;
import com.integraltech.brainsentry.dto.request.UpdateMemoryRequest;
import com.integraltech.brainsentry.dto.response.MemoryListResponse;
import com.integraltech.brainsentry.dto.response.MemoryResponse;
import com.integraltech.brainsentry.service.EntityGraphService;
import com.integraltech.brainsentry.service.MemoryService;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

import java.util.List;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Unit tests for MemoryController.
 * Uses standalone MockMvc setup with mocked service layer.
 */
@DisplayName("MemoryController Unit Tests")
class MemoryControllerTest {

    private MockMvc mockMvc;

    private final ObjectMapper objectMapper = new ObjectMapper();

    private final MemoryService memoryService = Mockito.mock(MemoryService.class);

    private final EntityGraphService entityGraphService = Mockito.mock(EntityGraphService.class);

    private final MemoryController memoryController = new MemoryController(memoryService, entityGraphService);

    private final String tenantId = "test-tenant";
    private final String memoryId = "mem_test123";

    @BeforeEach
    void setUp() {
        mockMvc = MockMvcBuilders.standaloneSetup(memoryController).build();
        // Set tenant context for all tests
        com.integraltech.brainsentry.config.TenantContext.setTenantId(tenantId);
    }

    @AfterEach
    void tearDown() {
        // Clear tenant context
        com.integraltech.brainsentry.config.TenantContext.clear();
    }

    @Nested
    @DisplayName("POST /v1/memories")
    class CreateMemoryTests {

        @Test
        @DisplayName("Should return 400 when content is empty")
        void shouldReturn400WhenContentEmpty() throws Exception {
            CreateMemoryRequest request = CreateMemoryRequest.builder()
                    .content("") // Invalid: empty content
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("java"))
                    .build();

            mockMvc.perform(post("/v1/memories")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isBadRequest());
        }

        @Test
        @DisplayName("Should return 400 when request body is empty")
        void shouldReturn400ForEmptyBody() throws Exception {
            String emptyRequest = "{}";

            mockMvc.perform(post("/v1/memories")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(emptyRequest))
                    .andExpect(status().isBadRequest());
        }

        @Test
        @DisplayName("Should return 201 when summary is empty (summary is optional)")
        void shouldReturn201WhenSummaryEmpty() throws Exception {
            CreateMemoryRequest request = CreateMemoryRequest.builder()
                    .content("Test content")
                    .summary("") // Summary is optional, only has @Size constraint
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("java"))
                    .build();

            Mockito.when(memoryService.createMemory(any(CreateMemoryRequest.class)))
                    .thenReturn(MemoryResponse.builder().id(memoryId).build());

            mockMvc.perform(post("/v1/memories")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isCreated());
        }

        @Test
        @DisplayName("Should return 201 when creating memory successfully")
        void shouldReturn201WhenCreatingMemory() throws Exception {
            CreateMemoryRequest request = CreateMemoryRequest.builder()
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("java"))
                    .build();

            MemoryResponse response = MemoryResponse.builder()
                    .id(memoryId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("java"))
                    .build();

            Mockito.when(memoryService.createMemory(any(CreateMemoryRequest.class)))
                    .thenReturn(response);

            mockMvc.perform(post("/v1/memories")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isCreated());
        }
    }

    @Nested
    @DisplayName("GET /v1/memories/{id}")
    class GetMemoryTests {

        @Test
        @DisplayName("Should return memory by id")
        void shouldReturnMemoryById() throws Exception {
            MemoryResponse response = MemoryResponse.builder()
                    .id("mem-001")
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .build();

            Mockito.when(memoryService.getMemory(eq("mem-001")))
                    .thenReturn(response);

            mockMvc.perform(get("/v1/memories/mem-001"))
                    .andExpect(status().isOk());
        }
    }

    @Nested
    @DisplayName("GET /v1/memories")
    class ListMemoriesTests {

        @Test
        @DisplayName("Should return memory list")
        void shouldReturnList() throws Exception {
            MemoryListResponse listResponse = MemoryListResponse.builder()
                    .memories(List.of())
                    .page(0)
                    .size(20)
                    .totalElements(0L)
                    .totalPages(1)
                    .hasNext(false)
                    .hasPrevious(false)
                    .build();
            Mockito.when(memoryService.listMemories(eq(0), eq(20)))
                    .thenReturn(listResponse);

            mockMvc.perform(get("/v1/memories")
                            .param("page", "0")
                            .param("size", "20"))
                    .andExpect(status().isOk());
        }

        @Test
        @DisplayName("Should use default pagination values")
        void shouldUseDefaultPagination() throws Exception {
            MemoryListResponse listResponse = MemoryListResponse.builder()
                    .memories(List.of())
                    .page(0)
                    .size(20)
                    .totalElements(0L)
                    .totalPages(1)
                    .hasNext(false)
                    .hasPrevious(false)
                    .build();
            Mockito.when(memoryService.listMemories(eq(0), eq(20)))
                    .thenReturn(listResponse);

            mockMvc.perform(get("/v1/memories"))
                    .andExpect(status().isOk());
        }
    }

    @Nested
    @DisplayName("PUT /v1/memories/{id}")
    class UpdateMemoryTests {

        @Test
        @DisplayName("Should return 200 when content is empty (updates are partial)")
        void shouldReturn200WhenContentIsEmpty() throws Exception {
            // UpdateMemoryRequest has no @NotBlank fields (partial updates allowed)
            // Empty string is a valid update value (though service layer may ignore it)
            UpdateMemoryRequest request = UpdateMemoryRequest.builder()
                    .content("")
                    .summary("Updated summary")
                    .category(MemoryCategory.DECISION)
                    .importance(ImportanceLevel.CRITICAL)
                    .tags(List.of("java"))
                    .build();

            MemoryResponse response = MemoryResponse.builder()
                    .id(memoryId)
                    .summary("Updated summary")
                    .build();

            Mockito.when(memoryService.updateMemory(eq(memoryId), any(UpdateMemoryRequest.class)))
                    .thenReturn(response);

            mockMvc.perform(put("/v1/memories/{id}", memoryId)
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk());
        }

        @Test
        @DisplayName("Should return 200 when updating memory successfully")
        void shouldReturn200WhenUpdatingMemory() throws Exception {
            UpdateMemoryRequest request = UpdateMemoryRequest.builder()
                    .content("Updated content")
                    .summary("Updated summary")
                    .category(MemoryCategory.DECISION)
                    .importance(ImportanceLevel.CRITICAL)
                    .tags(List.of("java"))
                    .build();

            MemoryResponse response = MemoryResponse.builder()
                    .id(memoryId)
                    .content("Updated content")
                    .summary("Updated summary")
                    .category(MemoryCategory.DECISION)
                    .importance(ImportanceLevel.CRITICAL)
                    .tags(List.of("java"))
                    .build();

            Mockito.when(memoryService.updateMemory(eq(memoryId), any(UpdateMemoryRequest.class)))
                    .thenReturn(response);

            mockMvc.perform(put("/v1/memories/{id}", memoryId)
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk());
        }
    }

    @Nested
    @DisplayName("DELETE /v1/memories/{id}")
    class DeleteMemoryTests {

        @Test
        @DisplayName("Should return 204 when deleting memory")
        void shouldReturn204WhenDeletingMemory() throws Exception {
            Mockito.when(memoryService.deleteMemory(eq("mem-001")))
                    .thenReturn(true);

            mockMvc.perform(delete("/v1/memories/mem-001"))
                    .andExpect(status().isNoContent());
        }
    }

    @Nested
    @DisplayName("POST /v1/memories/search")
    class SearchTests {

        @Test
        @DisplayName("Should return search results")
        void shouldReturnSearchResults() throws Exception {
            SearchRequest request = SearchRequest.builder()
                    .query("spring boot")
                    .limit(10)
                    .build();

            Mockito.when(memoryService.search(any(SearchRequest.class)))
                    .thenReturn(List.of());

            mockMvc.perform(post("/v1/memories/search")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk());
        }
    }

    @Nested
    @DisplayName("GET /v1/memories/by-category/{category}")
    class GetByCategoryTests {

        @Test
        @DisplayName("Should return memories by category")
        void shouldReturnMemoriesByCategory() throws Exception {
            Mockito.when(memoryService.getByCategory(eq("PATTERN")))
                    .thenReturn(List.of());

            mockMvc.perform(get("/v1/memories/by-category/{category}", "PATTERN"))
                    .andExpect(status().isOk());
        }
    }

    @Nested
    @DisplayName("GET /v1/memories/by-importance/{importance}")
    class GetByImportanceTests {

        @Test
        @DisplayName("Should return memories by importance")
        void shouldReturnMemoriesByImportance() throws Exception {
            Mockito.when(memoryService.getByImportance(eq("IMPORTANT")))
                    .thenReturn(List.of());

            mockMvc.perform(get("/v1/memories/by-importance/{importance}", "IMPORTANT"))
                    .andExpect(status().isOk());
        }
    }

    @Nested
    @DisplayName("GET /v1/memories/{id}/related")
    class GetRelatedTests {

        @Test
        @DisplayName("Should return related memories")
        void shouldReturnRelatedMemories() throws Exception {
            Mockito.when(memoryService.getRelated(eq(memoryId), eq(2)))
                    .thenReturn(List.of());

            mockMvc.perform(get("/v1/memories/{id}/related", memoryId)
                            .param("depth", "2"))
                    .andExpect(status().isOk());
        }

        @Test
        @DisplayName("Should use default depth value")
        void shouldUseDefaultDepth() throws Exception {
            Mockito.when(memoryService.getRelated(eq(memoryId), eq(1)))
                    .thenReturn(List.of());

            mockMvc.perform(get("/v1/memories/{id}/related", memoryId))
                    .andExpect(status().isOk());
        }
    }

    @Nested
    @DisplayName("POST /v1/memories/{id}/feedback")
    class RecordFeedbackTests {

        @Test
        @DisplayName("Should record helpful feedback")
        void shouldRecordHelpfulFeedback() throws Exception {
            Mockito.doNothing().when(memoryService).recordFeedback(eq(memoryId), eq(true));

            mockMvc.perform(post("/v1/memories/{id}/feedback", memoryId)
                            .param("helpful", "true"))
                    .andExpect(status().isOk());
        }

        @Test
        @DisplayName("Should record not helpful feedback")
        void shouldRecordNotHelpfulFeedback() throws Exception {
            Mockito.doNothing().when(memoryService).recordFeedback(eq(memoryId), eq(false));

            mockMvc.perform(post("/v1/memories/{id}/feedback", memoryId)
                            .param("helpful", "false"))
                    .andExpect(status().isOk());
        }
    }
}
