package com.integraltech.brainsentry.service;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.Executor;

/**
 * Service for generating text embeddings.
 *
 * Uses all-MiniLM-L6-v2 model for 384-dimensional embeddings.
 * This implementation uses a placeholder - actual integration
 * with DJL/ONNX will be added in Phase 2.
 */
@Slf4j
@Service
public class EmbeddingService {

    private static final int EMBEDDING_DIMENSIONS = 384;

    /**
     * Generate embedding for a single text.
     *
     * @param text the text to embed
     * @return 384-dimensional float array
     */
    public float[] embed(String text) {
        // TODO: Integrate with DJL/ONNX for all-MiniLM-L6-v2
        // For now, return a deterministic placeholder based on text hash
        log.debug("Generating embedding for text (length: {})", text.length());

        float[] embedding = new float[EMBEDDING_DIMENSIONS];

        // Simple hash-based placeholder (not for production!)
        int hash = text.hashCode();
        for (int i = 0; i < EMBEDDING_DIMENSIONS; i++) {
            // Deterministic pseudo-random values based on hash and position
            double val = Math.sin(hash * (i + 1) * 0.1) * 0.5 + 0.5;
            embedding[i] = (float) val;
        }

        // Normalize to unit length
        float norm = 0.0f;
        for (float v : embedding) {
            norm += v * v;
        }
        norm = (float) Math.sqrt(norm);
        if (norm > 0) {
            for (int i = 0; i < embedding.length; i++) {
                embedding[i] /= norm;
            }
        }

        return embedding;
    }

    /**
     * Generate embeddings for multiple texts in parallel.
     *
     * @param texts list of texts to embed
     * @return list of embeddings
     */
    public List<float[]> embedBatch(List<String> texts) {
        List<float[]> embeddings = new ArrayList<>(texts.size());

        // Process in parallel using virtual threads (configured via @Async)
        List<CompletableFuture<float[]>> futures = texts.stream()
            .map(text -> CompletableFuture.supplyAsync(() -> embed(text)))
            .toList();

        CompletableFuture.allOf(futures.toArray(new CompletableFuture[0])).join();

        for (CompletableFuture<float[]> future : futures) {
            embeddings.add(future.join());
        }

        return embeddings;
    }

    /**
     * Calculate cosine similarity between two embeddings.
     *
     * @param a first embedding
     * @param b second embedding
     * @return similarity score (0.0 to 1.0)
     */
    public float cosineSimilarity(float[] a, float[] b) {
        if (a.length != b.length) {
            throw new IllegalArgumentException("Embedding dimensions must match");
        }

        float dotProduct = 0.0f;
        float normA = 0.0f;
        float normB = 0.0f;

        for (int i = 0; i < a.length; i++) {
            dotProduct += a[i] * b[i];
            normA += a[i] * a[i];
            normB += b[i] * b[i];
        }

        return dotProduct / ((float) Math.sqrt(normA) * (float) Math.sqrt(normB));
    }

    /**
     * Check if the embedding service is ready.
     *
     * @return true if the service can generate embeddings
     */
    public boolean isReady() {
        // Service is always ready with the placeholder implementation
        return true;
    }

    /**
     * Get the embedding dimension.
     *
     * @return the dimension of embeddings
     */
    public int getDimension() {
        return EMBEDDING_DIMENSIONS;
    }
}
