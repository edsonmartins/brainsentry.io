package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.domain.enums.ValidationStatus;
import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.dto.request.SearchRequest;
import com.integraltech.brainsentry.dto.request.UpdateMemoryRequest;
import com.integraltech.brainsentry.dto.response.MemoryListResponse;
import com.integraltech.brainsentry.dto.response.MemoryResponse;
import com.integraltech.brainsentry.mapper.MemoryMapper;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.repository.MemoryRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.MockedStatic;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
@DisplayName("MemoryService Unit Tests")
class MemoryServiceTest {

    @Mock
    private MemoryJpaRepository memoryJpaRepo;

    @Mock
    private MemoryRepository memoryGraphRepo;

    @Mock
    private EmbeddingService embeddingService;

    @Mock
    private OpenRouterService openRouterService;

    @Mock
    private MemoryMapper memoryMapper;

    @InjectMocks
    private MemoryService memoryService;

    private final String tenantId = "test-tenant";
    private Memory testMemory;
    private MemoryResponse testResponse;

    @BeforeEach
    void setUp() {
        testMemory = createTestMemory();
        testResponse = createTestResponse();

        // Setup default mapper behavior
        when(memoryMapper.toResponse(any(Memory.class))).thenReturn(testResponse);
    }

    private Memory createTestMemory() {
        return Memory.builder()
                .id("mem-001")
                .tenantId(tenantId)
                .content("Test content")
                .summary("Test summary")
                .category(MemoryCategory.PATTERN)
                .importance(ImportanceLevel.IMPORTANT)
                .tags(List.of("tag1", "tag2"))
                .validationStatus(ValidationStatus.APPROVED)
                .version(1)
                .accessCount(0)
                .injectionCount(0)
                .helpfulCount(0)
                .notHelpfulCount(0)
                .createdAt(Instant.now())
                .build();
    }

    private MemoryResponse createTestResponse() {
        return MemoryResponse.builder()
                .id("mem-001")
                .tenantId(tenantId)
                .content("Test content")
                .summary("Test summary")
                .category(MemoryCategory.PATTERN)
                .importance(ImportanceLevel.IMPORTANT)
                .tags(List.of("tag1", "tag2"))
                .validationStatus(ValidationStatus.APPROVED)
                .version(1)
                .accessCount(0)
                .injectionCount(0)
                .helpfulCount(0)
                .build();
    }

    @Nested
    @DisplayName("createMemory()")
    class CreateMemoryTests {

        @Test
        @DisplayName("Should create memory with all fields provided")
        void shouldCreateMemoryWithAllFields() {
            try (MockedStatic<TenantContext> tenantContext = mockStatic(TenantContext.class)) {
                tenantContext.when(TenantContext::getTenantId).thenReturn(tenantId);

                CreateMemoryRequest request = CreateMemoryRequest.builder()
                        .content("Test content")
                        .summary("Test summary")
                        .category(MemoryCategory.PATTERN)
                        .importance(ImportanceLevel.IMPORTANT)
                        .tags(List.of("tag1", "tag2"))
                        .tenantId(tenantId)
                        .build();

                float[] embedding = new float[]{0.1f, 0.2f, 0.3f};
                when(embeddingService.embed(anyString())).thenReturn(embedding);
                when(memoryJpaRepo.save(any(Memory.class))).thenReturn(testMemory);
                when(memoryGraphRepo.save(any(Memory.class))).thenReturn(testMemory);

                MemoryResponse response = memoryService.createMemory(request);

                assertThat(response).isNotNull();
                assertThat(response.getId()).isEqualTo("mem-001");

                verify(memoryJpaRepo).save(any(Memory.class));
                verify(memoryGraphRepo).save(any(Memory.class));
                verify(embeddingService).embed("Test content");
            }
        }

