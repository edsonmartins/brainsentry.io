# Brain Sentry MCP Server API Documentation

## Overview

The Brain Sentry MCP (Model Context Protocol) Server provides AI agents with tools, resources, and prompts for interacting with the Brain Sentry memory system.

**Base URL:** `http://localhost:8080/api/mcp`

**Version:** 1.0.0

**Last Updated:** 2026-01-19

---

## Authentication & Multi-Tenancy

### Tenant Identification

All MCP operations require a `tenantId` parameter for multi-tenancy isolation.

**Tenant ID Validation:**
- Format: Alphanumeric characters, dashes, and underscores only
- Length: Maximum 64 characters
- Default: `"default"` when not specified

```json
{
  "tenantId": "my-organization"
}
```

**Example Tenant IDs:**
- `"default"` - Default tenant
- `"acme-corp"` - Organization tenant
- `"user-john-doe"` - User-specific tenant

---

## Tools

### 1. create_memory

Create a new memory in the Brain Sentry system.

**Endpoint:** `POST /api/mcp/tools/create_memory`

**Request:**
```json
{
  "content": "The content of the memory to be stored",
  "summary": "A brief summary of the memory",
  "category": "DECISION",
  "importance": "CRITICAL",
  "tags": ["architecture", "microservices"],
  "tenantId": "my-organization"
}
```

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `content` | string | Yes | The full content of the memory |
| `tenantId` | string | Yes | The tenant ID for isolation |
| `summary` | string | No | Brief summary of the memory |
| `category` | enum | No | Memory category (see below) |
| `importance` | enum | No | Importance level (see below) |
| `tags` | array | No | Tags for categorization |

**Categories:**
- `DECISION` - Architectural or technical decision
- `PATTERN` - Design pattern
- `ANTIPATTERN` - Anti-pattern to avoid
- `DOMAIN` - Domain knowledge
- `BUG` - Bug fix or workaround
- `OPTIMIZATION` - Performance optimization
- `INTEGRATION` - Integration pattern

**Importance Levels:**
- `CRITICAL` - Critical information
- `IMPORTANT` - Important but not critical
- `MINOR` - Minor information

**Response:**
```json
{
  "success": true,
  "memoryId": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Memory created successfully",
  "tenantId": "my-organization"
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Invalid tenant ID format",
  "errorCode": "validation",
  "errorCategory": "VALIDATION",
  "timestamp": "2026-01-19T10:00:00Z"
}
```

---

### 2. search_memories

Search memories using semantic search.

**Endpoint:** `POST /api/mcp/tools/search_memories`

**Request:**
```json
{
  "query": "how to handle authentication in microservices",
  "limit": 10,
  "tenantId": "my-organization"
}
```

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `query` | string | Yes | The search query |
| `tenantId` | string | Yes | The tenant ID |
| `limit` | number | No | Max results (default: 10) |

**Response:**
```json
{
  "success": true,
  "count": 3,
  "tenantId": "my-organization",
  "memories": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "summary": "JWT authentication pattern for microservices",
      "content": "Full content here...",
      "category": "PATTERN",
      "importance": "CRITICAL",
      "relevanceScore": 0.95
    }
  ]
}
```

---

### 3. get_memory

Retrieve a specific memory by ID.

**Endpoint:** `POST /api/mcp/tools/get_memory`

**Request:**
```json
{
  "memoryId": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "my-organization"
}
```

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `memoryId` | string | Yes | The memory ID |
| `tenantId` | string | Yes | The tenant ID |

**Response:**
```json
{
  "success": true,
  "tenantId": "my-organization",
  "memory": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "content": "Full memory content",
    "summary": "Brief summary",
    "category": "DECISION",
    "importance": "CRITICAL",
    "tags": ["tag1", "tag2"],
    "createdAt": "2026-01-19T10:00:00Z",
    "updatedAt": "2026-01-19T10:00:00Z"
  }
}
```

**Error Response (Tenant Mismatch):**
```json
{
  "success": false,
  "error": "Memory belongs to a different tenant: other-org",
  "errorCode": "authorization",
  "errorCategory": "AUTHORIZATION"
}
```

---

### 4. intercept_prompt

