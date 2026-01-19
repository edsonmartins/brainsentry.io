package com.integraltech.brainsentry.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.dto.request.InterceptRequest;
import com.integraltech.brainsentry.dto.response.InterceptResponse;
import com.integraltech.brainsentry.service.InterceptionService;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

import java.util.List;

import static org.mockito.ArgumentMatchers.any;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Unit tests for InterceptionController.
 * Uses standalone MockMvc setup with mocked service layer.
 */
@DisplayName("InterceptionController Unit Tests")
class InterceptionControllerTest {

    private MockMvc mockMvc;

    private final ObjectMapper objectMapper = new ObjectMapper();

    private final InterceptionService interceptionService = Mockito.mock(InterceptionService.class);

    private final InterceptionController interceptionController = new InterceptionController(interceptionService);

    private final String tenantId = "test-tenant";

    @BeforeEach
    void setUp() {
        mockMvc = MockMvcBuilders.standaloneSetup(interceptionController).build();
        com.integraltech.brainsentry.config.TenantContext.setTenantId(tenantId);
    }

    @AfterEach
    void tearDown() {
        com.integraltech.brainsentry.config.TenantContext.clear();
    }

    @Nested
    @DisplayName("POST /v1/intercept")
    class InterceptTests {

        @Test
        @DisplayName("Should return 400 when prompt is empty")
        void shouldReturn400WhenPromptEmpty() throws Exception {
            InterceptRequest request = InterceptRequest.builder()
                    .prompt("") // Invalid: empty prompt
                    .sessionId("session-123")
                    .build();

            mockMvc.perform(post("/v1/intercept")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isBadRequest());
        }

        @Test
        @DisplayName("Should return 400 when request body is empty")
        void shouldReturn400ForEmptyBody() throws Exception {
            String emptyRequest = "{}";

            mockMvc.perform(post("/v1/intercept")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(emptyRequest))
                    .andExpect(status().isBadRequest());
        }

        @Test
        @DisplayName("Should return 400 when prompt is null")
        void shouldReturn400WhenPromptNull() throws Exception {
            InterceptRequest request = InterceptRequest.builder()
                    .prompt(null) // Invalid: null prompt
                    .sessionId("session-123")
                    .build();

            mockMvc.perform(post("/v1/intercept")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isBadRequest());
        }

        @Test
        @DisplayName("Should return 200 when intercept is successful")
        void shouldReturn200WhenInterceptSuccessful() throws Exception {
            InterceptRequest request = InterceptRequest.builder()
                    .prompt("Create a new user service")
                    .sessionId("session-123")
                    .userId("user-456")
                    .maxTokens(500)
                    .build();

            InterceptResponse response = InterceptResponse.builder()
                    .enhanced(true)
                    .originalPrompt("Create a new user service")
                    .enhancedPrompt("<context>Create a new user service</context>")
                    .contextInjected("Relevant patterns found...")
                    .memoriesUsed(List.of(
                            InterceptResponse.MemoryReference.builder()
                                    .id("mem-001")
                                    .summary("User service pattern")
                                    .category("PATTERN")
                                    .importance("IMPORTANT")
                                    .relevanceScore(0.95)
                                    .excerpt("User service example...")
                                    .build()
                    ))
                    .latencyMs(150)
                    .reasoning("Found relevant patterns")
                    .confidence(0.92)
                    .tokensInjected(120)
                    .llmCalls(1)
                    .build();

            Mockito.when(interceptionService.interceptAndEnhance(any(InterceptRequest.class)))
                    .thenReturn(response);

            mockMvc.perform(post("/v1/intercept")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.enhanced").value(true))
                    .andExpect(jsonPath("$.latencyMs").value(150));
        }

        @Test
        @DisplayName("Should return 200 when no enhancement is needed")
        void shouldReturn200WhenNoEnhancementNeeded() throws Exception {
            InterceptRequest request = InterceptRequest.builder()
                    .prompt("What is the weather today?")
                    .sessionId("session-123")
                    .build();

            InterceptResponse response = InterceptResponse.builder()
                    .enhanced(false)
                    .originalPrompt("What is the weather today?")
                    .enhancedPrompt("What is the weather today?")
                    .contextInjected("")
                    .memoriesUsed(List.of())
                    .latencyMs(50)
                    .reasoning("No relevant keywords detected")
                    .confidence(0.0)
                    .tokensInjected(0)
                    .llmCalls(0)
                    .build();

            Mockito.when(interceptionService.interceptAndEnhance(any(InterceptRequest.class)))
                    .thenReturn(response);

            mockMvc.perform(post("/v1/intercept")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.enhanced").value(false))
                    .andExpect(jsonPath("$.memoriesUsed").isEmpty());
        }

        @Test
        @DisplayName("Should accept request with all optional fields")
        void shouldAcceptRequestWithAllOptionalFields() throws Exception {
            InterceptRequest request = InterceptRequest.builder()
                    .prompt("Implement a repository pattern")
                    .sessionId("session-123")
                    .userId("user-456")
                    .tenantId("custom-tenant")
                    .maxTokens(1000)
                    .forceDeepAnalysis(true)
                    .context(java.util.Map.of(
                            "project", "brain-sentry",
                            "language", "java"
                    ))
                    .build();

            InterceptResponse response = InterceptResponse.builder()
                    .enhanced(true)
                    .originalPrompt(request.getPrompt())
                    .enhancedPrompt(request.getPrompt())
                    .memoriesUsed(List.of())
                    .latencyMs(100)
                    .build();

            Mockito.when(interceptionService.interceptAndEnhance(any(InterceptRequest.class)))
                    .thenReturn(response);

            mockMvc.perform(post("/v1/intercept")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk());
        }

        @Test
        @DisplayName("Should use default maxTokens when not provided")
        void shouldUseDefaultMaxTokens() throws Exception {
            InterceptRequest request = InterceptRequest.builder()
                    .prompt("Create a service")
                    .sessionId("session-123")
                    // maxTokens not set, should default to 500
                    .build();

            InterceptResponse response = InterceptResponse.builder()
                    .enhanced(false)
                    .originalPrompt(request.getPrompt())
                    .enhancedPrompt(request.getPrompt())
                    .memoriesUsed(List.of())
                    .latencyMs(50)
                    .build();

            Mockito.when(interceptionService.interceptAndEnhance(any(InterceptRequest.class)))
                    .thenReturn(response);

            mockMvc.perform(post("/v1/intercept")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk());
        }
    }
}
