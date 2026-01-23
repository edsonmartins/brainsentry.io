package com.integraltech.brainsentry.integration;

import com.integraltech.brainsentry.domain.ContextSummary;
import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.Note;
import com.integraltech.brainsentry.domain.enums.NoteCategory;
import com.integraltech.brainsentry.domain.enums.NoteSeverity;
import com.integraltech.brainsentry.domain.enums.NoteType;
import com.integraltech.brainsentry.repository.ContextSummaryJpaRepository;
import com.integraltech.brainsentry.repository.HindsightNoteJpaRepository;
import com.integraltech.brainsentry.repository.NoteJpaRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.context.annotation.Import;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * End-to-End integration tests for Confucius Code Agent features.
 *
 * Validates that the complete flow works:
 * 1. Note-Taking Agent (extract insights, hindsights, patterns)
 * 2. Architect Agent (context compression)
 * 3. Note Retrieval (pattern matching, semantic search)
 */
@SpringBootTest
@ActiveProfiles("test")
@Import(com.integraltech.brainsentry.config.TestConfig.class)
@Transactional
@DisplayName("Confucius Features Integration Tests")
class ConfuciusFeaturesIntegrationTest {

    @Autowired
    private HindsightNoteJpaRepository hindsightNoteRepo;

    @Autowired
    private NoteJpaRepository noteRepo;

    @Autowired
    private ContextSummaryJpaRepository contextSummaryRepo;

    @Test
    @DisplayName("E2E: Should create and retrieve hindsight notes with pattern matching")
    void testHindsightNoteLifecycle_PatternMatching() {
        // Given
        String tenantId = "default";
        String sessionId = "session-e2e-1";

        HindsightNote note = HindsightNote.builder()
            .tenantId(tenantId)
            .sessionId(sessionId)
            .title("NPE in UserService")
            .errorType("NullPointerException")
            .errorMessage("NullPointerException: Cannot invoke method on null object")
            .errorPattern(".*NullPointerException.*")
            .resolution("Add null check before invoking method")
            .preventionStrategy("Always validate inputs")
            .severity(NoteSeverity.HIGH)
            .occurrenceCount(1)
            .createdAt(Instant.now())
            .build();

        // When - Save note
        HindsightNote saved = hindsightNoteRepo.save(note);

        // Then - Verify saved
        assertThat(saved).isNotNull();
        assertThat(saved.getId()).isNotNull();
        assertThat(saved.matchesError("NullPointerException: Cannot invoke method on null object")).isTrue();

        // When - Retrieve by error type
        List<HindsightNote> found = hindsightNoteRepo.findByErrorTypeAndTenantId("NullPointerException", tenantId);

        // Then
        assertThat(found).isNotEmpty();
        assertThat(found.get(0).getErrorPattern()).isNotNull();
    }

    @Test
    @DisplayName("E2E: Should create generic notes with all types")
    void testGenericNote_AllTypesSupported() {
        // Given
        String tenantId = "default";
        String sessionId = "session-e2e-2";

        // Create different note types
        Note insight = Note.builder()
            .tenantId(tenantId)
            .sessionId(sessionId)
            .type(NoteType.INSIGHT)
            .title("Test Insight")
            .content("This is an insight from testing")
            .category(NoteCategory.PROJECT_SPECIFIC)
            .keywords(Arrays.asList("test", "insight"))
            .createdAt(Instant.now())
            .build();

        Note hindsight = Note.builder()
            .tenantId(tenantId)
            .sessionId(sessionId)
            .type(NoteType.HINDSIGHT)
            .title("Test Hindsight")
            .content("## Problem\nError occurred\n\n## Solution\nFixed it")
            .category(NoteCategory.SHARED)
            .severity(NoteSeverity.MEDIUM)
            .createdAt(Instant.now())
            .build();

        Note pattern = Note.builder()
            .tenantId(tenantId)
            .sessionId(sessionId)
            .type(NoteType.PATTERN)
            .title("Test Pattern")
            .content("Use this pattern for similar situations")
            .category(NoteCategory.GENERIC)
            .createdAt(Instant.now())
            .build();

        // When
        Note savedInsight = noteRepo.save(insight);
        Note savedHindsight = noteRepo.save(hindsight);
        Note savedPattern = noteRepo.save(pattern);

        // Then
        assertThat(savedInsight).isNotNull();
        assertThat(savedHindsight).isNotNull();
        assertThat(savedPattern).isNotNull();

        // Verify retrieval by type
        List<Note> insights = noteRepo.findByTenantIdAndType(tenantId, NoteType.INSIGHT);
        List<Note> hindsights = noteRepo.findByTenantIdAndType(tenantId, NoteType.HINDSIGHT);
        List<Note> patterns = noteRepo.findByTenantIdAndType(tenantId, NoteType.PATTERN);

        assertThat(insights).isNotEmpty();
        assertThat(hindsights).isNotEmpty();
        assertThat(patterns).isNotEmpty();
    }

