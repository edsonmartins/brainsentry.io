package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.dto.request.CompressionRequest;
import com.integraltech.brainsentry.dto.response.CompressedContextResponse;
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

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
@DisplayName("ArchitectService Unit Tests")
class ArchitectServiceTest {

    @Mock
    private OpenRouterService openRouterService;

    @InjectMocks
    private ArchitectService architectService;

    @DisplayName("Context Compression Tests")
    @Nested
    class CompressionTests {

        @Test
        @DisplayName("Should not compress when below threshold")
        void testCompressContext_BelowThreshold_NoCompression() {
            // Given
            List<CompressionRequest.Message> messages = Arrays.asList(
                createMessage("user", "Short message"),
                createMessage("assistant", "Short response")
            );

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{\"summary\": {\"taskGoal\": \"Test task\"}}");

            // When
            CompressedContextResponse response = architectService.compressContext(messages, 100000);

            // Then
            assertThat(response.getCompressed()).isFalse();
            assertThat(response.getCompressionRatio()).isEqualTo(1.0);
            assertThat(response.getPreservedMessages()).hasSize(2);

            verify(openRouterService, never()).chat(anyString(), anyString());
        }

        @Test
        @DisplayName("Should compress when above threshold")
        void testCompressContext_AboveThreshold_Compresses() {
            // Given
            StringBuilder longContent = new StringBuilder();
            for (int i = 0; i < 10000; i++) {
                longContent.append("This is a long message content. ");
            }

            List<CompressionRequest.Message> messages = Arrays.asList(
                createMessage("user", longContent.toString()),
                createMessage("assistant", "Response")
            );

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{\"summary\": {\"taskGoal\": \"Complete task\"}, " +
                        "\"preservedMessages\": [{\"role\": \"user\", \"content\": \"Recent message\"}]}");

            // When
            CompressedContextResponse response = architectService.compressContext(messages, 1000);

            // Then
            assertThat(response.getCompressed()).isTrue();
            assertThat(response.getCompressionRatio()).isLessThan(1.0);

            verify(openRouterService).chat(anyString(), anyString());
        }

        @Test
        @DisplayName("Should compress with default threshold when null provided")
        void testCompressContext_NullThreshold_UsesDefault() {
            // Given
            // Create a large message list that exceeds default threshold (100k tokens)
            List<CompressionRequest.Message> messages = new ArrayList<>();
            for (int i = 0; i < 10000; i++) {
                messages.add(createMessage("user", "Large message content to exceed threshold ".repeat(10)));
            }

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{\"summary\": {}}");

            // When
            CompressedContextResponse response = architectService.compressContext(messages, null);

            // Then
            assertThat(response).isNotNull();
            verify(openRouterService).chat(anyString(), anyString());
        }
    }

    @DisplayName("Summary Extraction Tests")
    @Nested
    class SummaryExtractionTests {

        @Test
        @DisplayName("Should extract summary from messages")
        void testExtractSummary_ParsesResponse() {
            // Given
            List<CompressionRequest.Message> messages = Arrays.asList(
                createMessage("user", "I need to implement a feature"),
                createMessage("assistant", "I'll help you implement it"),
                createMessage("tool", "Code generated")
            );

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{\"taskGoal\": \"Implement feature\", \"keyDecisions\": [\"Use Spring Boot\"], " +
                        "\"openTodos\": [\"Write tests\"], \"criticalErrors\": [], " +
                        "\"importantFileChanges\": [\"UserController.java\"], " +
                        "\"additionalContext\": \"Requires authentication\"}");

            // When
            CompressedContextResponse.StructuredSummary summary =
                architectService.extractSummary(messages);

            // Then
            assertThat(summary).isNotNull();
            assertThat(summary.getTaskGoal()).contains("Implement");
            assertThat(summary.getKeyDecisions()).isNotEmpty();
            assertThat(summary.getOpenTodos()).isNotEmpty();
        }

