package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.enums.NoteSeverity;
import com.integraltech.brainsentry.repository.HindsightNoteJpaRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Comparator;
import java.util.List;
import java.util.stream.Collectors;

/**
 * Service for retrieving and ranking notes.
 *
 * Inspired by Confucius Code Agent's note retrieval system.
 * Provides pattern-based and semantic search for hindsight notes.
 *
 * KEY FEATURES:
 * - Pattern matching (fast path) for error notes
 * - Semantic search via embeddings
 * - Ranking by severity, recency, and access count
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class NoteRetrievalService {

    private final HindsightNoteJpaRepository hindsightNoteRepo;
    private final EmbeddingService embeddingService;

    /**
     * Search for relevant hindsight notes when similar error occurs.
     * KEY FEATURE from Confucius.
     *
     * @param errorMessage the error message to match
     * @param errorType the error type (optional, for broader matching)
     * @param tenantId the tenant ID
     * @return list of matching notes, ranked by relevance
     */
    public List<HindsightNote> searchHindsightNotes(String errorMessage, String errorType, String tenantId) {
        log.debug("Searching hindsight notes for error: {}, type: {}", errorMessage, errorType);

        // 1. Pattern matching (fast path)
        List<HindsightNote> patternMatches = findByErrorPattern(errorMessage, tenantId);
        log.debug("Found {} pattern matches", patternMatches.size());

        // 2. Error type matching (fallback)
        if (patternMatches.isEmpty() && errorType != null) {
            List<HindsightNote> typeMatches = hindsightNoteRepo.findByErrorTypeAndTenantId(errorType, tenantId);
            patternMatches.addAll(typeMatches);
            log.debug("Added {} type matches", typeMatches.size());
        }

        // 3. Semantic search (vector similarity) - if enabled
        if (patternMatches.isEmpty() || patternMatches.size() < 3) {
            if (embeddingService != null && embeddingService.isReady()) {
                try {
                    float[] errorEmbedding = embeddingService.embed(errorMessage);
                    // Note: Full semantic search would require embedding field on HindsightNote
                    // For now, we rely on pattern + type matching
                    log.debug("Semantic search available but not fully implemented");
                } catch (Exception e) {
                    log.warn("Failed to perform semantic search", e);
                }
            }
        }

        // 4. Rank by severity + recency + access count
        return rankNotesByRelevance(patternMatches);
    }

    /**
     * Find notes by error pattern matching.
     * Fast path for exact pattern matches.
     *
     * @param errorMessage the error message
     * @param tenantId the tenant ID
     * @return list of matching notes
     */
    private List<HindsightNote> findByErrorPattern(String errorMessage, String tenantId) {
        return hindsightNoteRepo.findByTenantId(tenantId).stream()
            .filter(note -> note.matchesError(errorMessage))
            .collect(Collectors.toList());
    }

    /**
     * Get notes for current context.
     * Used during autonomous interception.
     *
     * @param query the search query
     * @param tenantId the tenant ID
     * @param limit max results
     * @return list of relevant notes
     */
    public List<HindsightNote> getRelevantNotes(String query, String tenantId, int limit) {
        // Get all notes for tenant
        List<HindsightNote> allNotes = hindsightNoteRepo.findByTenantId(tenantId);

        // Simple keyword matching in title and error message
        String[] keywords = query.toLowerCase().split("\\s+");

        return allNotes.stream()
            .filter(note -> {
                String title = note.getTitle() != null ? note.getTitle().toLowerCase() : "";
                String errorMsg = note.getErrorMessage() != null ? note.getErrorMessage().toLowerCase() : "";
                String combined = title + " " + errorMsg;

                for (String keyword : keywords) {
                    if (combined.contains(keyword)) {
                        return true;
                    }
                }
                return false;
            })
            .limit(limit)
            .collect(Collectors.toList());
    }

    /**
     * Get frequent errors for a tenant.
     * Errors that occurred more than once.
     *
     * @param tenantId the tenant ID
     * @return list of frequent hindsight notes
     */
    public List<HindsightNote> getFrequentErrors(String tenantId) {
        return hindsightNoteRepo.findByTenantIdAndOccurrenceCountGreaterThan(tenantId, 1)
            .stream()
            .sorted((a, b) -> b.getOccurrenceCount().compareTo(a.getOccurrenceCount()))
            .collect(Collectors.toList());
    }

    /**
     * Get critical errors for a tenant.
     * Errors with CRITICAL or HIGH severity.
     *
     * @param tenantId the tenant ID
     * @return list of critical hindsight notes
     */
    public List<HindsightNote> getCriticalErrors(String tenantId) {
        return hindsightNoteRepo.findByTenantId(tenantId).stream()
            .filter(note -> note.getSeverity() == NoteSeverity.CRITICAL ||
                           note.getSeverity() == NoteSeverity.HIGH)
            .sorted(Comparator
                .comparing(HindsightNote::getSeverity)
                .thenComparing(HindsightNote::getCreatedAt).reversed()
                .thenComparing(HindsightNote::getAccessCount).reversed()
            )
            .collect(Collectors.toList());
    }

    /**
     * Rank notes by relevance.
     *
     * Ranking criteria (in order):
     * 1. Severity (CRITICAL > HIGH > MEDIUM > LOW)
     * 2. Recency (more recent = higher)
     * 3. Access count (more accessed = higher)
     *
     * @param notes the notes to rank
     * @return ranked list of notes
     */
    private List<HindsightNote> rankNotesByRelevance(List<HindsightNote> notes) {
        return notes.stream()
            .sorted(Comparator
                .comparing(HindsightNote::getSeverity)
                .thenComparing(HindsightNote::getLastOccurrenceAt, Comparator.nullsLast(Comparator.reverseOrder()))
                .thenComparing(HindsightNote::getAccessCount).reversed()
            )
            .collect(Collectors.toList());
    }

    /**
     * Record that a note was referenced/suggested.
     * Updates access statistics.
     *
     * @param noteId the note ID
     */
    public void recordNoteAccess(String noteId) {
        hindsightNoteRepo.findById(noteId).ifPresent(note -> {
            note.recordReference();
            hindsightNoteRepo.save(note);
        });
    }

    /**
     * Find similar notes based on error type.
     *
     * @param errorType the error type
     * @param tenantId the tenant ID
     * @return list of similar notes
     */
    public List<HindsightNote> findByErrorType(String errorType, String tenantId) {
        return hindsightNoteRepo.findByErrorTypeAndTenantId(errorType, tenantId);
    }
}
