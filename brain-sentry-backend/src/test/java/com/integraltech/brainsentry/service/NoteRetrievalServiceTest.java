package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.enums.NoteSeverity;
import com.integraltech.brainsentry.dto.request.CreateHindsightNoteRequest;
import com.integraltech.brainsentry.dto.response.HindsightNoteResponse;
import com.integraltech.brainsentry.repository.HindsightNoteJpaRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@DisplayName("NoteRetrievalService Unit Tests")
class NoteRetrievalServiceTest {

    @Mock
    private HindsightNoteJpaRepository hindsightNoteRepo;

    @Mock
    private EmbeddingService embeddingService;

    @InjectMocks
    private NoteRetrievalService noteRetrievalService;

    @DisplayName("Pattern Matching Tests")
    @Nested
    class PatternMatchingTests {

        @Test
        @DisplayName("Should find notes by error pattern")
        void testSearchHindsightNotes_PatternMatch() {
            // Given
            String errorMessage = "NullPointerException: Cannot invoke method on null object";
            String errorType = "NullPointerException";
            String tenantId = "tenant-123";

            HindsightNote note = HindsightNote.builder()
                .id("note-1")
                .tenantId(tenantId)
                .errorType(errorType)
                .errorMessage("NPE in UserService")
                .errorPattern(".*NullPointerException.*")
                .severity(NoteSeverity.HIGH)
                .build();

            when(hindsightNoteRepo.findByTenantId(tenantId))
                .thenReturn(Arrays.asList(note));

            // When
            List<HindsightNote> results = noteRetrievalService.searchHindsightNotes(
                errorMessage, errorType, tenantId
            );

            // Then
            assertThat(results).hasSize(1);
            assertThat(results.get(0).matchesError(errorMessage)).isTrue();
        }

        @Test
        @DisplayName("Should fallback to error type matching when pattern fails")
        void testSearchHindsightNotes_TypeMatchFallback() {
            // Given
            String errorMessage = "Unknown error";
            String errorType = "API_TIMEOUT";
            String tenantId = "tenant-123";

            HindsightNote note = HindsightNote.builder()
                .id("note-1")
                .tenantId(tenantId)
                .errorType(errorType)
                .errorMessage("API timeout occurred")
                .severity(NoteSeverity.MEDIUM)
                .build();

            when(hindsightNoteRepo.findByTenantId(tenantId))
                .thenReturn(Arrays.asList(note));
            when(hindsightNoteRepo.findByErrorTypeAndTenantId(errorType, tenantId))
                .thenReturn(Arrays.asList(note));

            // When
            List<HindsightNote> results = noteRetrievalService.searchHindsightNotes(
                errorMessage, errorType, tenantId
            );

            // Then
            assertThat(results).hasSize(1);
            assertThat(results.get(0).matchesErrorType(errorType)).isTrue();
        }
    }

    @DisplayName("Critical Errors Tests")
    @Nested
    class CriticalErrorsTests {

        @Test
        @DisplayName("Should return only CRITICAL and HIGH severity notes")
        void testGetCriticalErrors_FiltersBySeverity() {
            // Given
            String tenantId = "tenant-123";

            HindsightNote criticalNote = HindsightNote.builder()
                .id("note-1")
                .tenantId(tenantId)
                .severity(NoteSeverity.CRITICAL)
                .build();

            HindsightNote highNote = HindsightNote.builder()
                .id("note-2")
                .tenantId(tenantId)
                .severity(NoteSeverity.HIGH)
                .build();

            HindsightNote mediumNote = HindsightNote.builder()
                .id("note-3")
                .tenantId(tenantId)
                .severity(NoteSeverity.MEDIUM)
                .build();

            when(hindsightNoteRepo.findByTenantId(tenantId))
                .thenReturn(Arrays.asList(criticalNote, highNote, mediumNote));

            // When
            List<HindsightNote> results = noteRetrievalService.getCriticalErrors(tenantId);

            // Then
            assertThat(results).hasSize(2);
            assertThat(results).extracting("severity")
                .containsExactly(NoteSeverity.CRITICAL, NoteSeverity.HIGH);
        }
    }