        @Test
        @DisplayName("Should handle empty messages")
        void testExtractSummary_EmptyMessages() {
            // Given
            List<CompressionRequest.Message> messages = Collections.emptyList();

            // When
            CompressedContextResponse.StructuredSummary summary =
                architectService.extractSummary(messages);

            // Then
            assertThat(summary).isNotNull();
            verify(openRouterService, never()).chat(anyString(), anyString());
        }
    }

    @DisplayName("Critical Message Identification Tests")
    @Nested
    class CriticalMessageTests {

        @Test
        @DisplayName("Should identify error messages as critical")
        void testIdentifyCriticalMessages_ErrorMessages() {
            // Given
            List<CompressionRequest.Message> messages = Arrays.asList(
                createMessage("user", "Regular message"),
                createMessage("error", "Error occurred: timeout"),
                createMessage("assistant", "Here's the fix")
            );

            // When
            List<CompressionRequest.Message> critical =
                architectService.identifyCriticalMessages(messages, null);

            // Then
            assertThat(critical).hasSize(1);
            assertThat(critical.get(0).getRole()).isEqualTo("error");
        }

        @Test
        @DisplayName("Should identify messages with keywords as critical")
        void testIdentifyCriticalMessages_WithKeywords() {
            // Given
            List<CompressionRequest.Message> messages = Arrays.asList(
                createMessage("user", "I want to deploy to production"),
                createMessage("assistant", "Let me deploy"),
                createMessage("user", "Make sure to backup database first"),
                createMessage("assistant", "Good point on backup")
            );

            List<String> keywords = Arrays.asList("backup", "deploy");

            // When
            List<CompressionRequest.Message> critical =
                architectService.identifyCriticalMessages(messages, keywords);

            // Then - All 4 messages contain either "backup" or "deploy"
            assertThat(critical).hasSize(4);
        }

        @Test
        @DisplayName("Should identify system messages as critical")
        void testIdentifyCriticalMessages_SystemMessages() {
            // Given
            List<CompressionRequest.Message> messages = Arrays.asList(
                createMessage("user", "Regular message"),
                createMessage("system", "System initialization complete"),
                createMessage("assistant", "Response")
            );

            // When
            List<CompressionRequest.Message> critical =
                architectService.identifyCriticalMessages(messages, null);

            // Then
            assertThat(critical).hasSize(1);
            assertThat(critical.get(0).getRole()).isEqualTo("system");
        }
    }

    @DisplayName("Compression Check Tests")
    @Nested
    class CompressionCheckTests {

        @Test
        @DisplayName("Should return true when above threshold")
        void testShouldCompress_AboveThreshold_ReturnsTrue() {
            // Given
            StringBuilder longContent = new StringBuilder();
            for (int i = 0; i < 10000; i++) {
                longContent.append("Long content for compression. ");
            }

            List<CompressionRequest.Message> messages = Arrays.asList(
                createMessage("user", longContent.toString())
            );

            // When
            boolean shouldCompress = architectService.shouldCompress(messages, 1000);

            // Then
            assertThat(shouldCompress).isTrue();
        }

        @Test
        @DisplayName("Should return false when below threshold")
        void testShouldCompress_BelowThreshold_ReturnsFalse() {
            // Given
            List<CompressionRequest.Message> messages = Arrays.asList(
                createMessage("user", "Short message")
            );

            // When
            boolean shouldCompress = architectService.shouldCompress(messages, 100000);

            // Then
            assertThat(shouldCompress).isFalse();
        }

        @Test
        @DisplayName("Should use default threshold when null provided")
        void testShouldCompress_NullThreshold_UsesDefault() {
            // Given
            List<CompressionRequest.Message> messages = Collections.emptyList();

            // When
            boolean shouldCompress = architectService.shouldCompress(messages, null);

            // Then
            assertThat(shouldCompress).isFalse();
        }
    }

    // Helper method
    private CompressionRequest.Message createMessage(String role, String content) {
        return CompressionRequest.Message.builder()
            .role(role)
            .content(content)
            .timestamp(System.currentTimeMillis())
            .build();
    }
}