        @Test
        @DisplayName("Should use default tenant when not provided")
        void shouldUseDefaultTenant() {
            try (MockedStatic<TenantContext> tenantContext = mockStatic(TenantContext.class)) {
                tenantContext.when(TenantContext::getTenantId).thenReturn(tenantId);

                CreateMemoryRequest request = CreateMemoryRequest.builder()
                        .content("Test content")
                        .summary("Test summary")
                        .category(MemoryCategory.PATTERN)
                        .importance(ImportanceLevel.IMPORTANT)
                        .tags(List.of("tag1"))
                        .build();

                when(embeddingService.embed(anyString())).thenReturn(new float[]{0.1f});
                when(memoryJpaRepo.save(any(Memory.class))).thenReturn(testMemory);
                when(memoryGraphRepo.save(any(Memory.class))).thenReturn(testMemory);

                memoryService.createMemory(request);

                verify(memoryJpaRepo).save(argThat(mem ->
                        tenantId.equals(mem.getTenantId())
                ));
            }
        }

        @Test
        @DisplayName("Should auto-analyze when category not provided")
        void shouldAutoAnalyzeWhenCategoryNotProvided() {
            try (MockedStatic<TenantContext> tenantContext = mockStatic(TenantContext.class)) {
                tenantContext.when(TenantContext::getTenantId).thenReturn(tenantId);

                CreateMemoryRequest request = CreateMemoryRequest.builder()
                        .content("Important bug fix pattern")
                        .summary("Test summary")
                        .importance(ImportanceLevel.CRITICAL)
                        .tags(List.of("tag1"))
                        .build();

                when(openRouterService.analyzeImportance(anyString()))
                        .thenReturn(new OpenRouterService.ImportanceAnalysis(
                                true, "CRITICAL", "BUG", "", ""));

                when(embeddingService.embed(anyString())).thenReturn(new float[]{0.1f});
                when(memoryJpaRepo.save(any(Memory.class))).thenReturn(testMemory);
                when(memoryGraphRepo.save(any(Memory.class))).thenReturn(testMemory);

                memoryService.createMemory(request);

                verify(openRouterService).analyzeImportance("Important bug fix pattern");
                verify(memoryJpaRepo).save(argThat(mem ->
                        MemoryCategory.BUG.equals(mem.getCategory())
                ));
            }
        }
    }

    @Nested
    @DisplayName("getMemory()")
    class GetMemoryTests {

        @Test
        @DisplayName("Should return memory when found")
        void shouldReturnMemoryWhenFound() {
            when(memoryJpaRepo.findById("mem-001")).thenReturn(Optional.of(testMemory));

            MemoryResponse response = memoryService.getMemory("mem-001");

            assertThat(response).isNotNull();
            assertThat(response.getId()).isEqualTo("mem-001");

            verify(memoryJpaRepo).save(argThat(mem ->
                    mem.getAccessCount() == 1
            ));
        }

        @Test
        @DisplayName("Should throw exception when memory not found")
        void shouldThrowWhenNotFound() {
            when(memoryJpaRepo.findById("invalid-id")).thenReturn(Optional.empty());

            assertThatThrownBy(() -> memoryService.getMemory("invalid-id"))
                    .isInstanceOf(RuntimeException.class)
                    .hasMessageContaining("Memory not found");
        }
    }

    @Nested
    @DisplayName("listMemories()")
    class ListMemoriesTests {

        @Test
        @DisplayName("Should return paginated memory list")
        void shouldReturnPaginatedList() {
            List<Memory> memories = List.of(testMemory);
            Page<Memory> page = new PageImpl<>(memories, PageRequest.of(0, 20), 1);

            when(memoryJpaRepo.findAll(any(PageRequest.class))).thenReturn(page);

            MemoryListResponse response = memoryService.listMemories(0, 20);

            assertThat(response).isNotNull();
            assertThat(response.getMemories()).hasSize(1);
            assertThat(response.getPage()).isEqualTo(0);
            assertThat(response.getSize()).isEqualTo(20);
            assertThat(response.getTotalElements()).isEqualTo(1);
            assertThat(response.getTotalPages()).isEqualTo(1);
        }

        @Test
        @DisplayName("Should return empty list when no memories")
        void shouldReturnEmptyList() {
            Page<Memory> emptyPage = new PageImpl<>(List.of(), PageRequest.of(0, 20), 0);

            when(memoryJpaRepo.findAll(any(PageRequest.class))).thenReturn(emptyPage);

            MemoryListResponse response = memoryService.listMemories(0, 20);

            assertThat(response.getMemories()).isEmpty();
            assertThat(response.getTotalElements()).isEqualTo(0);
        }
    }

