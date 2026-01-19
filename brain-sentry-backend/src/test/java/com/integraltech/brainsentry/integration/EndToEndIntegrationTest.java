package com.integraltech.brainsentry.integration;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.dto.request.SearchRequest;
import com.integraltech.brainsentry.dto.response.MemoryResponse;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.repository.MemoryRelationshipJpaRepository;
import io.restassured.RestAssured;
import io.restassured.http.ContentType;
import org.junit.jupiter.api.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.server.LocalServerPort;
import org.springframework.security.test.context.support.WithMockUser;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.util.List;
import java.util.Map;

import static io.restassured.RestAssured.given;
import static org.assertj.core.api.Assertions.assertThat;
import static org.hamcrest.Matchers.*;

/**
 * End-to-end integration tests for the Brain Sentry API.
 *
 * Tests the complete flow from HTTP request to database persistence,
 * using Testcontainers for real PostgreSQL and RestAssured for HTTP testing.
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@Testcontainers(disabledWithoutDocker = true)
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
class EndToEndIntegrationTest {

    @LocalServerPort
    private int port;

    @Autowired
    private MemoryJpaRepository memoryRepository;

    @Autowired
    private MemoryRelationshipJpaRepository relationshipRepository;

    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:16-alpine")
            .withDatabaseName("brain_sentry_test")
            .withUsername("test")
            .withPassword("test");

    @DynamicPropertySource
    static void postgresProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
        registry.add("spring.jpa.hibernate.ddl-auto", () -> "create-drop");
    }

    private static String memoryId;
    private static final String TEST_TENANT = "e2e-test-tenant";
    private static final String TEST_USER = "e2e-user";

    @BeforeEach
    void setUp() {
        RestAssured.baseURI = "http://localhost";
        RestAssured.port = port;
        TenantContext.setTenantId(TEST_TENANT);
    }

    @AfterEach
    void tearDown() {
        TenantContext.clear();
    }

    @AfterAll
    static void cleanup(@Autowired MemoryJpaRepository repository,
                       @Autowired MemoryRelationshipJpaRepository relationshipRepo) {
        relationshipRepo.deleteAll();
        repository.deleteAll();
    }

    // ==================== Health & Startup Tests ====================

    @Test
    @Order(1)
    @DisplayName("Health check endpoint should return healthy status")
    void testHealthCheck() {
        given()
                .when()
                .get("/actuator/health")
                .then()
                .statusCode(200)
                .body("status", equalTo("UP"));
    }

    @Test
    @Order(2)
    @DisplayName("Info endpoint should return application information")
    void testInfoEndpoint() {
        given()
                .when()
                .get("/actuator/info")
                .then()
                .statusCode(200);
    }

    // ==================== Memory CRUD E2E Tests ====================

    @Test
    @Order(10)
    @WithMockUser(username = TEST_USER)
    @DisplayName("E2E: Create memory via REST API")
    void testCreateMemory() {
        CreateMemoryRequest request = CreateMemoryRequest.builder()
                .content("E2E Test: Spring Boot Actuator provides health checks and metrics")
                .summary("Test memory for E2E testing")
                .tags(List.of("e2e", "test", "spring-boot"))
                .tenantId(TEST_TENANT)
                .createdBy(TEST_USER)
                .build();

        memoryId = given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .body(request)
                .when()
                .post("/v1/memories")
                .then()
                .statusCode(201)
                .body("id", notNullValue())
                .body("content", equalTo(request.getContent()))
                .body("summary", equalTo(request.getSummary()))
                .body("tenantId", equalTo(TEST_TENANT))
                .extract()
                .path("id");

        assertThat(memoryId).isNotNull();
    }

    @Test
    @Order(11)
    @WithMockUser(username = TEST_USER)
    @DisplayName("E2E: Get memory by ID via REST API")
    void testGetMemory() {
        assertThat(memoryId).isNotNull();

        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .when()
                .get("/v1/memories/{id}", memoryId)
                .then()
                .statusCode(200)
                .body("id", equalTo(memoryId))
                .body("content", containsString("E2E Test"))
                .body("tenantId", equalTo(TEST_TENANT));
    }

    @Test
    @Order(12)
    @WithMockUser(username = TEST_USER)
    @DisplayName("E2E: List all memories via REST API")
    void testListMemories() {
        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .when()
                .get("/v1/memories")
                .then()
                .statusCode(200)
                .body("content", hasSize(greaterThan(0)))
                .body("totalElements", greaterThan(0))
                .body("content[0].id", notNullValue());
    }

    @Test
    @Order(13)
    @WithMockUser(username = TEST_USER)
    @DisplayName("E2E: Search memories via REST API")
    void testSearchMemories() {
        SearchRequest searchRequest = new SearchRequest();
        searchRequest.setQuery("Spring Boot Actuator");
        searchRequest.setLimit(10);

        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .body(searchRequest)
                .when()
                .post("/v1/memories/search")
                .then()
                .statusCode(200)
                .body("results", hasSize(greaterThan(0)));
    }

    @Test
    @Order(14)
    @WithMockUser(username = TEST_USER)
    @DisplayName("E2E: Update memory via REST API")
    void testUpdateMemory() {
        assertThat(memoryId).isNotNull();

        Map<String, Object> updates = Map.of(
                "summary", "Updated E2E test summary",
                "content", "E2E Test: Updated content with more details about Spring Boot"
        );

        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .body(updates)
                .when()
                .patch("/v1/memories/{id}", memoryId)
                .then()
                .statusCode(200)
                .body("summary", equalTo("Updated E2E test summary"));
    }

    @Test
    @Order(15)
    @WithMockUser(username = TEST_USER)
    @DisplayName("E2E: Delete memory via REST API")
    void testDeleteMemory() {
        // Create a new memory to delete
        CreateMemoryRequest request = CreateMemoryRequest.builder()
                .content("Memory to be deleted in E2E test")
                .summary("Deletion test")
                .tenantId(TEST_TENANT)
                .createdBy(TEST_USER)
                .build();

        String deleteId = given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .body(request)
                .when()
                .post("/v1/memories")
                .then()
                .statusCode(201)
                .extract()
                .path("id");

        // Delete it
        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .when()
                .delete("/v1/memories/{id}", deleteId)
                .then()
                .statusCode(204);

        // Verify it's gone
        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .when()
                .get("/v1/memories/{id}", deleteId)
                .then()
                .statusCode(404);
    }

    // ==================== Stats Endpoint E2E ====================

    @Test
    @Order(20)
    @WithMockUser(username = TEST_USER)
    @DisplayName("E2E: Get stats overview via REST API")
    void testStatsOverview() {
        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .when()
                .get("/v1/stats/overview")
                .then()
                .statusCode(200)
                .body("totalMemories", greaterThanOrEqualTo(0))
                .body("totalCategories", greaterThanOrEqualTo(0));
    }

    // ==================== MCP Endpoint E2E ====================

    @Test
    @Order(30)
    @DisplayName("E2E: Get MCP tools via REST API")
    void testGetMcpTools() {
        given()
                .contentType(ContentType.JSON)
                .when()
                .get("/v1/mcp/tools")
                .then()
                .statusCode(200)
                .body("size()", greaterThan(0));
    }

    @Test
    @Order(31)
    @DisplayName("E2E: Get MCP resources via REST API")
    void testGetMcpResources() {
        given()
                .contentType(ContentType.JSON)
                .when()
                .get("/v1/mcp/resources")
                .then()
                .statusCode(200)
                .body("size()", greaterThan(0));
    }

    @Test
    @Order(32)
    @DisplayName("E2E: MCP health check")
    void testMcpHealth() {
        given()
                .contentType(ContentType.JSON)
                .when()
                .get("/v1/mcp/health")
                .then()
                .statusCode(200)
                .body("status", equalTo("healthy"));
    }

    // ==================== Error Handling E2E ====================

    @Test
    @Order(40)
    @DisplayName("E2E: Invalid memory ID returns 404")
    void testGetNonExistentMemory() {
        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .when()
                .get("/v1/memories/non-existent-id")
                .then()
                .statusCode(404);
    }

    @Test
    @Order(41)
    @WithMockUser(username = TEST_USER)
    @DisplayName("E2E: Invalid request body returns 400")
    void testInvalidRequestBody() {
        Map<String, Object> invalidRequest = Map.of(
                "content", "",  // Invalid: empty content
                "summary", "Test"
        );

        given()
                .contentType(ContentType.JSON)
                .header("X-Tenant-ID", TEST_TENANT)
                .body(invalidRequest)
                .when()
                .post("/v1/memories")
                .then()
                .statusCode(400);
    }

    @Test
    @Order(42)
    @DisplayName("E2E: Missing tenant header should use default")
    void testMissingTenantHeader() {
        given()
                .contentType(ContentType.JSON)
                .when()
                .get("/v1/memories")
                .then()
                .statusCode(200);  // Should work with default tenant
    }
}
