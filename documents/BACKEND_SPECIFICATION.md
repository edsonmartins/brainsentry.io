# Brain Sentry - Backend Specification

**Version:** 1.0  
**Stack:** Java 17 + Spring Boot 3.2+  
**Database:** FalkorDB (Redis Graph)  
**Build Tool:** Maven  

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Domain Models](#domain-models)
3. [API Endpoints](#api-endpoints)
4. [Services](#services)
5. [Configuration](#configuration)
6. [Database Schema](#database-schema)
7. [Security](#security)

---

## Project Structure

```
brain-sentry-backend/
├── pom.xml
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── com/
│   │   │       └── integraltech/
│   │   │           └── brainsentry/
│   │   │               ├── BrainSentryApplication.java
│   │   │               │
│   │   │               ├── config/
│   │   │               │   ├── RedisConfig.java
│   │   │               │   ├── SecurityConfig.java
│   │   │               │   ├── WebConfig.java
│   │   │               │   └── LLMConfig.java
│   │   │               │
│   │   │               ├── domain/
│   │   │               │   ├── Memory.java
│   │   │               │   ├── MemoryRelationship.java
│   │   │               │   ├── AuditLog.java
│   │   │               │   ├── MemoryVersion.java
│   │   │               │   └── enums/
│   │   │               │       ├── MemoryCategory.java
│   │   │               │       ├── ImportanceLevel.java
│   │   │               │       ├── RelationshipType.java
│   │   │               │       └── ValidationStatus.java
│   │   │               │
│   │   │               ├── dto/
│   │   │               │   ├── request/
│   │   │               │   │   ├── InterceptRequest.java
│   │   │               │   │   ├── CreateMemoryRequest.java
│   │   │               │   │   ├── UpdateMemoryRequest.java
│   │   │               │   │   └── SearchRequest.java
│   │   │               │   │
│   │   │               │   └── response/
│   │   │               │       ├── InterceptResponse.java
│   │   │               │       ├── MemoryResponse.java
│   │   │               │       ├── MemoryListResponse.java
│   │   │               │       ├── StatsResponse.java
│   │   │               │       └── AuditLogResponse.java
│   │   │               │
│   │   │               ├── controller/
│   │   │               │   ├── InterceptionController.java
│   │   │               │   ├── MemoryController.java
│   │   │               │   ├── AuditController.java
│   │   │               │   └── StatsController.java
│   │   │               │
│   │   │               ├── service/
│   │   │               │   ├── InterceptionService.java
│   │   │               │   ├── MemoryService.java
│   │   │               │   ├── IntelligenceService.java
│   │   │               │   ├── EmbeddingService.java
│   │   │               │   ├── GraphRAGService.java
│   │   │               │   ├── AuditService.java
│   │   │               │   └── LearningService.java
│   │   │               │
│   │   │               ├── repository/
│   │   │               │   ├── MemoryRepository.java
│   │   │               │   ├── AuditLogRepository.java
│   │   │               │   └── impl/
│   │   │               │       ├── MemoryRepositoryImpl.java
│   │   │               │       └── AuditLogRepositoryImpl.java
│   │   │               │
│   │   │               ├── mapper/
│   │   │               │   ├── MemoryMapper.java
│   │   │               │   └── AuditLogMapper.java
│   │   │               │
│   │   │               ├── exception/
│   │   │               │   ├── BrainSentryException.java
│   │   │               │   ├── MemoryNotFoundException.java
│   │   │               │   ├── ValidationException.java
│   │   │               │   └── GlobalExceptionHandler.java
│   │   │               │
│   │   │               └── util/
│   │   │                   ├── CypherQueryBuilder.java
│   │   │                   ├── ContextFormatter.java
│   │   │                   └── VectorUtils.java
│   │   │
│   │   └── resources/
│   │       ├── application.yml
│   │       ├── application-dev.yml
│   │       ├── application-prod.yml
│   │       └── logback-spring.xml
│   │
│   └── test/
│       └── java/
│           └── com/
│               └── integraltech/
│                   └── brainsentry/
│                       ├── controller/
│                       ├── service/
│                       ├── repository/
│                       └── integration/
│
├── docker/
│   └── docker-compose.yml
│
└── README.md
```

---

## Domain Models

### Memory

```java
package com.integraltech.brainsentry.domain;

import lombok.Data;
import lombok.Builder;
import java.time.Instant;
import java.util.List;
import java.util.Map;

@Data
@Builder
public class Memory {
    private String id;
    private String content;
    private String summary;
    private MemoryCategory category;
    private ImportanceLevel importance;
    private ValidationStatus validationStatus;
    
    // Vector embedding
    private float[] embedding;
    
    // Metadata
    private Map<String, Object> metadata;
    private List<String> tags;
    
    // Provenance
    private String sourceType;      // "conversation", "code_commit", "manual"
    private String sourceReference; // URL, commit hash, etc
    private String createdBy;
    
    // Timestamps
    private Instant createdAt;
    private Instant updatedAt;
    private Instant lastAccessedAt;
    
    // Version control
    private Integer version;
    
    // Usage tracking
    private Integer accessCount;
    private Integer injectionCount;
    private Integer helpfulCount;
    private Integer notHelpfulCount;
    
    // Code example (optional)
    private String codeExample;
    private String programmingLanguage;
    
    // Relationships (loaded separately)
    private List<String> relatedMemoryIds;
    
    // Computed fields
    private Double helpfulnessRate;
    private Double relevanceScore;
}
```

### MemoryRelationship

```java
package com.integraltech.brainsentry.domain;

import lombok.Data;
import lombok.Builder;
import java.time.Instant;

@Data
@Builder
public class MemoryRelationship {
    private String id;
    private String fromMemoryId;
    private String toMemoryId;
    private RelationshipType type; // USED_WITH, CONFLICTS_WITH, SUPERSEDES, RELATED_TO
    
    // Metadata
    private Integer frequency;      // How many times used together
    private String severity;        // For conflicts
    private Double strength;        // 0.0 to 1.0
    
    private Instant createdAt;
    private Instant lastUsedAt;
}
```

### AuditLog

```java
package com.integraltech.brainsentry.domain;

import lombok.Data;
import lombok.Builder;
import java.time.Instant;
import java.util.List;
import java.util.Map;

@Data
@Builder
public class AuditLog {
    private String id;
    private String eventType; // "context_injection", "memory_created", etc
    private Instant timestamp;
    
    // Context
    private String userId;
    private String sessionId;
    private String userRequest;
    
    // Decision
    private Map<String, Object> decision;
    private String reasoning;
    private Double confidence;
    
    // Data
    private Map<String, Object> inputData;
    private Map<String, Object> outputData;
    
    // Memory tracking
    private List<String> memoriesAccessed;
    private List<String> memoriesCreated;
    private List<String> memoriesModified;
    
    // Performance
    private Integer latencyMs;
    private Integer llmCalls;
    private Integer tokensUsed;
    
    // Outcome
    private String outcome; // "success", "failed", "rejected"
    private Map<String, Object> userFeedback;
}
```

### Enums

```java
package com.integraltech.brainsentry.domain.enums;

public enum MemoryCategory {
    DECISION,       // Architectural decision
    PATTERN,        // Code pattern
    ANTIPATTERN,    // What not to do
    DOMAIN,         // Business domain knowledge
    BUG,            // Bug and fix
    OPTIMIZATION,   // Performance optimization
    INTEGRATION     // External integration details
}

public enum ImportanceLevel {
    CRITICAL,   // Must always be followed
    IMPORTANT,  // Should be followed
    MINOR       // Nice to know
}

public enum RelationshipType {
    USED_WITH,      // Frequently used together
    CONFLICTS_WITH, // Contradicts
    SUPERSEDES,     // Replaces/obsoletes
    RELATED_TO      // General relation
}

public enum ValidationStatus {
    APPROVED,
    PENDING,
    FLAGGED,
    REJECTED
}
```

---

## API Endpoints

### 1. Interception API

#### POST /api/v1/intercept
Intercepta e enriquece um prompt com contexto.

**Request:**
```json
{
  "prompt": "Add validation method to OrderAgent",
  "userId": "edson",
  "sessionId": "sess_123",
  "context": {
    "project": "vendax",
    "file": "OrderAgent.java"
  }
}
```

**Response:**
```json
{
  "enhanced": true,
  "originalPrompt": "Add validation method to OrderAgent",
  "enhancedPrompt": "<system_context>...\n</system_context>\n\nAdd validation method to OrderAgent",
  "contextInjected": "🚨 Pattern: Agents must validate...\n⚠️ Decision: Use Spring Events...",
  "memoriesUsed": [
    {
      "id": "mem_001",
      "category": "PATTERN",
      "summary": "Agents must validate with BeanValidator",
      "relevanceScore": 0.92
    }
  ],
  "latencyMs": 342,
  "reasoning": "Detected OrderAgent component, found 2 critical patterns"
}
```

---

### 2. Memory Management API

#### GET /api/v1/memories
Lista todas as memórias com paginação e filtros.

**Query Params:**
- `page` (default: 0)
- `size` (default: 20)
- `category` (optional)
- `importance` (optional)
- `status` (optional)
- `search` (optional text search)

**Response:**
```json
{
  "content": [...],
  "page": 0,
  "size": 20,
  "totalElements": 156,
  "totalPages": 8
}
```

#### GET /api/v1/memories/{id}
Retorna detalhes completos de uma memória.

**Response:**
```json
{
  "id": "mem_001",
  "content": "Full content...",
  "summary": "Agents must validate with BeanValidator",
  "category": "PATTERN",
  "importance": "CRITICAL",
  "validationStatus": "APPROVED",
  "tags": ["validation", "agents", "spring"],
  "metadata": {...},
  "sourceType": "conversation",
  "sourceReference": "https://claude.ai/chat/abc123",
  "createdBy": "edson",
  "createdAt": "2025-01-10T09:15:00Z",
  "accessCount": 23,
  "helpfulnessRate": 0.92,
  "relatedMemories": [
    {
      "id": "mem_042",
      "summary": "Spring Events for communication",
      "relationshipType": "USED_WITH"
    }
  ]
}
```

#### POST /api/v1/memories
Cria uma nova memória.

**Request:**
```json
{
  "content": "Agents must validate input with BeanValidator before processing",
  "category": "PATTERN",
  "importance": "CRITICAL",
  "tags": ["validation", "agents"],
  "sourceType": "manual",
  "codeExample": "...",
  "programmingLanguage": "java"
}
```

#### PUT /api/v1/memories/{id}
Atualiza uma memória existente.

#### DELETE /api/v1/memories/{id}
Soft delete de uma memória.

---

### 3. Search API

#### POST /api/v1/memories/search
Busca avançada com vetores e filtros.

**Request:**
```json
{
  "query": "validation patterns",
  "categories": ["PATTERN", "DECISION"],
  "minImportance": "IMPORTANT",
  "limit": 5,
  "includeRelated": true
}
```

**Response:**
```json
{
  "results": [
    {
      "memory": {...},
      "score": 0.92,
      "relatedMemories": [...]
    }
  ],
  "queryEmbedding": [...],
  "searchTimeMs": 45
}
```

---

### 4. Audit API

#### GET /api/v1/audit/logs
Lista audit logs com filtros.

**Query Params:**
- `eventType`
- `userId`
- `startDate`
- `endDate`
- `outcome`

#### GET /api/v1/audit/memory/{id}/history
Retorna histórico completo de uma memória.

---

### 5. Stats API

#### GET /api/v1/stats/overview
Retorna estatísticas gerais do sistema.

**Response:**
```json
{
  "totalMemories": 1834,
  "memoriesByCategory": {
    "PATTERN": 456,
    "DECISION": 123,
    ...
  },
  "memoriesByImportance": {
    "CRITICAL": 47,
    "IMPORTANT": 312,
    "MINOR": 1475
  },
  "requestsToday": 1247,
  "injectionRate": 0.34,
  "avgLatencyMs": 287,
  "helpfulnessRate": 0.89
}
```

#### GET /api/v1/stats/top-patterns
Retorna padrões mais usados.

**Response:**
```json
{
  "topPatterns": [
    {
      "memoryId": "mem_001",
      "summary": "Agent Validation Pattern",
      "usageCount": 89,
      "helpfulnessRate": 0.92
    }
  ],
  "period": "last_7_days"
}
```

---

## Services

### InterceptionService

```java
package com.integraltech.brainsentry.service;

@Service
@Slf4j
public class InterceptionService {
    
    private final MemoryService memoryService;
    private final IntelligenceService intelligenceService;
    private final GraphRAGService graphRAGService;
    private final AuditService auditService;
    
    /**
     * Intercepta e enriquece um prompt
     */
    public InterceptResponse interceptAndEnhance(InterceptRequest request) {
        long startTime = System.currentTimeMillis();
        
        // Quick check (fast path)
        if (!quickCheck(request.getPrompt())) {
            return InterceptResponse.passThrough(request.getPrompt());
        }
        
        // Deep analysis (LLM)
        RelevanceAnalysis analysis = intelligenceService
            .analyzeRelevance(request.getPrompt(), request.getContext());
        
        if (!analysis.isNeedsContext()) {
            return InterceptResponse.passThrough(request.getPrompt());
        }
        
        // Search relevant memories (GraphRAG)
        List<Memory> memories = graphRAGService.searchWithContext(
            request.getPrompt(),
            analysis.getCategories(),
            5 // limit
        );
        
        // Format context
        String context = formatContext(memories);
        
        // Inject
        String enhancedPrompt = injectContext(request.getPrompt(), context);
        
        // Audit
        long latency = System.currentTimeMillis() - startTime;
        auditService.logInterception(request, memories, latency);
        
        return InterceptResponse.builder()
            .enhanced(true)
            .originalPrompt(request.getPrompt())
            .enhancedPrompt(enhancedPrompt)
            .contextInjected(context)
            .memoriesUsed(toMemoryRefs(memories))
            .latencyMs((int) latency)
            .reasoning(analysis.getReasoning())
            .build();
    }
    
    private boolean quickCheck(String prompt) {
        // Regex-based fast filtering
        return RELEVANCE_PATTERNS.stream()
            .anyMatch(pattern -> pattern.matcher(prompt).find());
    }
    
    private String formatContext(List<Memory> memories) {
        // Format memories for injection
        // Group by importance, limit tokens
    }
    
    private String injectContext(String originalPrompt, String context) {
        return String.format(
            "<system_context>\n%s\n</system_context>\n\n<user_request>\n%s\n</user_request>",
            context,
            originalPrompt
        );
    }
}
```

### MemoryService

```java
package com.integraltech.brainsentry.service;

@Service
@Slf4j
public class MemoryService {
    
    private final MemoryRepository memoryRepository;
    private final EmbeddingService embeddingService;
    private final IntelligenceService intelligenceService;
    
    /**
     * Cria nova memória com análise automática
     */
    @Transactional
    public Memory createMemory(CreateMemoryRequest request) {
        // Generate embedding
        float[] embedding = embeddingService.embed(request.getContent());
        
        // Auto-analyze if not provided
        if (request.getImportance() == null) {
            ImportanceAnalysis analysis = intelligenceService
                .analyzeImportance(request.getContent());
            request.setImportance(analysis.getImportance());
            request.setCategory(analysis.getCategory());
        }
        
        Memory memory = Memory.builder()
            .id(generateId())
            .content(request.getContent())
            .summary(request.getSummary())
            .category(request.getCategory())
            .importance(request.getImportance())
            .embedding(embedding)
            .tags(request.getTags())
            .sourceType(request.getSourceType())
            .createdBy(request.getCreatedBy())
            .createdAt(Instant.now())
            .version(1)
            .validationStatus(ValidationStatus.PENDING)
            .build();
        
        return memoryRepository.save(memory);
    }
    
    /**
     * Atualiza memória com versionamento
     */
    @Transactional
    public Memory updateMemory(String id, UpdateMemoryRequest request) {
        Memory existing = memoryRepository.findById(id)
            .orElseThrow(() -> new MemoryNotFoundException(id));
        
        // Archive current version
        memoryRepository.archiveVersion(existing);
        
        // Update
        existing.setContent(request.getContent());
        existing.setSummary(request.getSummary());
        existing.setVersion(existing.getVersion() + 1);
        existing.setUpdatedAt(Instant.now());
        
        // Re-generate embedding if content changed
        if (request.getContent() != null) {
            float[] newEmbedding = embeddingService.embed(request.getContent());
            existing.setEmbedding(newEmbedding);
        }
        
        return memoryRepository.save(existing);
    }
    
    /**
     * Incrementa contadores de uso
     */
    public void recordUsage(String memoryId, boolean helpful) {
        Memory memory = memoryRepository.findById(memoryId)
            .orElseThrow(() -> new MemoryNotFoundException(memoryId));
        
        memory.setAccessCount(memory.getAccessCount() + 1);
        memory.setInjectionCount(memory.getInjectionCount() + 1);
        memory.setLastAccessedAt(Instant.now());
        
        if (helpful) {
            memory.setHelpfulCount(memory.getHelpfulCount() + 1);
        } else {
            memory.setNotHelpfulCount(memory.getNotHelpfulCount() + 1);
        }
        
        memoryRepository.save(memory);
    }
}
```

### GraphRAGService

```java
package com.integraltech.brainsentry.service;

@Service
@Slf4j
public class GraphRAGService {
    
    private final MemoryRepository memoryRepository;
    private final EmbeddingService embeddingService;
    
    /**
     * GraphRAG: Vector search + Graph expansion
     */
    public List<Memory> searchWithContext(
        String query,
        List<MemoryCategory> categories,
        int limit
    ) {
        // Generate query embedding
        float[] queryEmbedding = embeddingService.embed(query);
        
        // Vector search with graph expansion
        String cypherQuery = buildGraphRAGQuery(categories, limit);
        
        return memoryRepository.executeGraphQuery(
            cypherQuery,
            Map.of(
                "embedding", queryEmbedding,
                "categories", categories.stream()
                    .map(Enum::name)
                    .collect(Collectors.toList())
            )
        );
    }
    
    private String buildGraphRAGQuery(List<MemoryCategory> categories, int limit) {
        return String.format("""
            CALL db.idx.vector.queryNodes('Memory', 'embedding', %d, $embedding)
            YIELD node, score
            WHERE score > 0.7 
              AND node.category IN $categories
            
            WITH node, score
            OPTIONAL MATCH (node)-[r:USED_WITH|RELATED_TO]->(related:Memory)
            WHERE related.importance IN ['CRITICAL', 'IMPORTANT']
            
            RETURN node, score, collect(related) as relatedMemories
            ORDER BY score DESC
            """, limit);
    }
}
```

### IntelligenceService

```java
package com.integraltech.brainsentry.service;

@Service
@Slf4j
public class IntelligenceService {
    
    private final LlamaModel llamaModel;
    
    /**
     * Analisa se prompt precisa de contexto
     */
    public RelevanceAnalysis analyzeRelevance(String prompt, Map<String, Object> context) {
        String analysisPrompt = buildRelevancePrompt(prompt, context);
        
        String response = llamaModel.generate(analysisPrompt, GenerateParams.builder()
            .temperature(0.3)
            .maxTokens(300)
            .build());
        
        // Parse JSON response
        return parseRelevanceAnalysis(response);
    }
    
    /**
     * Analisa importância de conteúdo
     */
    public ImportanceAnalysis analyzeImportance(String content) {
        String prompt = String.format("""
            Analyze if this content is important enough to remember.
            
            Content: %s
            
            Respond in JSON:
            {
              "shouldRemember": true/false,
              "importance": "CRITICAL|IMPORTANT|MINOR",
              "category": "DECISION|PATTERN|ANTIPATTERN|...",
              "reasoning": "..."
            }
            """, content);
        
        String response = llamaModel.generate(prompt, GenerateParams.builder()
            .temperature(0.3)
            .maxTokens(500)
            .build());
        
        return parseImportanceAnalysis(response);
    }
}
```

---

## Configuration

### application.yml

```yaml
server:
  port: 8080
  servlet:
    context-path: /api

spring:
  application:
    name: brain-sentry
  
  redis:
    host: ${REDIS_HOST:localhost}
    port: ${REDIS_PORT:6379}
    password: ${REDIS_PASSWORD:}
    database: 0
    timeout: 2000ms
    lettuce:
      pool:
        max-active: 20
        max-idle: 10
        min-idle: 5

brain-sentry:
  # FalkorDB
  graph:
    name: brain_sentry
    
  # LLM
  llm:
    model-path: ${LLM_MODEL_PATH:./models/qwen2.5-7b.gguf}
    context-size: 4096
    gpu-layers: ${GPU_LAYERS:35}
    
  # Embeddings
  embedding:
    model: all-MiniLM-L6-v2
    dimensions: 384
    
  # Interception
  interception:
    quick-check-enabled: true
    deep-analysis-enabled: true
    max-context-tokens: 1000
    
  # Memory management
  memory:
    auto-capture: true
    auto-importance: true
    obsolete-threshold-days: 90

logging:
  level:
    com.integraltech.brainsentry: DEBUG
    org.springframework: INFO
    redis.clients: WARN
```

### RedisConfig.java

```java
package com.integraltech.brainsentry.config;

@Configuration
public class RedisConfig {
    
    @Value("${spring.redis.host}")
    private String host;
    
    @Value("${spring.redis.port}")
    private int port;
    
    @Bean
    public JedisPoolConfig jedisPoolConfig() {
        JedisPoolConfig config = new JedisPoolConfig();
        config.setMaxTotal(20);
        config.setMaxIdle(10);
        config.setMinIdle(5);
        config.setTestOnBorrow(true);
        return config;
    }
    
    @Bean
    public JedisPool jedisPool(JedisPoolConfig poolConfig) {
        return new JedisPool(poolConfig, host, port);
    }
}
```

---

## Database Schema (FalkorDB)

### Graph Schema

```cypher
// Node: Memory
CREATE (:Memory {
  id: String,
  content: String,
  summary: String,
  category: String,
  importance: String,
  validationStatus: String,
  embedding: Array<Float>,
  metadata: Map,
  tags: Array<String>,
  sourceType: String,
  sourceReference: String,
  createdBy: String,
  createdAt: Long,
  updatedAt: Long,
  lastAccessedAt: Long,
  version: Integer,
  accessCount: Integer,
  injectionCount: Integer,
  helpfulCount: Integer,
  notHelpfulCount: Integer,
  codeExample: String,
  programmingLanguage: String
})

// Relationships
CREATE (:Memory)-[:USED_WITH {frequency: Integer, lastUsedAt: Long}]->(:Memory)
CREATE (:Memory)-[:CONFLICTS_WITH {severity: String, detectedAt: Long}]->(:Memory)
CREATE (:Memory)-[:SUPERSEDES {date: Long, reason: String}]->(:Memory)
CREATE (:Memory)-[:RELATED_TO {strength: Float}]->(:Memory)

// Indexes
CREATE INDEX ON :Memory(id)
CREATE INDEX ON :Memory(category)
CREATE INDEX ON :Memory(importance)
CREATE VECTOR INDEX ON :Memory(embedding) WITH {dimension: 384, similarityFunction: 'cosine'}
```

### Initialization Script

```java
@Component
@Slf4j
public class FalkorDBInitializer implements ApplicationRunner {
    
    private final JedisPool jedisPool;
    
    @Override
    public void run(ApplicationArguments args) {
        try (Jedis jedis = jedisPool.getResource()) {
            // Create indexes
            String createIndexes = """
                CREATE INDEX ON :Memory(id)
                CREATE INDEX ON :Memory(category)
                CREATE INDEX ON :Memory(importance)
                CREATE VECTOR INDEX ON :Memory(embedding) 
                  WITH {dimension: 384, similarityFunction: 'cosine'}
                """;
            
            jedis.graphQuery("brain_sentry", createIndexes);
            log.info("FalkorDB initialized successfully");
        } catch (Exception e) {
            log.error("Failed to initialize FalkorDB", e);
        }
    }
}
```

---

## Security

### JWT Authentication

```java
@Configuration
@EnableWebSecurity
public class SecurityConfig {
    
    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        http
            .csrf().disable()
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/api/v1/public/**").permitAll()
                .requestMatchers("/actuator/health").permitAll()
                .anyRequest().authenticated()
            )
            .oauth2ResourceServer(OAuth2ResourceServerConfigurer::jwt);
        
        return http.build();
    }
}
```

### Rate Limiting

```java
@Component
public class RateLimitingFilter extends OncePerRequestFilter {
    
    private final LoadingCache<String, AtomicInteger> requestCounts;
    
    @Override
    protected void doFilterInternal(
        HttpServletRequest request,
        HttpServletResponse response,
        FilterChain filterChain
    ) throws ServletException, IOException {
        
        String clientId = getClientId(request);
        AtomicInteger count = requestCounts.get(clientId);
        
        if (count.incrementAndGet() > MAX_REQUESTS_PER_MINUTE) {
            response.setStatus(HttpStatus.TOO_MANY_REQUESTS.value());
            return;
        }
        
        filterChain.doFilter(request, response);
    }
}
```

---

## Testing Strategy

### Unit Tests
```java
@SpringBootTest
class MemoryServiceTest {
    
    @MockBean
    private MemoryRepository memoryRepository;
    
    @MockBean
    private EmbeddingService embeddingService;
    
    @Autowired
    private MemoryService memoryService;
    
    @Test
    void shouldCreateMemoryWithEmbedding() {
        // Given
        CreateMemoryRequest request = CreateMemoryRequest.builder()
            .content("Test content")
            .category(MemoryCategory.PATTERN)
            .build();
        
        float[] mockEmbedding = new float[]{0.1f, 0.2f, 0.3f};
        when(embeddingService.embed(any())).thenReturn(mockEmbedding);
        
        // When
        Memory result = memoryService.createMemory(request);
        
        // Then
        assertNotNull(result.getId());
        assertArrayEquals(mockEmbedding, result.getEmbedding());
        verify(memoryRepository).save(any(Memory.class));
    }
}
```

### Integration Tests
```java
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@Testcontainers
class InterceptionControllerIT {
    
    @Container
    static GenericContainer<?> redis = new GenericContainer<>("falkordb/falkordb:latest")
        .withExposedPorts(6379);
    
    @Autowired
    private TestRestTemplate restTemplate;
    
    @Test
    void shouldInterceptAndEnhancePrompt() {
        // Given
        InterceptRequest request = new InterceptRequest();
        request.setPrompt("Add method to OrderAgent");
        
        // When
        ResponseEntity<InterceptResponse> response = restTemplate.postForEntity(
            "/api/v1/intercept",
            request,
            InterceptResponse.class
        );
        
        // Then
        assertEquals(HttpStatus.OK, response.getStatusCode());
        assertTrue(response.getBody().isEnhanced());
    }
}
```

---

**Document Status:** ✅ Complete  
**Ready for:** Phase 1 implementation  
**Dependencies:** FalkorDB, Qwen 2.5-7B model