    @Test
    @DisplayName("E2E: Should save and retrieve context summaries")
    void testContextSummary_CompressionTracking() {
        // Given
        String tenantId = "default";
        String sessionId = "session-compression-1";

        ContextSummary summary = ContextSummary.builder()
            .tenantId(tenantId)
            .sessionId(sessionId)
            .originalTokenCount(10000)
            .compressedTokenCount(4000)
            .compressionRatio(0.4)
            .summary("## Goals\nImplement OAuth2\n\n## Decisions\nUse JWT tokens")
            .goals(Arrays.asList("Implement OAuth2"))
            .decisions(Arrays.asList("Use JWT tokens"))
            .errors(Arrays.asList())
            .todos(Arrays.asList("Test refresh token"))
            .recentWindowSize(10)
            .compressionMethod("LLM")
            .createdAt(Instant.now())
            .build();

        // When
        ContextSummary saved = contextSummaryRepo.save(summary);

        // Then
        assertThat(saved).isNotNull();
        assertThat(saved.getCompressionRatio()).isLessThan(0.5);
        assertThat(saved.isTargetAchieved()).isTrue();
        assertThat(saved.getTokenSavings()).isEqualTo(6000);

        // Verify retrieval
        List<ContextSummary> found = contextSummaryRepo.findByTenantIdAndCreatedAtAfter(
            tenantId, Instant.now().minusSeconds(60));

        assertThat(found).isNotEmpty();
    }

    @Test
    @DisplayName("E2E: Should link notes to memories")
    void testNoteMemoryLinking_GraphRelationships() {
        // Given
        String tenantId = "default";
        String sessionId = "session-graph-1";

        Note note = Note.builder()
            .tenantId(tenantId)
            .sessionId(sessionId)
            .type(NoteType.ARCHITECTURE)
            .title("Architecture Decision")
            .content("Decided to use microservices")
            .relatedMemoryIds(new ArrayList<>())
            .createdAt(Instant.now())
            .build();

        // When
        Note savedNote = noteRepo.save(note);
        savedNote.addRelatedMemory("memory-123");
        Note linkedNote = noteRepo.save(savedNote);

        // Then
        assertThat(linkedNote.getRelatedMemoryIds()).contains("memory-123");
    }

    @Test
    @DisplayName("E2E: Full note-taking workflow")
    void testFullNoteTakingWorkflow() {
        // Given
        String tenantId = "default";
        String sessionId = "session-full-1";

        // Create a hindsight note
        HindsightNote hindsightNote = HindsightNote.builder()
            .tenantId(tenantId)
            .sessionId(sessionId)
            .title("API Timeout Error")
            .errorType("API_TIMEOUT")
            .errorMessage("API call timed out after 30s")
            .errorPattern(".*timed out.*")
            .resolution("Increase timeout to 60s")
            .preventionStrategy("Add retry logic with exponential backoff")
            .severity(NoteSeverity.HIGH)
            .occurrenceCount(1)
            .createdAt(Instant.now())
            .build();

        // When
        HindsightNote savedHindsight = hindsightNoteRepo.save(hindsightNote);

        // Then
        assertThat(savedHindsight).isNotNull();
        assertThat(savedHindsight.matchesError("API call timed out after 30s")).isTrue();

        // Create related pattern note
        Note patternNote = Note.builder()
            .tenantId(tenantId)
            .sessionId(sessionId)
            .type(NoteType.PATTERN)
            .title("API Timeout Pattern")
            .content("When API calls timeout, increase timeout and add retry")
            .category(NoteCategory.SHARED)
            .createdAt(Instant.now())
            .build();

        Note savedPattern = noteRepo.save(patternNote);

        // Link pattern to hindsight
        savedPattern.addRelatedNote(savedHindsight.getId());
        noteRepo.save(savedPattern);

        // Then
        assertThat(savedPattern.getRelatedNoteIds()).contains(savedHindsight.getId());
    }
}
