package com.integraltech.brainsentry.performance;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.dto.request.SearchRequest;
import com.integraltech.brainsentry.dto.response.MemoryListResponse;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.service.EmbeddingService;
import com.integraltech.brainsentry.service.MemoryService;
import org.junit.jupiter.api.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.security.test.context.support.WithMockUser;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Performance tests for the Brain Sentry API.
 *
 * Tests response times, throughput, and resource usage
 * under various load conditions.
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@Testcontainers(disabledWithoutDocker = true)
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
class PerformanceBenchmarkTest {

    @Autowired
    private MemoryService memoryService;

    @Autowired
    private MemoryJpaRepository memoryRepository;

    @Autowired
    private EmbeddingService embeddingService;

    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:16-alpine")
            .withDatabaseName("brain_perf_test")
            .withUsername("test")
            .withPassword("test");

    @DynamicPropertySource
    static void postgresProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
        registry.add("spring.jpa.hibernate.ddl-auto", () -> "create-drop");
    }

    private static final String PERF_TENANT = "perf-test-tenant";

    @BeforeEach
    void setUp() {
        TenantContext.setTenantId(PERF_TENANT);
    }

    @AfterEach
    void tearDown() {
        TenantContext.clear();
    }

    @AfterAll
    static void cleanup(@Autowired MemoryJpaRepository repository) {
        repository.deleteAll();
    }

    // ==================== Memory Creation Performance ====================

    @Test
    @Order(1)
    @DisplayName("Performance: Create 100 memories in acceptable time")
    void testBulkMemoryCreation() {
        int count = 100;
        long startTime = System.currentTimeMillis();

        for (int i = 0; i < count; i++) {
            CreateMemoryRequest request = CreateMemoryRequest.builder()
                    .content("Performance test memory #" + i + " with some content to make it realistic")
                    .summary("Perf test #" + i)
                    .tenantId(PERF_TENANT)
                    .createdBy("perf-test")
                    .build();

            memoryService.createMemory(request);
        }

        long duration = System.currentTimeMillis() - startTime;
        double avgTime = (double) duration / count;

        System.out.println(String.format("Created %d memories in %d ms (avg: %.2f ms per memory)",
                count, duration, avgTime));

        // Assertions for performance thresholds
        assertThat(avgTime).isLessThan(100);  // Average should be under 100ms per memory
        assertThat(duration).isLessThan(10000);  // Total should be under 10 seconds
    }

    @Test
    @Order(2)
    @DisplayName("Performance: Single memory creation should be fast")
    void testSingleMemoryCreationSpeed() {
        CreateMemoryRequest request = CreateMemoryRequest.builder()
                .content("Quick performance test memory")
                .summary("Speed test")
                .tenantId(PERF_TENANT)
                .createdBy("perf-test")
                .build();

        long startTime = System.nanoTime();
        var response = memoryService.createMemory(request);
        long endTime = System.nanoTime();

        double durationMs = (endTime - startTime) / 1_000_000.0;

        System.out.println(String.format("Single memory creation took: %.2f ms", durationMs));

        assertThat(durationMs).isLessThan(200);  // Should complete in under 200ms
        assertThat(response.getId()).isNotNull();
    }

    // ==================== Search Performance ====================

    @Test
    @Order(3)
    @DisplayName("Performance: Search through memories")
    void testSearchPerformance() {
        // Setup: Create some memories
        for (int i = 0; i < 20; i++) {
            CreateMemoryRequest request = CreateMemoryRequest.builder()
                    .content("Performance test content #" + i + " about programming and software development")
                    .summary("Perf #" + i)
                    .tenantId(PERF_TENANT)
                    .createdBy("perf-test")
                    .build();
            memoryService.createMemory(request);
        }

        // Test search performance
        SearchRequest searchRequest = new SearchRequest();
        searchRequest.setQuery("programming software development");
        searchRequest.setLimit(50);

        long startTime = System.nanoTime();
        var results = memoryService.search(searchRequest);
        long endTime = System.nanoTime();

        double durationMs = (endTime - startTime) / 1_000_000.0;

        System.out.println(String.format("Search took: %.2f ms, returned %d results",
                durationMs, results.size()));

        assertThat(durationMs).isLessThan(500);  // Search should be under 500ms
    }

    // ==================== Batch Operations Performance ====================

    @Test
    @Order(4)
    @DisplayName("Performance: Bulk retrieval of memories")
    void testBulkRetrievalPerformance() {
        // Create some memories if needed
        long count = memoryRepository.count();
        if (count < 20) {
            for (int i = 0; i < 20 - count; i++) {
                CreateMemoryRequest request = CreateMemoryRequest.builder()
                        .content("Bulk retrieval test memory #" + i)
                        .tenantId(PERF_TENANT)
                        .build();
                memoryService.createMemory(request);
            }
        }

        long startTime = System.nanoTime();
        var memories = memoryService.listMemories(0, 20);
        long endTime = System.nanoTime();

        double durationMs = (endTime - startTime) / 1_000_000.0;

        System.out.println(String.format("Retrieved %d memories in %.2f ms",
                memories.getMemories().size(), durationMs));

        assertThat(memories.getMemories()).isNotEmpty();
        assertThat(durationMs).isLessThan(200);  // Should be under 200ms
    }

    // ==================== Embedding Generation Performance ====================

    @Test
    @Order(5)
    @DisplayName("Performance: Embedding generation speed")
    void testEmbeddingGenerationSpeed() {
        String testContent = "This is a test content for embedding generation performance testing " +
                "with enough text to be representative of actual use cases in the system.";

        long startTime = System.nanoTime();
        float[] embedding = embeddingService.embed(testContent);
        long endTime = System.nanoTime();

        double durationMs = (endTime - startTime) / 1_000_000.0;

        System.out.println(String.format("Embedding generation took: %.2f ms (dimension: %d)",
                durationMs, embedding.length));

        assertThat(embedding).isNotNull();
        assertThat(embedding.length).isGreaterThan(0);
        assertThat(durationMs).isLessThan(1000);  // Should be under 1 second
    }

    // ==================== Concurrent Access Performance ====================

    @Test
    @Order(6)
    @DisplayName("Performance: Concurrent memory creation")
    void testConcurrentCreation() throws InterruptedException {
        int threadCount = 10;
        int memoriesPerThread = 10;
        Thread[] threads = new Thread[threadCount];

        long startTime = System.currentTimeMillis();

        for (int i = 0; i < threadCount; i++) {
            final int threadId = i;
            threads[i] = new Thread(() -> {
                for (int j = 0; j < memoriesPerThread; j++) {
                    try {
                        CreateMemoryRequest request = CreateMemoryRequest.builder()
                                .content("Concurrent test - thread " + threadId + " memory " + j)
                                .tenantId(PERF_TENANT)
                                .createdBy("concurrent-test")
                                .build();
                        memoryService.createMemory(request);
                    } catch (Exception e) {
                        System.err.println("Error in thread " + threadId + ": " + e.getMessage());
                    }
                }
            });
            threads[i].start();
        }

        for (Thread thread : threads) {
            thread.join();
        }

        long duration = System.currentTimeMillis() - startTime;
        int totalMemories = threadCount * memoriesPerThread;

        System.out.println(String.format("Created %d memories concurrently in %d ms (%.2f memories/sec)",
                totalMemories, duration, (totalMemories * 1000.0) / duration));

        assertThat(duration).isLessThan(30000);  // Should complete in under 30 seconds
    }

    // ==================== Memory Usage ====================

    @Test
    @Order(7)
    @DisplayName("Performance: Memory usage during bulk operations")
    void testMemoryUsage() {
        Runtime runtime = Runtime.getRuntime();
        runtime.gc();

        long memoryBefore = runtime.totalMemory() - runtime.freeMemory();

        // Create 100 memories
        for (int i = 0; i < 100; i++) {
            CreateMemoryRequest request = CreateMemoryRequest.builder()
                    .content("Memory usage test content #" + i + " with some data to consume memory")
                    .tenantId(PERF_TENANT)
                    .build();
            memoryService.createMemory(request);
        }

        runtime.gc();
        long memoryAfter = runtime.totalMemory() - runtime.freeMemory();

        long memoryUsed = memoryAfter - memoryBefore;
        double memoryUsedMb = memoryUsed / (1024.0 * 1024.0);

        System.out.println(String.format("Memory used for 100 memories: %.2f MB", memoryUsedMb));

        // Memory usage should be reasonable
        assertThat(memoryUsedMb).isLessThan(50);
    }

    // ==================== Pagination Performance ====================

    @Test
    @Order(8)
    @DisplayName("Performance: Pagination performance")
    void testPaginationPerformance() {
        // Test different page sizes
        int[] pageSizes = {10, 25, 50};

        for (int pageSize : pageSizes) {
            long startTime = System.nanoTime();
            var page = memoryService.listMemories(0, pageSize);
            long endTime = System.nanoTime();

            double durationMs = (endTime - startTime) / 1_000_000.0;

            System.out.println(String.format("Pagination (page size %d) took: %.2f ms, returned %d items",
                    pageSize, durationMs, page.getMemories().size()));

            assertThat(durationMs).isLessThan(100);  // Each page should load quickly
        }
    }

    // ==================== Stress Test ====================

    @Test
    @Order(9)
    @DisplayName("Performance: Stress test with rapid sequential operations")
    void testStressTest() {
        int operations = 100;
        long startTime = System.currentTimeMillis();

        for (int i = 0; i < operations; i++) {
            CreateMemoryRequest request = CreateMemoryRequest.builder()
                    .content("Stress test operation #" + i)
                    .tenantId(PERF_TENANT)
                    .build();

            memoryService.createMemory(request);

            // Do a search every 10 operations
            if (i % 10 == 0) {
                SearchRequest searchRequest = new SearchRequest();
                searchRequest.setQuery("stress test");
                searchRequest.setLimit(10);
                memoryService.search(searchRequest);
            }
        }

        long duration = System.currentTimeMillis() - startTime;

        System.out.println(String.format("Stress test: %d operations completed in %d ms (%.2f ops/sec)",
                operations, duration, (operations * 1000.0) / duration));

        // System should handle this load
        assertThat(duration).isLessThan(60000);  // Under 1 minute
    }
}
