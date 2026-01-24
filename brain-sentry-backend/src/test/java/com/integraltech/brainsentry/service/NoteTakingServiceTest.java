package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.Note;
import com.integraltech.brainsentry.dto.request.CreateHindsightNoteRequest;
import com.integraltech.brainsentry.dto.request.SessionAnalysisRequest;
import com.integraltech.brainsentry.dto.response.HindsightNoteResponse;
import com.integraltech.brainsentry.dto.response.SessionAnalysisResponse;
import com.integraltech.brainsentry.repository.AuditLogJpaRepository;
import com.integraltech.brainsentry.repository.HindsightNoteJpaRepository;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.repository.NoteJpaRepository;
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

import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
@DisplayName("NoteTakingService Unit Tests")
class NoteTakingServiceTest {

    private static final String TENANT_ID = "11111111-1111-1111-1111-111111111111";

    @Mock
    private HindsightNoteJpaRepository hindsightNoteRepo;

    @Mock
    private AuditLogJpaRepository auditLogRepo;

    @Mock
    private MemoryJpaRepository memoryRepo;

    @Mock
    private NoteJpaRepository noteRepo;

    @Mock
    private OpenRouterService openRouterService;

    @InjectMocks
    private NoteTakingService noteTakingService;

    @DisplayName("Session Analysis Tests")
    @Nested
    class SessionAnalysisTests {

        @Test
        @DisplayName("Should analyze session and return response")
        void testSessionAnalysis_ExtractsInformation() {
            // Given
            String sessionId = "test-session-123";
            String tenantId = TENANT_ID;

            SessionAnalysisRequest request = SessionAnalysisRequest.builder()
                .sessionId(sessionId)
                .tenantId(tenantId)
                .includeDecisions(true)
                .includeInsights(true)
                .includeFailures(true)
                .maxInsights(10)
                .build();

            when(auditLogRepo.findByTenantIdAndTimestampBetween(
                eq(tenantId), any(Instant.class), any(Instant.class)
            )).thenReturn(Collections.emptyList());

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{decisions: [{title: 'Test Decision'}]}");

            // When
            SessionAnalysisResponse response = noteTakingService.analyzeSession(request);

            // Then
            assertThat(response).isNotNull();
            assertThat(response.getSessionId()).isEqualTo(sessionId);
            assertThat(response.getTenantId()).isEqualTo(tenantId);

            verify(openRouterService).chat(anyString(), contains("decisions"));
        }

        @Test
        @DisplayName("Should handle session with no audit logs")
        void testSessionAnalysis_NoAuditLogs() {
            // Given
            SessionAnalysisRequest request = SessionAnalysisRequest.builder()
                .sessionId("empty-session")
                .tenantId(TENANT_ID)
                .build();

            when(auditLogRepo.findByTenantIdAndTimestampBetween(
                eq(TENANT_ID), any(Instant.class), any(Instant.class)
            )).thenReturn(Collections.emptyList());

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{decisions: [], insights: [], failures: []}");

            // When
            SessionAnalysisResponse response = noteTakingService.analyzeSession(request);

            // Then
            assertThat(response.getTotalDecisions()).isZero();
            assertThat(response.getTotalInsights()).isZero();
            assertThat(response.getTotalFailures()).isZero();
        }
    }

    @DisplayName("Hindsight Note Tests")
    @Nested
    class HindsightNoteTests {

        @Test
        @DisplayName("Should create new hindsight note")
        void testCreateHindsightNote_NewNote() {
            // Given
            CreateHindsightNoteRequest request = CreateHindsightNoteRequest.builder()
                .sessionId("session-123")
                .errorType("NullPointerException")
                .errorMessage("Null pointer in UserService")
                .resolution("Added null check")
                .preventionStrategy("Always validate inputs")
                .priority("HIGH")
                .tags(Arrays.asList("bug", "npe"))
                .build();

            when(hindsightNoteRepo.findSimilarErrors(
                anyString(), eq("NullPointerException"), anyString()
            )).thenReturn(Collections.emptyList());

            HindsightNote savedNote = HindsightNote.builder()
                .id("note-123")
                .errorType("NullPointerException")
                .errorMessage("Null pointer in UserService")
                .resolution("Added null check")
                .preventionStrategy("Always validate inputs")
                .priority("HIGH")
                .tags(Arrays.asList("bug", "npe"))
                .occurrenceCount(1)
                .build();

            when(hindsightNoteRepo.save(any(HindsightNote.class))).thenReturn(savedNote);

            // When
            HindsightNoteResponse response = noteTakingService.createHindsightNote(request);

            // Then
            assertThat(response).isNotNull();
            assertThat(response.getErrorType()).isEqualTo("NullPointerException");
            assertThat(response.getPriority()).isEqualTo("HIGH");

            verify(hindsightNoteRepo).save(any(HindsightNote.class));
        }

