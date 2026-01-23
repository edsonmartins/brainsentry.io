package com.integraltech.brainsentry.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

import java.util.HashMap;
import java.util.Map;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Unit tests for StatsController.
 * Uses standalone MockMvc setup with mocked repository.
 */
@DisplayName("StatsController Unit Tests")
class StatsControllerTest {

    private MockMvc mockMvc;

    private final ObjectMapper objectMapper = new ObjectMapper();

    private final MemoryJpaRepository memoryJpaRepository = Mockito.mock(MemoryJpaRepository.class);

    private final StatsController statsController = new StatsController(memoryJpaRepository);

    private final String tenantId = "test-tenant";

    @BeforeEach
    void setUp() {
        mockMvc = MockMvcBuilders.standaloneSetup(statsController).build();
        com.integraltech.brainsentry.config.TenantContext.setTenantId(tenantId);
    }

    @AfterEach
    void tearDown() {
        com.integraltech.brainsentry.config.TenantContext.clear();
    }

    @Nested
    @DisplayName("GET /v1/stats/overview")
    class OverviewTests {

        @Test
        @DisplayName("Should return stats overview")
        void shouldReturnStatsOverview() throws Exception {
            // Mock repository responses
            Mockito.when(memoryJpaRepository.count()).thenReturn(150L);
            Mockito.when(memoryJpaRepository.countByCategory(eq(MemoryCategory.DECISION))).thenReturn(45L);
            Mockito.when(memoryJpaRepository.countByCategory(eq(MemoryCategory.PATTERN))).thenReturn(60L);
            Mockito.when(memoryJpaRepository.countByImportance(eq(ImportanceLevel.CRITICAL))).thenReturn(20L);
            Mockito.when(memoryJpaRepository.countByImportance(eq(ImportanceLevel.IMPORTANT))).thenReturn(65L);
            Mockito.when(memoryJpaRepository.countByImportance(eq(ImportanceLevel.MINOR))).thenReturn(65L);

            mockMvc.perform(get("/v1/stats/overview"))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.totalMemories").value(150))
                    .andExpect(jsonPath("$.memoriesByCategory.DECISION").value(45))
                    .andExpect(jsonPath("$.memoriesByCategory.PATTERN").value(60))
                    .andExpect(jsonPath("$.memoriesByImportance.CRITICAL").value(20))
                    .andExpect(jsonPath("$.memoriesByImportance.IMPORTANT").value(65))
                    .andExpect(jsonPath("$.memoriesByImportance.MINOR").value(65))
                    .andExpect(jsonPath("$.requestsToday").value(0))
                    .andExpect(jsonPath("$.injectionRate").value(0.0))
                    .andExpect(jsonPath("$.totalInjections").value(0));
        }

        @Test
        @DisplayName("Should return zero stats when no memories exist")
        void shouldReturnZeroStatsWhenNoMemories() throws Exception {
            Mockito.when(memoryJpaRepository.count()).thenReturn(0L);
            Mockito.when(memoryJpaRepository.countByCategory(any())).thenReturn(0L);
            Mockito.when(memoryJpaRepository.countByImportance(any())).thenReturn(0L);

            mockMvc.perform(get("/v1/stats/overview"))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.totalMemories").value(0))
                    .andExpect(jsonPath("$.memoriesByCategory.DECISION").value(0))
                    .andExpect(jsonPath("$.memoriesByCategory.PATTERN").value(0))
                    .andExpect(jsonPath("$.memoriesByImportance.CRITICAL").value(0));
        }

        @Test
        @DisplayName("Should return stats with large numbers")
        void shouldReturnStatsWithLargeNumbers() throws Exception {
            Mockito.when(memoryJpaRepository.count()).thenReturn(10000L);
            Mockito.when(memoryJpaRepository.countByCategory(eq(MemoryCategory.DECISION))).thenReturn(3500L);
            Mockito.when(memoryJpaRepository.countByCategory(eq(MemoryCategory.PATTERN))).thenReturn(4500L);
            Mockito.when(memoryJpaRepository.countByImportance(eq(ImportanceLevel.CRITICAL))).thenReturn(500L);
            Mockito.when(memoryJpaRepository.countByImportance(eq(ImportanceLevel.IMPORTANT))).thenReturn(6000L);
            Mockito.when(memoryJpaRepository.countByImportance(eq(ImportanceLevel.MINOR))).thenReturn(3500L);

            mockMvc.perform(get("/v1/stats/overview"))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.totalMemories").value(10000))
                    .andExpect(jsonPath("$.memoriesByCategory.DECISION").value(3500))
                    .andExpect(jsonPath("$.memoriesByImportance.IMPORTANT").value(6000));
        }
    }

    @Nested
    @DisplayName("GET /v1/stats/health")
    class HealthTests {

        @Test
        @DisplayName("Should return health status UP")
        void shouldReturnHealthStatusUp() throws Exception {
            mockMvc.perform(get("/v1/stats/health"))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.status").value("UP"))
                    .andExpect(jsonPath("$.service").value("brain-sentry"))
                    .andExpect(jsonPath("$.tenant").value(tenantId))
                    .andExpect(jsonPath("$.timestamp").exists());
        }

        @Test
        @DisplayName("Should return health with all required fields")
        void shouldReturnHealthWithAllRequiredFields() throws Exception {
            mockMvc.perform(get("/v1/stats/health"))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.status").isNotEmpty())
                    .andExpect(jsonPath("$.timestamp").isNumber())
                    .andExpect(jsonPath("$.service").isNotEmpty())
                    .andExpect(jsonPath("$.tenant").isNotEmpty());
        }

        @Test
        @DisplayName("Should return valid timestamp")
        void shouldReturnValidTimestamp() throws Exception {
            long beforeRequest = System.currentTimeMillis();

            String response = mockMvc.perform(get("/v1/stats/health"))
                    .andExpect(status().isOk())
                    .andReturn()
                    .getResponse()
                    .getContentAsString();

            long afterRequest = System.currentTimeMillis();

            // Parse timestamp from response and verify it's within valid range
            Map<String, Object> healthMap = objectMapper.readValue(response, HashMap.class);
            Long timestamp = ((Number) healthMap.get("timestamp")).longValue();

            org.junit.jupiter.api.Assertions.assertTrue(
                timestamp >= beforeRequest && timestamp <= afterRequest + 1000,
                "Timestamp should be within request timeframe"
            );
        }
    }

    @Nested
    @DisplayName("Content Type Tests")
    class ContentTypeTests {

        @Test
        @DisplayName("Should return JSON content type")
        void shouldReturnJsonContentType() throws Exception {
            mockMvc.perform(get("/v1/stats/health"))
                    .andExpect(status().isOk())
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON));
        }

        @Test
        @DisplayName("Should return JSON content type for overview")
        void shouldReturnJsonContentTypeForOverview() throws Exception {
            Mockito.when(memoryJpaRepository.count()).thenReturn(0L);
            Mockito.when(memoryJpaRepository.countByCategory(any())).thenReturn(0L);
            Mockito.when(memoryJpaRepository.countByImportance(any())).thenReturn(0L);

            mockMvc.perform(get("/v1/stats/overview"))
                    .andExpect(status().isOk())
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON));
        }
    }
}
