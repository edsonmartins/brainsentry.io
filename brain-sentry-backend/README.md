# Brain Sentry Backend

Agent Memory System for Developers - Backend service.

## Overview

Brain Sentry is a next-generation Agent Memory System that goes beyond traditional RAG. It provides persistent, autonomous, and intelligent memory for developers using AI agents.

## Tech Stack

- **Java 25** with Spring Boot 4.0
- **PostgreSQL 16** for audit logs and user data
- **FalkorDB** for graph and vector storage
- **x-ai/grok-4.1-fast** via OpenRouter for LLM analysis
- **all-MiniLM-L6-v2** for embeddings (384 dimensions)
- **Maven** for build management

## Features

- 4 types of memory: Semantic, Episodic, Procedural, Associative
- Graph-native storage with FalkorDB
- Autonomous prompt interception and context injection
- Vector search for semantic memory retrieval
- Full audit trail for production requirements
- Multi-tenant support

## Quick Start

### Prerequisites

- Java 25+
- Maven 3.9+
- Docker & Docker Compose
- OpenRouter API key

### 1. Start Database Services

```bash
cd brain-sentry-backend/docker
docker-compose up -d
```

This starts:
- PostgreSQL 16 on port 5432
- FalkorDB on port 6379
- Adminer on port 8081 (optional)

### 2. Configure Environment

Create a `.env` file or set environment variables:

```bash
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=brainsentry
POSTGRES_USER=brainsentry
POSTGRES_PASSWORD=brainsentry_dev

# FalkorDB
FALKORDB_HOST=localhost
FALKORDB_PORT=6379

# OpenRouter API
OPENROUTER_API_KEY=your-api-key-here
LLM_MODEL=x-ai/grok-4.1-fast
```

### 3. Build and Run

```bash
cd brain-sentry-backend
mvn clean install
mvn spring-boot:run
```

The API will be available at `http://localhost:8080/api`

## API Endpoints

### Memory Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/memories` | Create a new memory |
| GET | `/v1/memories/{id}` | Get a memory by ID |
| GET | `/v1/memories` | List memories (paginated) |
| PUT | `/v1/memories/{id}` | Update a memory |
| DELETE | `/v1/memories/{id}` | Delete a memory |
| POST | `/v1/memories/search` | Semantic search |
| GET | `/v1/memories/by-category/{category}` | Filter by category |
| GET | `/v1/memories/by-importance/{importance}` | Filter by importance |
| GET | `/v1/memories/{id}/related` | Find related memories |
| POST | `/v1/memories/{id}/feedback` | Record feedback |

### Interception

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/intercept` | Intercept and enhance a prompt |

### Statistics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/stats/overview` | System overview statistics |
| GET | `/v1/stats/health` | Health check |

## Example Usage

### Create a Memory

```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Agents must validate input with BeanValidator before processing",
    "summary": "Always validate agent input",
    "category": "PATTERN",
    "importance": "CRITICAL",
    "tags": ["validation", "agents", "best-practice"],
    "sourceType": "manual",
    "tenantId": "default"
  }'
```

### Intercept a Prompt

```bash
curl -X POST http://localhost:8080/api/v1/intercept \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a new agent that processes user orders",
    "userId": "user-123",
    "sessionId": "session-456",
    "tenantId": "default",
    "context": {
      "project": "myapp",
      "file": "OrderAgent.java"
    }
  }'
```

## Development

### Running Tests

```bash
mvn test
```

### Code Coverage

```bash
mvn test jacoco:report
```

### Linting

```bash
mvn checkstyle:check
```

## Configuration

See `application.yml` for all configuration options.

Key configuration areas:
- Database connection settings
- OpenRouter API configuration
- Embedding model settings
- Interception behavior
- Multi-tenancy

## License

Copyright © 2025 IntegrAllTech. All rights reserved.