Intercept and enhance a prompt with relevant memory context.

**Endpoint:** `POST /api/mcp/tools/intercept_prompt`

**Request:**
```json
{
  "prompt": "Create a new service for user management",
  "sessionId": "session-123",
  "userId": "user-456",
  "tenantId": "my-organization",
  "maxTokens": 500,
  "forceDeepAnalysis": false,
  "context": {
    "project": "brain-sentry",
    "filePath": "src/main/java/com/example/UserService.java"
  }
}
```

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt` | string | Yes | The original user prompt |
| `tenantId` | string | Yes | The tenant ID |
| `sessionId` | string | No | Session ID for tracking |
| `userId` | string | No | User ID making the request |
| `maxTokens` | number | No | Max tokens to inject (default: 500) |
| `forceDeepAnalysis` | boolean | No | Skip quick check (default: false) |
| `context` | object | No | Additional context |

**Response:**
```json
{
  "success": true,
  "enhanced": true,
  "originalPrompt": "Create a new service for user management",
  "enhancedPrompt": "<system_context>...</system_context>\n\nCreate a new service...",
  "contextInjected": "<system_context>\n[1] CRITICAL - PATTERN\n    Service layer pattern for brain-sentry\n</system_context>",
  "memoriesUsed": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "summary": "Service layer pattern",
      "category": "PATTERN",
      "importance": "CRITICAL",
      "relevanceScore": 0.92,
      "excerpt": "All services should follow the standard pattern..."
    }
  ],
  "latencyMs": 45,
  "reasoning": "Found 1 relevant memories",
  "confidence": 0.92,
  "tokensInjected": 125,
  "llmCalls": 1,
  "tenantId": "my-organization"
}
```

---

## Resources

### 1. list_memories

List all memories for a tenant.

**Endpoint:** `GET /api/mcp/resources/list_memories?tenantId=my-organization`

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tenantId` | string | Yes | The tenant ID |

**Response:**
```json
{
  "memories": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "summary": "Memory summary",
      "category": "PATTERN",
      "importance": "CRITICAL",
      "createdAt": "2026-01-19T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 0,
  "pageSize": 100
}
```

---

## Prompts

### Available Prompts

| Name | Description |
|------|-------------|
| `capture_pattern` | Capture code patterns from context |
| `extract_learning` | Extract learnings from development sessions |
| `summarize_discussion` | Summarize technical discussions |
| `context_builder` | Build comprehensive context from stored memories |

### Get Prompt

**Endpoint:** `GET /api/mcp/prompts/{name}`

**Response for `context_builder`:**
```
You are the Brain Sentry Context Builder. Your task is to construct
a comprehensive context block from the provided memories.

Input: A list of memories with the following fields:
- id: Unique identifier
- summary: Brief description
- category: Memory type (DECISION, PATTERN, ANTIPATTERN, DOMAIN, BUG, OPTIMIZATION, INTEGRATION)
- importance: Importance level (CRITICAL, IMPORTANT, MINOR)
- content: Full memory content
- codeExample: Optional code snippet
- programmingLanguage: Language for code example

Output Format:
<system_context>
The following relevant patterns and decisions from the system were found:

[1] {IMPORTANCE} - {CATEGORY}
    {summary}
    {full_content}
    {code_example_if_available}

</system_context>

Guidelines:
1. Prioritize CRITICAL and IMPORTANT memories first
2. Group related memories by category
3. Include code examples when available
4. Keep summaries concise but informative
5. Omit MINOR importance memories if context is too long
6. Target maximum of 500 tokens for the entire context block
```

---

## Error Handling

### Error Categories

| Category | Code | Description |
|----------|------|-------------|
| `VALIDATION` | `validation` | Invalid input parameters |
| `AUTHORIZATION` | `authorization` | Access denied |
| `NOT_FOUND` | `not_found` | Resource not found |
| `INTERNAL` | `internal` | Internal server error |
| `TENANT` | `tenant` | Tenant-related error |
| `RATE_LIMIT` | `rate_limit` | Too many requests |
| `TIMEOUT` | `timeout` | Operation timed out |

### Standard Error Response