    @Nested
    @DisplayName("updateMemory()")
    class UpdateMemoryTests {

        @Test
        @DisplayName("Should update memory fields")
        void shouldUpdateMemoryFields() {
            UpdateMemoryRequest request = UpdateMemoryRequest.builder()
                    .content("Updated content")
                    .summary("Updated summary")
                    .category(MemoryCategory.INTEGRATION)
                    .build();

            when(memoryJpaRepo.findById("mem-001")).thenReturn(Optional.of(testMemory));
            when(embeddingService.embed(anyString())).thenReturn(new float[]{0.1f});
            when(memoryJpaRepo.save(any(Memory.class))).thenReturn(testMemory);

            MemoryResponse response = memoryService.updateMemory("mem-001", request);

            assertThat(response).isNotNull();
            verify(memoryGraphRepo).archiveVersion(testMemory);
            verify(memoryJpaRepo).save(argThat(mem ->
                    "Updated content".equals(mem.getContent()) &&
                            "Updated summary".equals(mem.getSummary()) &&
                            MemoryCategory.INTEGRATION.equals(mem.getCategory()) &&
                            mem.getVersion() == 2
            ));
        }

        @Test
        @DisplayName("Should throw exception when updating non-existent memory")
        void shouldThrowWhenUpdatingNonExistent() {
            UpdateMemoryRequest request = UpdateMemoryRequest.builder()
                    .content("Updated content")
                    .build();

            when(memoryJpaRepo.findById("invalid-id")).thenReturn(Optional.empty());

            assertThatThrownBy(() -> memoryService.updateMemory("invalid-id", request))
                    .isInstanceOf(RuntimeException.class)
                    .hasMessageContaining("Memory not found");
        }
    }

    @Nested
    @DisplayName("deleteMemory()")
    class DeleteMemoryTests {

        @Test
        @DisplayName("Should delete memory when exists")
        void shouldDeleteMemoryWhenExists() {
            when(memoryJpaRepo.existsById("mem-001")).thenReturn(true);
            when(memoryGraphRepo.deleteById("mem-001")).thenReturn(true);

            boolean result = memoryService.deleteMemory("mem-001");

            assertThat(result).isTrue();
            verify(memoryJpaRepo).deleteById("mem-001");
            verify(memoryGraphRepo).deleteById("mem-001");
        }

        @Test
        @DisplayName("Should return false when memory not found")
        void shouldReturnFalseWhenNotFound() {
            when(memoryJpaRepo.existsById("invalid-id")).thenReturn(false);

            boolean result = memoryService.deleteMemory("invalid-id");

            assertThat(result).isFalse();
            verify(memoryJpaRepo, never()).deleteById(anyString());
            verify(memoryGraphRepo, never()).deleteById(anyString());
        }
    }

    @Nested
    @DisplayName("search()")
    class SearchTests {

        @Test
        @DisplayName("Should search memories using vector search")
        void shouldSearchUsingVectorSearch() {
            try (MockedStatic<TenantContext> tenantContext = mockStatic(TenantContext.class)) {
                tenantContext.when(TenantContext::getTenantId).thenReturn(tenantId);

                SearchRequest request = SearchRequest.builder()
                        .query("test query")
                        .limit(10)
                        .build();

                float[] embedding = new float[]{0.1f, 0.2f};
                when(embeddingService.embed("test query")).thenReturn(embedding);
                when(memoryGraphRepo.vectorSearch(eq(embedding), eq(10), eq(tenantId)))
                        .thenReturn(List.of(testMemory));

                List<MemoryResponse> results = memoryService.search(request);

                assertThat(results).hasSize(1);
                assertThat(results.get(0).getId()).isEqualTo("mem-001");

                verify(embeddingService).embed("test query");
                verify(memoryGraphRepo).vectorSearch(eq(embedding), eq(10), eq(tenantId));
            }
        }

        @Test
        @DisplayName("Should use default limit when not provided")
        void shouldUseDefaultLimit() {
            try (MockedStatic<TenantContext> tenantContext = mockStatic(TenantContext.class)) {
                tenantContext.when(TenantContext::getTenantId).thenReturn(tenantId);

                SearchRequest request = SearchRequest.builder()
                        .query("test query")
                        .build();

                when(embeddingService.embed(anyString())).thenReturn(new float[]{0.1f});
                when(memoryGraphRepo.vectorSearch(any(), eq(10), any()))
                        .thenReturn(List.of(testMemory));

                memoryService.search(request);

                verify(memoryGraphRepo).vectorSearch(any(), eq(10), any());
            }
        }
    }