        @Test
        @DisplayName("Should update existing hindsight note when similar error found")
        void testCreateHindsightNote_ExistingNote() {
            // Given
            CreateHindsightNoteRequest request = CreateHindsightNoteRequest.builder()
                .sessionId("session-123")
                .errorType("NullPointerException")
                .errorMessage("Null pointer in UserService")
                .build();

            HindsightNote existingNote = HindsightNote.builder()
                .id("existing-note")
                .errorType("NullPointerException")
                .errorMessage("Null pointer")
                .occurrenceCount(1)
                .build();

            when(hindsightNoteRepo.findSimilarErrors(
                anyString(), eq("NullPointerException"), anyString()
            )).thenReturn(Arrays.asList(existingNote));

            when(hindsightNoteRepo.save(any(HindsightNote.class))).thenReturn(existingNote);

            // When
            HindsightNoteResponse response = noteTakingService.createHindsightNote(request);

            // Then
            assertThat(response).isNotNull();
            verify(hindsightNoteRepo).save(existingNote);
        }

        @Test
        @DisplayName("Should get hindsight notes for tenant")
        void testGetHindsightNotes() {
            // Given
            String tenantId = TENANT_ID;

            HindsightNote note1 = HindsightNote.builder()
                .id("note-1")
                .tenantId(tenantId)
                .errorType("Error1")
                .build();

            HindsightNote note2 = HindsightNote.builder()
                .id("note-2")
                .tenantId(tenantId)
                .errorType("Error2")
                .build();

            when(hindsightNoteRepo.findByTenantId(tenantId))
                .thenReturn(Arrays.asList(note1, note2));

            // When
            List<HindsightNoteResponse> notes = noteTakingService.getHindsightNotes(tenantId);

            // Then
            assertThat(notes).hasSize(2);
            assertThat(notes.get(0).getErrorType()).isEqualTo("Error1");
            assertThat(notes.get(1).getErrorType()).isEqualTo("Error2");
        }

        @Test
        @DisplayName("Should get frequent errors")
        void testGetFrequentErrors() {
            // Given
            String tenantId = TENANT_ID;

            HindsightNote frequentNote = HindsightNote.builder()
                .id("note-1")
                .tenantId(tenantId)
                .errorType("FrequentError")
                .occurrenceCount(5)
                .build();

            when(hindsightNoteRepo.findByTenantIdAndOccurrenceCountGreaterThan(tenantId, 3))
                .thenReturn(Arrays.asList(frequentNote));

            // When
            List<HindsightNoteResponse> notes = noteTakingService.getFrequentErrors(tenantId);

            // Then
            assertThat(notes).hasSize(1);
            assertThat(notes.get(0).getErrorType()).isEqualTo("FrequentError");
            assertThat(notes.get(0).getOccurrenceCount()).isEqualTo(5);
        }
    }

    @DisplayName("Markdown Export Tests")
    @Nested
    class MarkdownExportTests {

        @Test
        @DisplayName("Should generate markdown summary")
        void testGenerateMarkdownSummary() {
            // Given
            String sessionId = "session-123";
            String tenantId = TENANT_ID;

            when(auditLogRepo.findByTenantIdAndTimestampBetween(
                eq(tenantId), any(Instant.class), any(Instant.class)
            )).thenReturn(Collections.emptyList());

            when(hindsightNoteRepo.findBySessionId(sessionId))
                .thenReturn(Collections.emptyList());

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("# Session Summary\n\n## Overview\nNo activity.\n");

            // When
            String markdown = noteTakingService.generateMarkdownSummary(sessionId, tenantId);

            // Then
            assertThat(markdown).isNotNull();
            assertThat(markdown).contains("# Session Summary");
        }
    }

    @DisplayName("Session Distillation Tests")
    @Nested
    class SessionDistillationTests {

        @Test
        @DisplayName("Should distill session into memories")
        void testDistillSession_CreatesMemories() {
            // Given
            String sessionId = "session-123";
            String tenantId = TENANT_ID;

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{decisions: [], insights: [{category: PATTERN, content: Test pattern, importance: HIGH, relatedTo: Test}], failures: []}");

            Memory savedMemory = Memory.builder()
                .id("memory-123")
                .content("Test pattern")
                .build();

            when(memoryRepo.save(any(Memory.class)))
                .thenReturn(savedMemory);

            // When
            List<Memory> memories =
                noteTakingService.distillSession(sessionId, tenantId);

            // Then
            assertThat(memories).isNotNull();
            verify(memoryRepo, atLeastOnce()).save(any(Memory.class));
        }
    }

    @DisplayName("Confucius Spec Methods Tests")
    @Nested
    class ConfuciusSpecMethodsTests {

