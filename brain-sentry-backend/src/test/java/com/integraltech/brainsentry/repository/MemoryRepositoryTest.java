package com.integraltech.brainsentry.repository;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.domain.enums.ValidationStatus;
import com.integraltech.brainsentry.repository.impl.MemoryRepositoryImpl;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;

import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@DisplayName("MemoryRepository Unit Tests")
class MemoryRepositoryTest {

    @Mock
    private JedisPool jedisPool;

    @Mock
    private Jedis jedis;

    private MemoryRepository repository;
    private ObjectMapper objectMapper;

    private final String tenantId = "test-tenant";
    private final String memoryId = "mem_test123";

    @BeforeEach
    void setUp() {
        objectMapper = JsonMapper.builder()
                .addModule(new JavaTimeModule())
                .build();
        // Updated constructor signature for MemoryRepositoryImpl
        repository = new MemoryRepositoryImpl(jedisPool, objectMapper, "test_brainsentry",
                "localhost", 6379, "");

        // Setup default JedisPool behavior
        lenient().when(jedisPool.getResource()).thenReturn(jedis);
    }

    @AfterEach
    void tearDown() {
        // Clean up - no strict verification as different tests have different interactions
    }

    private Memory createTestMemory(String id) {
        return Memory.builder()
                .id(id)
                .tenantId(tenantId)
                .content("Test content")
                .summary("Test summary")
                .category(MemoryCategory.PATTERN)
                .importance(ImportanceLevel.IMPORTANT)
                .tags(List.of("java", "spring", "redis"))
                .validationStatus(ValidationStatus.APPROVED)
                .version(1)
                .accessCount(5)
                .injectionCount(3)
                .helpfulCount(2)
                .notHelpfulCount(0)
                .createdAt(Instant.now())
                .updatedAt(Instant.now())
                .lastAccessedAt(Instant.now())
                .build();
    }

    private Memory createTestMemory() {
        return createTestMemory(memoryId);
    }

    private String toJson(Memory memory) {
        try {
            return objectMapper.writeValueAsString(memory);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Failed to serialize memory to JSON", e);
        }
    }

    @Nested
    @DisplayName("save()")
    class SaveTests {

        @Test
        @DisplayName("Should save memory with generated ID when ID is null")
        void shouldSaveWithGeneratedId() {
            Memory memory = Memory.builder()
                    .id(null)
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("tag1"))
                    .build();

            when(jedis.set(anyString(), anyString())).thenReturn("OK");
            when(jedis.sadd(anyString(), anyString())).thenReturn(1L);
            when(jedis.sadd(anyString(), anyString())).thenReturn(1L);

            Memory saved = repository.save(memory);

            assertThat(saved).isNotNull();
            assertThat(saved.getId()).isNotNull();
            assertThat(saved.getId()).startsWith("mem_");
            assertThat(saved.getCreatedAt()).isNotNull();
            assertThat(saved.getUpdatedAt()).isNotNull();

            verify(jedis).set(startsWith("memory:"), anyString());
            verify(jedis).sadd("tenant_memories:" + tenantId, saved.getId());
        }

        @Test
        @DisplayName("Should save memory with existing ID")
        void shouldSaveWithExistingId() {
            Memory memory = createTestMemory();

            when(jedis.set(anyString(), anyString())).thenReturn("OK");
            when(jedis.sadd(anyString(), anyString())).thenReturn(1L);

            Memory saved = repository.save(memory);

            assertThat(saved).isEqualTo(memory);
            verify(jedis).set(eq("memory:" + memoryId), anyString());
        }

        @Test
        @DisplayName("Should index tags correctly")
        void shouldIndexTagsCorrectly() {
            Memory memory = createTestMemory();

            when(jedis.set(anyString(), anyString())).thenReturn("OK");
            when(jedis.sadd(anyString(), anyString())).thenReturn(1L);

            repository.save(memory);

            verify(jedis).sadd("tag_idx:java", memoryId);
            verify(jedis).sadd("tag_idx:spring", memoryId);
            verify(jedis).sadd("tag_idx:redis", memoryId);
        }

