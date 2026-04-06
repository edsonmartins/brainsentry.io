# Brain Sentry Go

AI Agent Memory System backend written in Go. Provides persistent, multi-tenant memory for AI agents with cognitive-inspired features: semantic search, knowledge graphs, temporal decay, spreading activation, automatic reflection, fact reconciliation, and cross-session learning.

## Prerequisites

- **Go** 1.25+
- **PostgreSQL** 16+ with pgvector extension
- **Redis** 7+ (for caching and async task scheduling)
- **FalkorDB** (optional, for knowledge graph features)
- **Docker** (optional, for containerized setup)

## Quick Start

### 1. Start infrastructure

```bash
make infra-up
```

This starts PostgreSQL, Redis, and FalkorDB via Docker Compose.

### 2. Run migrations

```bash
export DATABASE_URL="postgres://brainsentry:brainsentry@localhost:5432/brainsentry?sslmode=disable"
make migrate-up
```

### 3. Run the server

```bash
# Development mode
make dev

# Or build and run
make build && make run
```

The server starts on `http://localhost:8080`.

### 4. Verify

```bash
curl http://localhost:8080/health
# {"status":"UP"}
```

## TUI (Terminal User Interface)

Interactive terminal app for managing memories, sessions, and search — built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and the [Charm](https://charm.sh) ecosystem.

### Running

```bash
# Run directly
make tui

# Or build and run
make tui-build
./bin/brainsentry-tui
```

### Configuration

| Env Variable | Description | Default |
|---|---|---|
| `BRAINSENTRY_URL` | Backend API URL | `http://localhost:8080` |
| `BRAINSENTRY_TENANT` | Tenant ID | `default` |
| `BRAINSENTRY_TOKEN_FILE` | Path to persist auth token | `~/.brainsentry-token` |

### Features

- **Dashboard** — Stats cards, category/importance bar charts, system metrics, recent memories
- **Memory List** — Paginated table with sorting, colored category/importance badges, delete with confirmation
- **Memory Detail** — Markdown-rendered content (via glamour), metadata badges, tags, related memories, code examples with syntax highlighting
- **Memory Form** — Create/edit with category dropdown, tags, content textarea (powered by huh forms with Catppuccin theme)
- **Semantic Search** — Full-text query with relevance scores, results in sortable table
- **Sessions** — Session list with colored status (Active/Expired/Completed), memory/interception/note counts
- **Relationships** — Visual strength bars, relationship types, navigate to related memories
- **Help overlay** — Full keybinding reference (`?` to toggle)

### Keybindings

| Key | Action |
|---|---|
| `1-4` | Switch view (Dashboard, Memories, Search, Sessions) |
| `j/k` | Navigate up/down |
| `g/G` | Jump to top/bottom |
| `Enter` | Select/open |
| `Esc` | Back/cancel |
| `/` | Open search |
| `n` | New memory |
| `e` | Edit memory |
| `d` | Delete memory |
| `r` | View relationships |
| `Ctrl+D/U` | Page down/up |
| `Ctrl+S` | Save (in forms) |
| `?` | Toggle help |
| `q` | Quit |

### Stack

| Library | Purpose |
|---|---|
| [bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework (Elm architecture) |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling (CSS-like) |
| [huh](https://github.com/charmbracelet/huh) | Forms with Catppuccin theme |
| [glamour](https://github.com/charmbracelet/glamour) | Markdown rendering with Dracula theme |
| [bubble-table](https://github.com/evertras/bubble-table) | Tables with sorting, pagination, styled cells |
| [bubbles](https://github.com/charmbracelet/bubbles) | TextInput, Textarea, Viewport, Spinner |

## Docker

### Full stack

```bash
docker compose up -d
```

### Build only

```bash
make docker-build
```

## Configuration

Configuration is loaded from `config.yaml` with environment variable overrides:

| Env Variable | Description | Default |
|---|---|---|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | Database name | `brainsentry` |
| `DB_USER` | Database user | `brainsentry` |
| `DB_PASSWORD` | Database password | `brainsentry` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `FALKORDB_HOST` | FalkorDB host | `localhost` |
| `FALKORDB_PORT` | FalkorDB port | `6379` |
| `JWT_SECRET` | JWT signing secret | (required) |
| `BRAINSENTRY_AI_AGENTIC_MODEL_API_KEY` | OpenRouter API key | (optional) |
| `CONFIG_PATH` | Custom config file path | `config.yaml` |

## Features

### Core Memory Operations
- CRUD with soft delete and versioning
- Conflict detection on concurrent updates
- Batch import/export
- Paginated listing with filters by category, importance, tags

### Search & Retrieval
- **Semantic search** via pgvector embeddings (cosine similarity)
- **Full-text search** via PostgreSQL tsvector/tsquery
- **Composite hybrid scoring**: `sigmoid(α*sim_boost + β*token_overlap + γ*graph_proximity + δ*recency + ε*tag_match + ζ*importance)` with explainable score traces
- **Intent-aware retrieval planning** with reflection loops (multi-round gap-filling queries, 80% coverage target)
- **Pluggable rerankers**: NoOp, BM25, LLM-based, HybridScore
- **SimHash deduplication** (Hamming distance ≤ 3 boosts instead of inserting duplicates)

### Knowledge Graph (FalkorDB)
- Entity extraction from memories (NER via LLM)
- Automatic relationship detection (shared tags, entities)
- **Spreading activation**: BFS propagation through graph neighbors with per-hop decay
- **Natural language to Cypher** translation with retry-on-empty feedback loop
- **Louvain community detection** for discovering memory clusters
- Mention counter on relationships (popularity signal)
- GraphRAG multi-hop context enrichment

### Cognitive Memory Features
- **8 memory types**: Semantic, Episodic, Procedural, Personality, Preference, Thread, Task, Emotion
- **Automatic classification** via pattern-based classifier (keywords + regex + category/tag heuristics)
- **Temporal decay per type**: personality (0.001/day) to thread (0.05/day), with formula `baseScore × exp(-rate×age) × importanceFactor × log(frequency+1) × emotionalFactor`
- **Temporal supersession**: `valid_from`/`valid_to` fields, auto-supersede contradictory facts
- **Emotional weight** (-1 to +1) influencing decay and search relevance
- **Automatic reflection loop**: SimHash clustering → saliency scoring → LLM synthesis of higher-order insights
- **LLM fact reconciliation**: extract atomic facts → search similar → LLM decides ADD/UPDATE/DELETE/NONE per fact

### User Profiles
- **Static profile**: stable facts, preferences, expertise extracted from accumulated memories
- **Dynamic profile**: recent topics, active tasks, current context (last 7 days)
- Injectable as MCP prompt resource for system prompt personalization

### Context Injection & Interception
- Quick check (pattern matching) + deep analysis (LLM relevance scoring)
- **Token-budgeted context injection** with greedy packing by relevance
- **PII detection and masking** before sending to LLM (email, phone, SSN, credit cards, API keys, JWT)
- Cross-session context from previous sessions with 3-tier redaction (none/partial/full)
- Hindsight notes injection for error-related prompts

### Cross-Session Learning
- Session lifecycle hooks (start/end)
- Typed event recording: Decision, Bugfix, Feature, Refactor, Discovery, Change
- LLM-powered observation extraction with direct fallback
- Provenance chains with `superseded_by` tracking
- Cross-session context injection with configurable lookback window

### Async Task Scheduling (Redis Streams)
- Consumer group-based distributed processing
- 8 task types: entity extraction, summarization, reflection, reconciliation, embedding, graph update, decay cleanup, profile update
- Per-tenant priority weights
- Auto-recovery of stuck tasks via XClaim
- Retry with configurable max attempts
- Inline fallback when Redis unavailable

### External Connectors
- Pluggable connector interface with registry pattern
- **GitHub**: issues, PRs (via REST API)
- **Notion**: pages and databases
- **Google Drive**: documents and sheets
- **Web Crawler**: arbitrary URLs
- Intelligent document chunking (text with overlap, code by logical blocks)
- Async embedding via task scheduler

### Session Management
- In-memory cache with PostgreSQL persistence
- Auto-expire on idle timeout
- Background cleanup goroutine
- Session counters (memories, interceptions, notes)

### Observability & Resilience
- **Prometheus metrics** endpoint
- **LLM observability**: per-operation metrics, cost estimation, buffered event collection
- **Circuit breaker** for external services (CLOSED/OPEN/HALF_OPEN with exponential backoff + jitter)
- Structured logging via slog
- Audit trails with detailed event logging

### Cognitive Memory Pipeline (new)
- **Memory Compression** — LLM-driven extraction of `facts[]`, `concepts[]`, `narrative`, `importance` before storage
- **Self-Correcting LLM** — JSON output validation with retry and error feedback loop
- **Semantic/Procedural Consolidation** — Automatic extraction of cross-session facts and workflow procedures
- **Query Expansion** — LLM generates 3-5 query reformulations for better search recall
- **RRF Scoring** — Reciprocal Rank Fusion (`1/(k+rank)`, k=60) combining vector, text, and graph streams
- **Session Diversity** — Max 3 results per session to avoid result skew
- **Auto-Forget** — TTL expiry + contradiction detection (Jaccard >0.9) + low-value cleanup
- **Cascading Staleness** — BFS propagation of staleness through knowledge graph when memories are superseded
- **Sliding Window Enrichment** — Entity resolution (pronouns → names), preference extraction, context bridges
- **Fallback Chain Provider** — Sequential LLM provider fallback with per-provider circuit breakers

### Multi-Agent Coordination (new)
- **Actions** — Workflow items with status lifecycle (pending → in_progress → blocked → completed)
- **Leases** — Distributed locks on actions with TTL (1-60 min) and auto-expiry
- **Dependency Propagation** — Automatic unblocking when parent action completes
- **P2P Mesh Sync** — Peer registration with SSRF validation, scope-based data sharing, LWW merge

### Privacy & Security
- **Privacy Stripping** — Complete removal of secrets, env vars, `<private>` tags, GitHub PATs, AWS keys, Slack tokens before storage
- **PII Detection & Masking** — Email, phone, SSN, credit card, API key, JWT, IP address, private key detection
- JWT authentication with refresh tokens
- SSO/OIDC integration
- Multi-tenant isolation via context
- Webhook HMAC signatures

### Search Quality Benchmarking
- Metrics: Recall, Precision, F1, MRR, NDCG (graded relevance)
- Latency distribution: P50, P95, P99
- Throughput measurement (queries/sec)
- Per-category breakdown
- Synthetic dataset generation
- Formatted report output

## API Endpoints

### Auth
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/demo` | Demo login |
| POST | `/api/v1/auth/logout` | Logout |
| POST | `/api/v1/auth/refresh` | Refresh token |

### Memories
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/memories` | Create memory |
| GET | `/api/v1/memories` | List memories (paginated) |
| GET | `/api/v1/memories/{id}` | Get memory by ID |
| PUT | `/api/v1/memories/{id}` | Update memory |
| DELETE | `/api/v1/memories/{id}` | Delete memory |
| POST | `/api/v1/memories/search` | Search memories (hybrid scoring) |
| GET | `/api/v1/memories/by-category/{category}` | Filter by category |
| GET | `/api/v1/memories/by-importance/{importance}` | Filter by importance |
| POST | `/api/v1/memories/{id}/feedback` | Record feedback |

### Interception
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/intercept` | Intercept prompt and inject context |

### Relationships
| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/relationships` | List relationships |
| POST | `/api/v1/relationships` | Create relationship |
| POST | `/api/v1/relationships/bidirectional` | Create bidirectional |
| GET | `/api/v1/relationships/{memoryId}/related` | Get related memories |
| POST | `/api/v1/relationships/{memoryId}/suggest` | Auto-detect relationships |

### Entity Graph (requires FalkorDB)
| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/entity-graph/knowledge-graph` | Get knowledge graph |
| GET | `/api/v1/entity-graph/search` | Search entities |
| POST | `/api/v1/entity-graph/extract/{memoryId}` | Extract entities |
| POST | `/api/v1/entity-graph/extract-batch` | Batch extraction |

### Notes
| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/notes` | List notes |
| POST | `/api/v1/notes/analyze` | Analyze session |
| GET | `/api/v1/notes/hindsight` | List hindsight notes |
| POST | `/api/v1/notes/hindsight` | Create hindsight note |

### Compression
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/compression/compress` | Compress context |
| GET | `/api/v1/compression/session/{sessionId}` | Get session summaries |

### MCP (Model Context Protocol)
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/mcp/message` | JSON-RPC 2.0 message |
| POST | `/api/v1/mcp/sse` | SSE transport |
| POST | `/api/v1/mcp/batch` | Batch messages |

### Cognitive Pipeline (new)
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/auto-forget` | Run auto-forget (TTL + contradictions + low-value) |
| POST | `/api/v1/semantic/consolidate` | Extract semantic facts and procedural workflows |

### Actions & Leases (new)
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/actions` | Create action |
| GET | `/api/v1/actions` | List actions (filter by status) |
| GET | `/api/v1/actions/{id}` | Get action by ID |
| PUT | `/api/v1/actions/{id}/status` | Update action status |
| POST | `/api/v1/actions/{id}/lease` | Acquire lease (distributed lock) |
| DELETE | `/api/v1/actions/{id}/lease` | Release lease |

### P2P Mesh (new)
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/mesh/peers` | Register peer for sync |
| GET | `/api/v1/mesh/peers` | List registered peers |
| POST | `/api/v1/mesh/sync` | Sync scope with all peers |

### System
| Method | Path | Description |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |
| GET | `/swagger.json` | OpenAPI spec |
| GET | `/api/v1/stats/overview` | System stats |
| GET | `/api/v1/audit-logs` | Audit logs |

## MCP Tools

The MCP server exposes these tools for AI agents:

| Tool | Description |
|---|---|
| `intercept_prompt` | Intercept and enhance a prompt with relevant context |
| `create_memory` | Store a new memory |
| `search_memories` | Search memories by query (hybrid scoring) |
| `get_memory` | Retrieve a specific memory |
| `list_memories` | List all memories |
| `update_memory` | Update a memory |
| `delete_memory` | Delete a memory |

### MCP Resources

| Resource | Description |
|---|---|
| All Memories | Browse all stored memories |
| All Notes | Browse all notes |
| Hindsight Notes | Browse error-resolution notes |

### MCP Prompts

| Prompt | Description |
|---|---|
| `capture_pattern` | Capture a coding pattern or practice |
| `extract_learning` | Extract learnings from a session |
| `summarize_discussion` | Summarize a discussion |
| `context_builder` | Build context for a task from memories |
| `agent_context` | Get agent-ready context for a task |
| `memory_summary` | Generate a memory summary |
| `hindsight_review` | Review hindsight notes for a topic |

## Architecture

```
brain-sentry-go/
├── cmd/
│   ├── server/                 # HTTP server entrypoint
│   ├── tui/                    # Terminal UI app
│   │   ├── main.go             # TUI entrypoint
│   │   ├── app.go              # Root model, view routing, navigation stack
│   │   ├── keys/               # Vim-style keybindings
│   │   ├── theme/              # Catppuccin Mocha palette, reusable styles
│   │   ├── components/         # StatusBar, Spinner, Toast, Confirm, Charts
│   │   └── views/              # Login, Dashboard, MemoryList, MemoryDetail,
│   │                           # MemoryForm, Search, Sessions, Relationships, Help
│   └── cli/                    # CLI tool
├── internal/
│   ├── cache/                  # Redis cache layer
│   ├── client/                 # HTTP SDK client (auth, memories, sessions, etc.)
│   ├── config/                 # YAML + env config
│   ├── domain/                 # Domain models, enums, value objects
│   ├── dto/                    # Request/Response DTOs
│   ├── handler/                # HTTP handlers (Chi router)
│   ├── mcp/                    # MCP protocol server (JSON-RPC 2.0)
│   ├── middleware/             # Auth, CORS, Tenant, Rate Limit, Metrics
│   ├── repository/
│   │   ├── postgres/           # PostgreSQL repositories + migrations
│   │   └── graph/              # FalkorDB graph repositories
│   └── service/                # Business logic (38 service files)
│       ├── memory.go           # Core CRUD + hybrid search
│       ├── interception.go     # Context injection pipeline
│       ├── scoring.go          # Composite hybrid scoring
│       ├── classifier.go       # Auto memory type classification
│       ├── decay.go            # Temporal decay computation
│       ├── reconciliation.go   # LLM fact reconciliation
│       ├── retrieval_planner.go # Intent-aware retrieval
│       ├── profile.go          # User profile generation
│       ├── reflection.go       # Automatic reflection loop
│       ├── spreading_activation.go # Graph activation propagation
│       ├── nl_cypher.go        # Natural language → Cypher
│       ├── louvain.go          # Community detection
│       ├── cross_session.go    # Cross-session pipeline
│       ├── task_scheduler.go   # Redis Streams scheduler
│       ├── connector.go        # External connectors
│       ├── benchmark.go        # Benchmarking framework
│       ├── circuitbreaker.go   # Circuit breaker pattern
│       ├── reranker.go         # Pluggable rerankers
│       ├── llm_provider.go    # LLM abstraction + FallbackChainProvider
│       ├── memory_compression.go # LLM-driven fact/concept extraction
│       ├── query_expansion.go # Query reformulation for search
│       ├── auto_forget.go     # TTL + contradiction + low-value cleanup
│       ├── cascading_staleness.go # Graph staleness propagation
│       ├── semantic_memory.go # Semantic/procedural consolidation
│       ├── self_correcting_llm.go # JSON validation + retry
│       ├── rrf_scoring.go     # Reciprocal Rank Fusion
│       ├── search_quality.go  # IR metrics (Recall, NDCG, MRR)
│       ├── privacy_stripping.go # Secret/PII removal
│       ├── sliding_window.go  # Entity resolution + enrichment
│       ├── actions.go         # Multi-agent actions + leases
│       ├── mesh_sync.go       # P2P sync with SSRF validation
│       └── ...                 # + 20 more service files
├── pkg/tenant/                 # Tenant context utilities
├── docs/                       # Swagger annotations
├── config.yaml
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

### Key Design Decisions

- **Multi-tenancy** via `context.Context` — every query scoped by `tenant_id`
- **FalkorDB optional** — graph features degrade gracefully if unavailable
- **Redis optional** — caching and async tasks fall back to in-memory/inline processing
- **LLM integration** via OpenRouter API with retry, backoff, circuit breaker, and observability
- **MCP protocol** support for AI agent tool integration (JSON-RPC 2.0 + SSE)
- **Cognitive-inspired memory model** — temporal decay, spreading activation, reflection, fact reconciliation
- **Token-budgeted context** — greedy packing respects configurable token limits
- **PII-safe by default** — sensitive data masked before external LLM calls

## Testing

```bash
# Unit tests
make test

# With coverage report
make test-cover

# Run benchmarks
go test -bench=. ./internal/service/ -benchmem

# Integration tests (requires Docker)
go test -tags=integration ./internal/repository/postgres/ -v
```

## Comparison with Java Backend

| Metric | Java (Spring Boot) | Go |
|---|---|---|
| Binary size | ~100 MB+ (JAR + JVM) | 12 MB |
| Startup time | 3-8 seconds | <100ms |
| Memory usage | 256-512 MB | ~20-50 MB |
| Dependencies | JDK 23 + Maven | Single binary |
| Docker image | ~400 MB | ~20 MB |

## License

Proprietary - Integral Tech