    @DisplayName("Relevant Notes Tests")
    @Nested
    class RelevantNotesTests {

        @Test
        @DisplayName("Should find notes by keyword in title or message")
        void testGetRelevantNotes_KeywordMatch() {
            // Given
            String tenantId = "tenant-123";
            String query = "oauth authentication";

            HindsightNote note1 = HindsightNote.builder()
                .id("note-1")
                .tenantId(tenantId)
                .title("OAuth Token Validation Error")
                .errorMessage("Failed to validate OAuth token")
                .build();

            HindsightNote note2 = HindsightNote.builder()
                .id("note-2")
                .tenantId(tenantId)
                .title("Database Connection Issue")
                .errorMessage("Connection timeout")
                .build();

            when(hindsightNoteRepo.findByTenantId(tenantId))
                .thenReturn(Arrays.asList(note1, note2));

            // When
            List<HindsightNote> results = noteRetrievalService.getRelevantNotes(query, tenantId, 10);

            // Then
            assertThat(results).hasSize(1);
            assertThat(results.get(0).getId()).isEqualTo("note-1");
        }

        @Test
        @DisplayName("Should respect limit parameter")
        void testGetRelevantNotes_RespectsLimit() {
            // Given
            String tenantId = "tenant-123";

            when(hindsightNoteRepo.findByTenantId(tenantId))
                .thenReturn(Arrays.asList(
                    HindsightNote.builder().id("note-1").tenantId(tenantId).title("Note 1").build(),
                    HindsightNote.builder().id("note-2").tenantId(tenantId).title("Note 2").build(),
                    HindsightNote.builder().id("note-3").tenantId(tenantId).title("Note 3").build()
                ));

            // When
            List<HindsightNote> results = noteRetrievalService.getRelevantNotes("note", tenantId, 2);

            // Then
            assertThat(results).hasSize(2);
        }
    }

    @DisplayName("Frequent Errors Tests")
    @Nested
    class FrequentErrorsTests {

        @Test
        @DisplayName("Should return errors with occurrence count > 1")
        void testGetFrequentErrors_OnlyFrequent() {
            // Given
            String tenantId = "tenant-123";

            HindsightNote frequentNote = HindsightNote.builder()
                .id("note-1")
                .tenantId(tenantId)
                .occurrenceCount(5)
                .build();

            HindsightNote singleNote = HindsightNote.builder()
                .id("note-2")
                .tenantId(tenantId)
                .occurrenceCount(1)
                .build();

            when(hindsightNoteRepo.findByTenantIdAndOccurrenceCountGreaterThan(tenantId, 1))
                .thenReturn(Arrays.asList(frequentNote));

            // When
            List<HindsightNote> results = noteRetrievalService.getFrequentErrors(tenantId);

            // Then
            assertThat(results).hasSize(1);
            assertThat(results.get(0).getOccurrenceCount()).isEqualTo(5);
        }
    }

    @DisplayName("Note Access Recording Tests")
    @Nested
    class NoteAccessTests {

        @Test
        @DisplayName("Should record note access")
        void testRecordNoteAccess_UpdatesStats() {
            // Given
            String noteId = "note-123";

            HindsightNote note = HindsightNote.builder()
                .id(noteId)
                .referenceCount(2)
                .accessCount(2)
                .build();

            when(hindsightNoteRepo.findById(noteId))
                .thenReturn(java.util.Optional.of(note));

            // When
            noteRetrievalService.recordNoteAccess(noteId);

            // Then
            verify(hindsightNoteRepo).save(argThat(savedNote -> {
                return savedNote.getReferenceCount() == 3 && savedNote.getAccessCount() == 3;
            }));
        }
    }
}