        @Test
        @DisplayName("Should handle empty tags list")
        void shouldHandleEmptyTags() {
            Memory memory = Memory.builder()
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.DECISION)
                    .importance(ImportanceLevel.CRITICAL)
                    .tags(List.of())
                    .build();

            when(jedis.set(anyString(), anyString())).thenReturn("OK");
            when(jedis.sadd(anyString(), anyString())).thenReturn(1L);

            Memory saved = repository.save(memory);

            assertThat(saved).isNotNull();
            verify(jedis, never()).sadd(startsWith("tag_idx:"), anyString());
        }

        @Test
        @DisplayName("Should handle null tags")
        void shouldHandleNullTags() {
            Memory memory = Memory.builder()
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.DECISION)
                    .importance(ImportanceLevel.CRITICAL)
                    .tags(null)
                    .build();

            when(jedis.set(anyString(), anyString())).thenReturn("OK");
            when(jedis.sadd(anyString(), anyString())).thenReturn(1L);

            Memory saved = repository.save(memory);

            assertThat(saved).isNotNull();
            verify(jedis, never()).sadd(startsWith("tag_idx:"), anyString());
        }

        @Test
        @DisplayName("Should set timestamps correctly")
        void shouldSetTimestampsCorrectly() {
            Instant beforeSave = Instant.now();

            Memory memory = Memory.builder()
                    .id(memoryId)
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of())
                    .createdAt(null)
                    .build();

            when(jedis.set(anyString(), anyString())).thenReturn("OK");
            when(jedis.sadd(anyString(), anyString())).thenReturn(1L);

            Memory saved = repository.save(memory);

            assertThat(saved.getCreatedAt()).isNotNull();
            assertThat(saved.getCreatedAt()).isAfterOrEqualTo(beforeSave);
            assertThat(saved.getUpdatedAt()).isNotNull();
        }

        @Test
        @DisplayName("Should throw exception when Jedis fails")
        void shouldThrowExceptionWhenJedisFails() {
            Memory memory = createTestMemory();

            when(jedis.set(anyString(), anyString()))
                    .thenThrow(new RuntimeException("Connection failed"));

            assertThatThrownBy(() -> repository.save(memory))
                    .isInstanceOf(RuntimeException.class)
                    .hasMessageContaining("Failed to save memory");
        }
    }

    @Nested
    @DisplayName("findById()")
    class FindByIdTests {

        @Test
        @DisplayName("Should return memory when found")
        void shouldReturnMemoryWhenFound() {
            Memory memory = createTestMemory();
            String json = toJson(memory);

            when(jedis.get("memory:" + memoryId)).thenReturn(json);

            Optional<Memory> result = repository.findById(memoryId);

            assertThat(result).isPresent();
            assertThat(result.get().getId()).isEqualTo(memoryId);
            assertThat(result.get().getContent()).isEqualTo("Test content");
            assertThat(result.get().getTenantId()).isEqualTo(tenantId);
        }

        @Test
        @DisplayName("Should return empty when memory not found")
        void shouldReturnEmptyWhenNotFound() {
            when(jedis.get("memory:" + memoryId)).thenReturn(null);

            Optional<Memory> result = repository.findById(memoryId);

            assertThat(result).isEmpty();
        }

        @Test
        @DisplayName("Should return empty on deserialization error")
        void shouldReturnEmptyOnError() {
            when(jedis.get("memory:" + memoryId)).thenReturn("invalid json");

            Optional<Memory> result = repository.findById(memoryId);

            assertThat(result).isEmpty();
        }
    }

    @Nested
    @DisplayName("findByTenantId()")
    class FindByTenantIdTests {

        @Test
        @DisplayName("Should return all memories for tenant")
        void shouldReturnAllMemoriesForTenant() {
            Memory memory1 = createTestMemory();
            Memory memory2 = Memory.builder()
                    .id("mem_test456")
                    .tenantId(tenantId)
                    .content("Test content 2")
                    .summary("Test summary 2")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .createdAt(Instant.now())
                    .build();

            Set<String> memoryIds = Set.of(memoryId, "mem_test456");

            when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(memoryIds);
            when(jedis.get("memory:" + memoryId)).thenReturn(toJson(memory1));
            when(jedis.get("memory:mem_test456")).thenReturn(toJson(memory2));

            List<Memory> result = repository.findByTenantId(tenantId);

            assertThat(result).hasSize(2);
        }

        @Test
        @DisplayName("Should return empty list when tenant has no memories")
        void shouldReturnEmptyListForNoMemories() {
            when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(Set.of());

            List<Memory> result = repository.findByTenantId(tenantId);

            assertThat(result).isEmpty();
        }

        @Test
        @DisplayName("Should handle partial missing memories gracefully")
        void shouldHandlePartialMissingMemories() {
            Set<String> memoryIds = Set.of(memoryId, "mem_missing");

            when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(memoryIds);
            when(jedis.get("memory:" + memoryId)).thenReturn(toJson(createTestMemory()));
            when(jedis.get("memory:mem_missing")).thenReturn(null);

            List<Memory> result = repository.findByTenantId(tenantId);

            assertThat(result).hasSize(1);
            assertThat(result.get(0).getId()).isEqualTo(memoryId);
        }
    }

    @Nested
    @DisplayName("findByCategory()")
    class FindByCategoryTests {

        @Test
        @DisplayName("Should filter memories by category")
        void shouldFilterByCategory() {
            Memory patternMemory = createTestMemory();
            Memory decisionMemory = Memory.builder()
                    .id("mem_decision")
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.DECISION)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .createdAt(Instant.now())
                    .build();

            Set<String> memoryIds = Set.of(memoryId, "mem_decision");

            when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(memoryIds);
            when(jedis.get(anyString())).thenAnswer(invocation -> {
                String id = invocation.getArgument(0).toString().replace("memory:", "");
                if (id.equals(memoryId)) return toJson(patternMemory);
                if (id.equals("mem_decision")) return toJson(decisionMemory);
                return null;
            });

            List<Memory> result = repository.findByCategory("PATTERN", tenantId);

            assertThat(result).hasSize(1);
            assertThat(result.get(0).getCategory()).isEqualTo(MemoryCategory.PATTERN);
        }

        @Test
        @DisplayName("Should return empty list when no memories match category")
        void shouldReturnEmptyWhenNoMatch() {
            Set<String> memoryIds = Set.of(memoryId);

            lenient().when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(memoryIds);
            lenient().when(jedis.get(anyString())).thenReturn(toJson(createTestMemory()));

            List<Memory> result = repository.findByCategory("BUG", tenantId);

            assertThat(result).isEmpty();
        }
    }

    @Nested
    @DisplayName("findByImportance()")
    class FindByImportanceTests {

        @Test
        @DisplayName("Should filter and sort by importance")
        void shouldFilterAndSortByImportance() {
            Memory criticalMemory = Memory.builder()
                    .id("mem_critical")
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.CRITICAL)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .accessCount(100)
                    .createdAt(Instant.now())
                    .build();

            Memory importantMemory = Memory.builder()
                    .id("mem_important")
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .accessCount(50)
                    .createdAt(Instant.now())
                    .build();

            Memory minorMemory = Memory.builder()
                    .id("mem_minor")
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.CRITICAL)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .accessCount(10)
                    .createdAt(Instant.now())
                    .build();

            Set<String> memoryIds = Set.of("mem_critical", "mem_important", "mem_minor");

            when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(memoryIds);
            when(jedis.get(anyString())).thenAnswer(invocation -> {
                String id = invocation.getArgument(0).toString().replace("memory:", "");
                if (id.equals("mem_critical")) return toJson(criticalMemory);
                if (id.equals("mem_important")) return toJson(importantMemory);
                if (id.equals("mem_minor")) return toJson(minorMemory);
                return null;
            });

            List<Memory> result = repository.findByImportance("CRITICAL", tenantId);

            assertThat(result).hasSize(2);
            assertThat(result.get(0).getId()).isEqualTo("mem_critical"); // Higher accessCount
            assertThat(result.get(1).getId()).isEqualTo("mem_minor");
        }
    }

    @Nested
    @DisplayName("findByTags()")
    class FindByTagsTests {

        @Test
        @DisplayName("Should find memories with all specified tags")
        void shouldFindWithAllTags() {
            Memory memory1 = createTestMemory();
            Memory memory2 = Memory.builder()
                    .id("mem_test2")
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("java", "spring"))
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .createdAt(Instant.now())
                    .build();

            Set<String> javaAndSpring = Set.of(memoryId, "mem_test2");

            when(jedis.smembers("tag_idx:java")).thenReturn(javaAndSpring);
            when(jedis.smembers("tag_idx:spring")).thenReturn(javaAndSpring);
            when(jedis.get(anyString())).thenAnswer(invocation -> {
                String id = invocation.getArgument(0).toString().replace("memory:", "");
                if (id.equals(memoryId)) return toJson(memory1);
                if (id.equals("mem_test2")) return toJson(memory2);
                return null;
            });

            List<Memory> result = repository.findByTags(List.of("java", "spring"), tenantId);

            assertThat(result).hasSize(2);
        }

        @Test
        @DisplayName("Should return empty when no memories have all tags")
        void shouldReturnEmptyWhenNoMatch() {
            Set<String> javaMemories = Set.of(memoryId);
            Set<String> springMemories = Set.of("mem_other");

            when(jedis.smembers("tag_idx:java")).thenReturn(javaMemories);
            when(jedis.smembers("tag_idx:spring")).thenReturn(springMemories);

            List<Memory> result = repository.findByTags(List.of("java", "spring"), tenantId);

            assertThat(result).isEmpty();
        }

        @Test
        @DisplayName("Should filter by tenant")
        void shouldFilterByTenant() {
            Memory memory1 = createTestMemory();
            Memory memory2 = Memory.builder()
                    .id("mem_other_tenant")
                    .tenantId("other-tenant")
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("java"))
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .createdAt(Instant.now())
                    .build();

            Set<String> javaMemories = Set.of(memoryId, "mem_other_tenant");

            when(jedis.smembers("tag_idx:java")).thenReturn(javaMemories);
            when(jedis.get(anyString())).thenAnswer(invocation -> {
                String id = invocation.getArgument(0).toString().replace("memory:", "");
                if (id.equals(memoryId)) return toJson(memory1);
                if (id.equals("mem_other_tenant")) return toJson(memory2);
                return null;
            });

            List<Memory> result = repository.findByTags(List.of("java"), tenantId);

            assertThat(result).hasSize(1);
            assertThat(result.get(0).getTenantId()).isEqualTo(tenantId);
        }
    }

    @Nested
    @DisplayName("vectorSearch()")
    class VectorSearchTests {

        @Test
        @DisplayName("Should return memories sorted by access count (placeholder)")
        void shouldReturnSortedByAccessCount() {
            Memory memory1 = Memory.builder()
                    .id("mem_1")
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .accessCount(100)
                    .createdAt(Instant.now())
                    .build();

            Memory memory2 = Memory.builder()
                    .id("mem_2")
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .accessCount(50)
                    .createdAt(Instant.now())
                    .build();

            Set<String> memoryIds = Set.of("mem_1", "mem_2");

            when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(memoryIds);
            when(jedis.get(anyString())).thenAnswer(invocation -> {
                String id = invocation.getArgument(0).toString().replace("memory:", "");
                if (id.equals("mem_1")) return toJson(memory1);
                if (id.equals("mem_2")) return toJson(memory2);
                return null;
            });

            float[] embedding = new float[]{0.1f, 0.2f, 0.3f};
            List<Memory> result = repository.vectorSearch(embedding, 10, tenantId);

            assertThat(result).hasSize(2);
            assertThat(result.get(0).getId()).isEqualTo("mem_1"); // Higher accessCount
        }

        @Test
        @DisplayName("Should limit results")
        void shouldLimitResults() {
            Set<String> memoryIds = Set.of("mem_1", "mem_2", "mem_3");

            when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(memoryIds);
            when(jedis.get(anyString())).thenAnswer(invocation -> {
                String id = invocation.getArgument(0).toString().replace("memory:", "");
                return toJson(createTestMemory(id));
            });

            float[] embedding = new float[]{0.1f, 0.2f, 0.3f};
            List<Memory> result = repository.vectorSearch(embedding, 2, tenantId);

            assertThat(result).hasSize(2);
        }
    }

    @Nested
    @DisplayName("findRelated()")
    class FindRelatedTests {

        @Test
        @DisplayName("Should return memories with same tags")
        @org.junit.jupiter.api.Disabled("Requires FalkorDB mock - will be covered by integration tests")
        void shouldReturnMemoriesWithSameTags() {
            Memory sourceMemory = createTestMemory();
            Memory relatedMemory = Memory.builder()
                    .id("mem_related")
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("java", "spring", "redis"))
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .createdAt(Instant.now())
                    .build();

            Set<String> tagMembers = Set.of(memoryId, "mem_related");

            // Use lenient for optional stubbing since fallback paths may vary
            lenient().when(jedis.get("memory:" + memoryId)).thenReturn(toJson(sourceMemory));
            lenient().when(jedis.get("memory:mem_related")).thenReturn(toJson(relatedMemory));
            // Mock all three tags to return the same set of memory IDs
            lenient().when(jedis.smembers("tag_idx:java")).thenReturn(tagMembers);
            lenient().when(jedis.smembers("tag_idx:spring")).thenReturn(tagMembers);
            lenient().when(jedis.smembers("tag_idx:redis")).thenReturn(tagMembers);

            // The fallback also needs findById to work through the getResource() already set up
            lenient().when(jedis.get(anyString())).thenAnswer(invocation -> {
                String key = invocation.getArgument(0);
                if (key.equals("memory:" + memoryId)) return toJson(sourceMemory);
                if (key.equals("memory:mem_related")) return toJson(relatedMemory);
                return null;
            });

            List<Memory> result = repository.findRelated(memoryId, 2, tenantId);

            assertThat(result).isNotEmpty();
            assertThat(result).noneMatch(m -> m.getId().equals(memoryId));
        }

        @Test
        @DisplayName("Should return empty or fallback when memory has no tags")
        void shouldReturnEmptyWhenNoTags() {
            Memory memory = Memory.builder()
                    .id(memoryId)
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .createdAt(Instant.now())
                    .build();

            lenient().when(jedis.get("memory:" + memoryId)).thenReturn(toJson(memory));
            lenient().when(jedis.smembers("tenant_memories:" + tenantId)).thenReturn(Set.of(memoryId));

            List<Memory> result = repository.findRelated(memoryId, 2, tenantId);

            // Fallback now returns most accessed memories, filtered to not include original
            // Since only the original memory exists, result should be empty
            assertThat(result).isEmpty();
        }

        @Test
        @DisplayName("Should return empty when memory not found")
        void shouldReturnEmptyWhenNotFound() {
            lenient().when(jedis.get("memory:" + memoryId)).thenReturn(null);

            List<Memory> result = repository.findRelated(memoryId, 2, tenantId);

            assertThat(result).isEmpty();
        }
    }

    @Nested
    @DisplayName("deleteById()")
    class DeleteByIdTests {

        @Test
        @DisplayName("Should delete memory and cleanup indexes")
        void shouldDeleteAndCleanup() {
            Memory memory = createTestMemory();

            when(jedis.get("memory:" + memoryId)).thenReturn(toJson(memory));
            when(jedis.del("memory:" + memoryId)).thenReturn(1L);
            when(jedis.srem(anyString(), anyString())).thenReturn(1L);

            boolean result = repository.deleteById(memoryId);

            assertThat(result).isTrue();

            verify(jedis).del("memory:" + memoryId);
            verify(jedis).srem("tenant_memories:" + tenantId, memoryId);
            verify(jedis).srem("tag_idx:java", memoryId);
            verify(jedis).srem("tag_idx:spring", memoryId);
            verify(jedis).srem("tag_idx:redis", memoryId);
        }

        @Test
        @DisplayName("Should return false when memory not found")
        void shouldReturnFalseWhenNotFound() {
            when(jedis.get("memory:" + memoryId)).thenReturn(null);

            boolean result = repository.deleteById(memoryId);

            assertThat(result).isFalse();

            verify(jedis, never()).del(anyString());
            verify(jedis, never()).srem(anyString(), anyString());
        }

        @Test
        @DisplayName("Should handle null tenantId")
        void shouldHandleNullTenantId() {
            Memory memory = Memory.builder()
                    .id(memoryId)
                    .tenantId(null)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of("java", "spring", "redis"))
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(1)
                    .createdAt(Instant.now())
                    .build();

            when(jedis.get("memory:" + memoryId)).thenReturn(toJson(memory));
            when(jedis.del("memory:" + memoryId)).thenReturn(1L);

            repository.deleteById(memoryId);

            verify(jedis, never()).srem(startsWith("tenant_memories:"), anyString());
        }
    }

    @Nested
    @DisplayName("countByTenantId()")
    class CountByTenantIdTests {

        @Test
        @DisplayName("Should return count of memories")
        void shouldReturnCount() {
            when(jedis.scard("tenant_memories:" + tenantId)).thenReturn(42L);

            long count = repository.countByTenantId(tenantId);

            assertThat(count).isEqualTo(42L);
        }

        @Test
        @DisplayName("Should return 0 when tenant has no memories")
        void shouldReturnZeroForNoMemories() {
            when(jedis.scard("tenant_memories:" + tenantId)).thenReturn(0L);

            long count = repository.countByTenantId(tenantId);

            assertThat(count).isEqualTo(0L);
        }
    }

    @Nested
    @DisplayName("archiveVersion()")
    class ArchiveVersionTests {

        @Test
        @DisplayName("Should log archive operation")
        void shouldLogArchive() {
            Memory memory = createTestMemory();

            repository.archiveVersion(memory);

            // Placeholder - just verifies no exception is thrown
            assertThat(memory.getId()).isEqualTo(memoryId);
        }
    }

    @Nested
    @DisplayName("getVersionHistory()")
    class GetVersionHistoryTests {

        @Test
        @DisplayName("Should return version list")
        void shouldReturnVersionList() {
            Memory memory = Memory.builder()
                    .id(memoryId)
                    .tenantId(tenantId)
                    .content("Test content")
                    .summary("Test summary")
                    .category(MemoryCategory.PATTERN)
                    .importance(ImportanceLevel.IMPORTANT)
                    .tags(List.of())
                    .validationStatus(ValidationStatus.APPROVED)
                    .version(3)
                    .createdAt(Instant.now())
                    .build();

            when(jedis.get("memory:" + memoryId)).thenReturn(toJson(memory));

            List<Integer> versions = repository.getVersionHistory(memoryId, tenantId);

            assertThat(versions).containsExactly(3);
        }

        @Test
        @DisplayName("Should return empty list when memory not found")
        void shouldReturnEmptyWhenNotFound() {
            when(jedis.get("memory:" + memoryId)).thenReturn(null);

            List<Integer> versions = repository.getVersionHistory(memoryId, tenantId);

            assertThat(versions).isEmpty();
        }
    }
}