        @Test
        @DisplayName("Should extract insights from session")
        void testExtractInsights() {
            // Given
            String sessionId = "session-123";
            String tenantId = TENANT_ID;

            when(auditLogRepo.findByTenantIdAndTimestampBetween(
                eq(tenantId), any(Instant.class), any(Instant.class)
            )).thenReturn(Collections.emptyList());

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{insights: [{title: 'Test Insight', content: 'Test content', category: 'PATTERN', keywords: ['test']}]");

            Note savedNote = Note.builder()
                .id("note-123")
                .title("Session Insight")
                .build();

            when(noteRepo.save(any(Note.class))).thenReturn(savedNote);

            // When
            List<Note> insights = noteTakingService.extractInsights(sessionId, tenantId);

            // Then
            assertThat(insights).isNotNull();
            verify(noteRepo).save(any(Note.class));
        }

        @Test
        @DisplayName("Should extract hindsights from errors")
        void testExtractHindsights() {
            // Given
            String sessionId = "session-123";
            String tenantId = TENANT_ID;

            when(auditLogRepo.findByTenantIdAndTimestampBetween(
                eq(tenantId), any(Instant.class), any(Instant.class)
            )).thenReturn(Collections.emptyList());

            Note savedNote = Note.builder()
                .id("note-456")
                .type(com.integraltech.brainsentry.domain.enums.NoteType.HINDSIGHT)
                .build();

            when(noteRepo.save(any(Note.class))).thenReturn(savedNote);

            // When
            List<Note> hindsights = noteTakingService.extractHindsights(sessionId, tenantId);

            // Then
            assertThat(hindsights).isNotNull();
        }

        @Test
        @DisplayName("Should identify patterns from session")
        void testIdentifyPatterns() {
            // Given
            String sessionId = "session-123";
            String tenantId = TENANT_ID;

            when(auditLogRepo.findByTenantIdAndTimestampBetween(
                eq(tenantId), any(Instant.class), any(Instant.class)
            )).thenReturn(Collections.emptyList());

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{patterns: [{title: 'Test Pattern', description: 'Found pattern', type: 'DESIGN', keywords: ['pattern']}]");

            Note savedNote = Note.builder()
                .id("note-789")
                .build();

            when(noteRepo.save(any(Note.class))).thenReturn(savedNote);

            // When
            List<Note> patterns = noteTakingService.identifyPatterns(sessionId, tenantId);

            // Then
            assertThat(patterns).isNotNull();
            verify(noteRepo).save(any(Note.class));
        }

        @Test
        @DisplayName("Should extract architectural decisions")
        void testExtractArchitecturalDecisions() {
            // Given
            String sessionId = "session-123";
            String tenantId = TENANT_ID;

            when(auditLogRepo.findByTenantIdAndTimestampBetween(
                eq(tenantId), any(Instant.class), any(Instant.class)
            )).thenReturn(Collections.emptyList());

            when(openRouterService.chat(anyString(), anyString()))
                .thenReturn("{decisions: [{title: 'Use PostgreSQL', rationale: 'Scalability', impact: 'High', alternatives: ['MySQL']}]");

            Note savedNote = Note.builder()
                .id("note-arch")
                .build();

            when(noteRepo.save(any(Note.class))).thenReturn(savedNote);

            // When
            List<Note> decisions = noteTakingService.extractArchitecturalDecisions(sessionId, tenantId);

            // Then
            assertThat(decisions).isNotNull();
            verify(noteRepo).save(any(Note.class));
        }

        @Test
        @DisplayName("Should link notes to memories")
        void testLinkToMemories() {
            // Given
            String noteId = "note-123";
            List<String> memoryIds = Arrays.asList("mem-1", "mem-2");

            Note existingNote = Note.builder()
                .id(noteId)
                .relatedMemoryIds(new ArrayList<>())
                .build();

            when(noteRepo.findById(noteId)).thenReturn(java.util.Optional.of(existingNote));

            Note updatedNote = Note.builder()
                .id(noteId)
                .relatedMemoryIds(memoryIds)
                .build();

            when(noteRepo.save(any(Note.class))).thenReturn(updatedNote);

            // When
            noteTakingService.linkToMemories(noteId, memoryIds);

            // Then
            verify(noteRepo).save(any(Note.class));
        }

        @Test
        @DisplayName("Should generate session summary")
        void testGenerateSessionSummary() {
            // Given
            String sessionId = "session-456";
            String tenantId = "tenant-xyz";

            List<Note> notes = Arrays.asList(
                Note.builder()
                    .id("note-1")
                    .type(com.integraltech.brainsentry.domain.enums.NoteType.INSIGHT)
                    .title("Insight 1")
                    .content("Content 1")
                    .build(),
                Note.builder()
                    .id("note-2")
                    .type(com.integraltech.brainsentry.domain.enums.NoteType.HINDSIGHT)
                    .title("Hindsight 1")
                    .content("Resolution 1")
                    .build()
            );

            Note savedSummary = Note.builder()
                .id("summary-123")
                .title("Session Summary: session-456")
                .build();

            when(noteRepo.save(any(Note.class))).thenAnswer(invocation -> invocation.getArgument(0));

            // When
            Note summary = noteTakingService.generateSessionSummary(sessionId, notes);

            // Then
            assertThat(summary).isNotNull();
            assertThat(summary.getTitle()).contains("Session Summary");
            assertThat(summary.getContent()).contains("## Insights Captured");
            assertThat(summary.getContent()).contains("## Failures & Resolutions");
            verify(noteRepo).save(any(Note.class));
        }
    }
}