    @Nested
    @DisplayName("recordFeedback()")
    class RecordFeedbackTests {

        @Test
        @DisplayName("Should increment helpful count when feedback is positive")
        void shouldIncrementHelpfulCount() {
            when(memoryJpaRepo.findById("mem-001")).thenReturn(Optional.of(testMemory));
            when(memoryJpaRepo.save(any(Memory.class))).thenReturn(testMemory);

            memoryService.recordFeedback("mem-001", true);

            verify(memoryJpaRepo).save(argThat(mem ->
                    mem.getHelpfulCount() == 1
            ));
        }

        @Test
        @DisplayName("Should increment not helpful count when feedback is negative")
        void shouldIncrementNotHelpfulCount() {
            when(memoryJpaRepo.findById("mem-001")).thenReturn(Optional.of(testMemory));
            when(memoryJpaRepo.save(any(Memory.class))).thenReturn(testMemory);

            memoryService.recordFeedback("mem-001", false);

            verify(memoryJpaRepo).save(argThat(mem ->
                    mem.getNotHelpfulCount() == 1
            ));
        }

        @Test
        @DisplayName("Should throw exception when memory not found")
        void shouldThrowWhenRecordingFeedbackForNonExistent() {
            when(memoryJpaRepo.findById("invalid-id")).thenReturn(Optional.empty());

            assertThatThrownBy(() -> memoryService.recordFeedback("invalid-id", true))
                    .isInstanceOf(RuntimeException.class)
                    .hasMessageContaining("Memory not found");
        }
    }

    @Nested
    @DisplayName("getByCategory()")
    class GetByCategoryTests {

        @Test
        @DisplayName("Should return memories by category")
        void shouldReturnMemoriesByCategory() {
            when(memoryJpaRepo.findByCategory("PATTERN")).thenReturn(List.of(testMemory));

            List<MemoryResponse> results = memoryService.getByCategory("PATTERN");

            assertThat(results).hasSize(1);
            assertThat(results.get(0).getId()).isEqualTo("mem-001");

            verify(memoryJpaRepo).findByCategory("PATTERN");
        }
    }

    @Nested
    @DisplayName("getByImportance()")
    class GetByImportanceTests {

        @Test
        @DisplayName("Should return memories by importance")
        void shouldReturnMemoriesByImportance() {
            when(memoryJpaRepo.findByImportance("CRITICAL")).thenReturn(List.of(testMemory));

            List<MemoryResponse> results = memoryService.getByImportance("CRITICAL");

            assertThat(results).hasSize(1);
            assertThat(results.get(0).getId()).isEqualTo("mem-001");

            verify(memoryJpaRepo).findByImportance("CRITICAL");
        }
    }

    @Nested
    @DisplayName("getRelated()")
    class GetRelatedTests {

        @Test
        @DisplayName("Should return related memories")
        void shouldReturnRelatedMemories() {
            try (MockedStatic<TenantContext> tenantContext = mockStatic(TenantContext.class)) {
                tenantContext.when(TenantContext::getTenantId).thenReturn(tenantId);

                Memory relatedMemory = Memory.builder()
                        .id("mem-002")
                        .tenantId(tenantId)
                        .content("Related content")
                        .summary("Related summary")
                        .category(MemoryCategory.PATTERN)
                        .importance(ImportanceLevel.MINOR)
                        .tags(List.of())
                        .version(1)
                        .accessCount(0)
                        .injectionCount(0)
                        .helpfulCount(0)
                        .notHelpfulCount(0)
                        .createdAt(Instant.now())
                        .build();

                when(memoryGraphRepo.findRelated("mem-001", 2, tenantId))
                        .thenReturn(List.of(testMemory, relatedMemory));

                List<MemoryResponse> results = memoryService.getRelated("mem-001", 2);

                assertThat(results).hasSize(2);

                verify(memoryGraphRepo).findRelated("mem-001", 2, tenantId);
            }
        }
    }
}
