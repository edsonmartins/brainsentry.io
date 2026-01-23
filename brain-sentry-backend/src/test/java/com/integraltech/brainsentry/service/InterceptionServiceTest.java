package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.domain.enums.NoteSeverity;
import com.integraltech.brainsentry.dto.request.InterceptRequest;
import com.integraltech.brainsentry.dto.response.InterceptResponse;
import com.integraltech.brainsentry.repository.MemoryRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
@DisplayName("InterceptionService Unit Tests")
class InterceptionServiceTest {

    @Mock
    private OpenRouterService openRouterService;

    @Mock
    private EmbeddingService embeddingService;

    @Mock
    private MemoryRepository memoryRepository;

    @Mock
    private AuditService auditService;

    @Mock
    private NoteRetrievalService noteRetrievalService;

    @InjectMocks
    private InterceptionService interceptionService;

    @DisplayName("Note Integration Tests (Confucius Spec)")
    @Nested
    class NoteIntegrationTests {

        @Test
        @DisplayName("Should retrieve hindsight notes when error keywords present")
        void testInterceptAndEnhance_WithErrorKeywords_RetrievesHindsightNotes() {
            // Given
            InterceptRequest request = InterceptRequest.builder()
                .prompt("Fix the NullPointerException in UserService")
                .sessionId("session-123")
                .tenantId("tenant-abc")
                .build();

            HindsightNote hindsightNote = HindsightNote.builder()
                .id("note-1")
                .tenantId("tenant-abc")
                .title("NPE Fix")
                .errorType("NullPointerException")
                .severity(NoteSeverity.HIGH)
                .resolution("Add null check")
                .build();

            when(openRouterService.analyzeRelevance(anyString(), any()))
                .thenReturn(new OpenRouterService.RelevanceAnalysis(true, "Error detected", 0.9));

            when(noteRetrievalService.searchHindsightNotes(anyString(), eq("NullPointerException"), eq("tenant-abc")))
                .thenReturn(Arrays.asList(hindsightNote));

            // When
            InterceptResponse response = interceptionService.interceptAndEnhance(request);

            // Then
            assertThat(response).isNotNull();
            assertThat(response.getEnhanced()).isTrue();
            assertThat(response.getNotesUsed()).isNotNull();
            assertThat(response.getNotesUsed()).hasSize(1);
            assertThat(response.getNotesUsed().get(0).getType()).isEqualTo("HINDSIGHT");

            verify(noteRetrievalService).recordNoteAccess("note-1");
        }

        @Test
        @DisplayName("Should include notes in formatted context")
        void testFormatContextWithNotes_IncludesNotesSection() {
            // Given
            List<Memory> memories = Arrays.asList(
                Memory.builder()
                    .id("mem-1")
                    .summary("Use Spring Boot")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .build()
            );

            HindsightNote hindsightNote = HindsightNote.builder()
                .id("note-1")
                .title("NPE Fix")
                .resolution("Add null check")
                .severity(NoteSeverity.HIGH)
                .build();

            when(openRouterService.analyzeRelevance(anyString(), any()))
                .thenReturn(new OpenRouterService.RelevanceAnalysis(true, "Relevant", 0.8));

            when(noteRetrievalService.searchHindsightNotes(anyString(), anyString(), anyString()))
                .thenReturn(Arrays.asList(hindsightNote));

            when(memoryRepository.vectorSearch(any(float[].class), eq(5), eq("tenant-abc")))
                .thenReturn(Arrays.asList(memories.get(0)));

            // When
            InterceptResponse response = interceptionService.interceptAndEnhance(
                InterceptRequest.builder()
                    .prompt("Help with error")
                    .tenantId("tenant-abc")
                    .build()
            );

            // Then
            assertThat(response).isNotNull();
            assertThat(response.getContextInjected()).contains("Past Learnings");
        }

        @Test
        @DisplayName("Should not enhance when no memories and no notes found")
        void testInterceptAndEnhance_NoContext_PassesThrough() {
            // Given
            InterceptRequest request = InterceptRequest.builder()
                .prompt("Simple question")
                .tenantId("tenant-xyz")
                .build();

            when(openRouterService.analyzeRelevance(anyString(), any()))
                .thenReturn(new OpenRouterService.RelevanceAnalysis(true, "Maybe relevant", 0.6));

            when(memoryRepository.vectorSearch(any(float[].class), anyInt(), anyString()))
                .thenReturn(Collections.emptyList());

            when(noteRetrievalService.searchHindsightNotes(anyString(), anyString(), anyString()))
                .thenReturn(Collections.emptyList());

            when(noteRetrievalService.getRelevantNotes(anyString(), anyString(), anyInt()))
                .thenReturn(Collections.emptyList());

            // When
            InterceptResponse response = interceptionService.interceptAndEnhance(request);

            // Then
            assertThat(response.getEnhanced()).isFalse();
            assertThat(response.getMemoriesUsed()).isEmpty();
            assertThat(response.getNotesUsed()).isEmpty();
        }
    }

    @DisplayName("Error Detection Tests")
    @Nested
    class ErrorDetectionTests {

        @Test
        @DisplayName("Should detect error keywords in prompt")
        void testContainsErrorKeywords_DetectsErrors() {
            // Given
            String errorPrompt = "Fix the NullPointerException timeout error";

            // Use reflection to test private method
            try {
                java.lang.reflect.Method method = InterceptionService.class
                    .getDeclaredMethod("containsErrorKeywords", String.class);
                method.setAccessible(true);

                // When
                boolean hasError = (boolean) method.invoke(interceptionService, errorPrompt);

                // Then
                assertThat(hasError).isTrue();
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }

        @Test
        @DisplayName("Should extract error type from prompt")
        void testExtractErrorType_ExtractsCorrectly() {
            // Given
            String npePrompt = "NullPointerException in UserService";

            // Use reflection to test private method
            try {
                java.lang.reflect.Method method = InterceptionService.class
                    .getDeclaredMethod("extractErrorType", String.class);
                method.setAccessible(true);

                // When
                String errorType = (String) method.invoke(interceptionService, npePrompt);

                // Then
                assertThat(errorType).isEqualTo("NullPointerException");
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }
    }
}