```json
{
  "success": false,
  "error": "Error message description",
  "errorCode": "validation",
  "errorCategory": "VALIDATION",
  "errorType": "IllegalArgumentException",
  "timestamp": "2026-01-19T10:00:00Z",
  "details": {
    "context": "create_memory",
    "field": "tenantId"
  }
}
```

---

## Integration Example

### Java (Spring Boot)

```java
@RestController
@RequestMapping("/api/v1/mcp")
@RequiredArgsConstructor
public class McpController {

    private final McpServer mcpServer;

    @PostMapping("/tools/{toolName}")
    public ResponseEntity<String> executeTool(
        @PathVariable String toolName,
        @RequestBody Map<String, Object> arguments
    ) {
        String result = mcpServer.executeTool(toolName, arguments);
        return ResponseEntity.ok(result);
    }
}
```

### cURL

```bash
# Create a memory
curl -X POST http://localhost:8080/api/mcp/tools/create_memory \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Use virtual threads for I/O-bound operations",
    "summary": "Virtual thread pattern recommendation",
    "category": "PATTERN",
    "importance": "IMPORTANT",
    "tags": ["java21", "concurrency"],
    "tenantId": "my-org"
  }'

# Search memories
curl -X POST http://localhost:8080/api/mcp/tools/search_memories \
  -H "Content-Type: application/json" \
  -d '{
    "query": "virtual threads",
    "tenantId": "my-org"
  }'

# Intercept and enhance prompt
curl -X POST http://localhost:8080/api/mcp/tools/intercept_prompt \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "How should I handle concurrent requests?",
    "tenantId": "my-org"
  }'
```

### Python

```python
import requests

BASE_URL = "http://localhost:8080/api/mcp"

def create_memory(content, tenant_id):
    response = requests.post(
        f"{BASE_URL}/tools/create_memory",
        json={
            "content": content,
            "summary": "Brief summary",
            "category": "PATTERN",
            "importance": "IMPORTANT",
            "tenantId": tenant_id
        }
    )
    return response.json()

def search_memories(query, tenant_id):
    response = requests.post(
        f"{BASE_URL}/tools/search_memories",
        json={
            "query": query,
            "tenantId": tenant_id,
            "limit": 10
        }
    )
    return response.json()

def intercept_prompt(prompt, tenant_id):
    response = requests.post(
        f"{BASE_URL}/tools/intercept_prompt",
        json={
            "prompt": prompt,
            "tenantId": tenant_id
        }
    )
    return response.json()
```

---

## Architecture

### MCP Server Components

```
brain-sentry-backend/src/main/java/com/integraltech/brainsentry/mcp/
├── McpServer.java                 # Main MCP service
├── McpTenantContext.java          # Tenant isolation
├── McpErrorHandler.java           # Centralized error handling
├── tools/
│   ├── CreateMemoryTool.java     # create_memory tool
│   ├── SearchMemoryTool.java     # search_memories tool
│   ├── GetMemoryTool.java        # get_memory tool
│   └── InterceptPromptTool.java  # intercept_prompt tool
├── resources/
│   └── ListMemoriesResource.java # list_memories resource
└── prompts/
    ├── AgentPrompts.java         # Prompt templates
    └── ContextBuilderPrompt.java # Context builder
```

### Multi-Tenancy Flow

```
1. Request arrives with tenantId
   ↓
2. McpTenantContext validates and normalizes tenantId
   ↓
3. TenantContext.setTenantId(tenantId)
   ↓
4. Tool/Resource executes with tenant context
   ↓
5. TenantContext.clear() (finally block)
```

---

## Performance Considerations

- **Virtual Threads**: Tenant context uses ThreadLocal, compatible with Java 21+ virtual threads
- **Connection Pooling**: FalkorDB connections are pooled per tenant
- **Caching**: Embeddings are cached to reduce recomputation
- **Rate Limiting**: Consider implementing per-tenant rate limits

---

## Security

- **Tenant Isolation**: All operations are scoped to tenantId
- **Input Validation**: All parameters are validated before processing
- **Access Control**: Cross-tenant access is blocked
- **Audit Trail**: All operations are logged via AuditService

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/brainsentry/brainsentry/issues
- Documentation: https://docs.brainsentry.io
