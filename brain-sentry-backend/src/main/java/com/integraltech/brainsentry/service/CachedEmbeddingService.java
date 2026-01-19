package com.integraltech.brainsentry.service;

import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.cache.Cache;
import org.springframework.cache.CacheManager;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CompletableFuture;

/**
 * Cached version of EmbeddingService for improved performance.
 *
 * Caches embeddings in Redis to avoid recomputation.
 */
@Slf4j
@Service
@ConditionalOnProperty(name = "brainsentry.redis.host")
public class CachedEmbeddingService extends EmbeddingService {

    private final CacheManager cacheManager;
    private static final String CACHE_NAME = "embeddings";

    public CachedEmbeddingService(CacheManager cacheManager) {
        this.cacheManager = cacheManager;
    }

    @Override
    public float[] embed(String text) {
        // Create a cache key from text hash
        String cacheKey = "embedding:" + text.hashCode();

        Cache cache = cacheManager.getCache(CACHE_NAME);
        if (cache != null) {
            // Try to get from cache
            Float[] cached = cache.get(cacheKey, Float[].class);
            if (cached != null) {
                // Convert Float[] back to float[]
                float[] result = new float[cached.length];
                for (int i = 0; i < cached.length; i++) {
                    result[i] = cached[i];
                }
                log.debug("Cache hit for text hash: {}", text.hashCode());
                return result;
            }
        }

        // Cache miss - generate embedding
        log.debug("Cache miss for text hash: {}, generating embedding", text.hashCode());
        float[] embedding = super.embed(text);

        // Store in cache
        if (cache != null) {
            // Convert float[] to Float[] for caching
            Float[] toCache = new Float[embedding.length];
            for (int i = 0; i < embedding.length; i++) {
                toCache[i] = embedding[i];
            }
            cache.put(cacheKey, toCache);
        }

        return embedding;
    }

    @Override
    public List<float[]> embedBatch(List<String> texts) {
        // For batch, check cache first for each text
        List<float[]> embeddings = new ArrayList<>(texts.size());
        List<Integer> missedIndices = new ArrayList<>();

        for (int i = 0; i < texts.size(); i++) {
            String text = texts.get(i);
            String cacheKey = "embedding:" + text.hashCode();

            Cache cache = cacheManager.getCache(CACHE_NAME);
            if (cache != null) {
                Float[] cached = cache.get(cacheKey, Float[].class);
                if (cached != null) {
                    float[] result = new float[cached.length];
                    for (int j = 0; j <cached.length; j++) {
                        result[j] = cached[j];
                    }
                    embeddings.add(result);
                    continue;
                }
            }
            missedIndices.add(i);
            embeddings.add(null); // placeholder
        }

        // Generate embeddings for missed texts
        if (!missedIndices.isEmpty()) {
            List<float[]> generated = super.embedBatch(
                missedIndices.stream()
                    .map(texts::get)
                    .toList()
            );

            int genIdx = 0;
            Cache cache = cacheManager.getCache(CACHE_NAME);
            for (int idx : missedIndices) {
                float[] embedding = generated.get(genIdx++);
                embeddings.set(idx, embedding);

                // Cache the generated embedding
                if (cache != null) {
                    String text = texts.get(idx);
                    String cacheKey = "embedding:" + text.hashCode();
                    Float[] toCache = new Float[embedding.length];
                    for (int j = 0; j < embedding.length; j++) {
                        toCache[j] = embedding[j];
                    }
                    cache.put(cacheKey, toCache);
                }
            }
        }

        return embeddings;
    }
}
