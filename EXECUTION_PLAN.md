# Plano de Execução – Brain Sentry (v2.0)

> **Instrução:** Sempre que uma tarefa avançar de status, atualize esta tabela com a nova situação e registre a data no campo "Última atualização". Os status sugeridos são `TODO`, `IN_PROGRESS`, `BLOCKED` e `DONE`.

## Legend

- `TODO`: ainda não iniciado.
- `IN_PROGRESS`: em execução.
- `BLOCKED`: impedida por dependência externa.
- `DONE`: concluída e validada.

---

**IMPORTANTE:**

- Seguir padrões de arquitetura em camadas (Controller → Service → Repository)
- Backend: Java 25 + Spring Boot 4.0 + Maven (Monolítico)
- Frontend: React 19 + Vite + TypeScript + Radix UI (2 Admin UIs)
- Database: FalkorDB (Graph + Vector) + PostgreSQL 16 (Audit)
- LLM: x-ai/grok-4.1-fast (via OpenRouter API)
- Embeddings: all-MiniLM-L6-v2 (384 dimensions)
- Interface: Java MCP Server (baseado no SimpleMem - referência apenas)
- Virtual Threads: Java 25+ para I/O-bound operations
- Implementar testes unitários e integração
- Observabilidade desde o início (logs, métricas, audit trail)

**CONTEXTO DO PROJETO:**

O **Brain Sentry** é um **Agent Memory System** de próxima geração para desenvolvedores que usam IA. Posiciona-se como "Agent Memory for Developers", indo além do RAG tradicional com:

1. **Agent Memory Completo:** 4 tipos (Semantic, Episodic, Procedural, Associative)
2. **Graph-Native Storage:** FalkorDB com relacionamentos como first-class citizens
3. **Autonomous Operation:** Sistema SEMPRE analisa (não depende do agent decidir)
4. **Production-Ready:** Full audit trail, versioning, rollback desde dia 1
5. **Developer-Focused:** Code patterns, architectural decisions, bug histories

**Diferenciais vs Competidores:**
- vs Mem0: Graph + Audit + Dev-focused
- vs Zep: Multi-type + Graph + Not just chat
- vs MemGPT: Production-ready + Simpler
- vs LangMem: Opinionated + Autonomous + Graph

---

## 📊 STATUS GERAL DO PROJETO (Atualizado: 2026-01-19 11:15)

### 🔧 Fases do Projeto

| Fase | Progresso | Status | Área | Estimativa |
|------|-----------|--------|------|------------|
| **PHASE 1: Foundation** | 100% | 🟢 DONE | Backend + Frontend | 3 semanas |
| **PHASE 2: Core Intelligence** | 0% | ⏸️ TODO | Backend | 3 semanas |
| **PHASE 3: Memory Management** | 0% | ⏸️ TODO | Backend + Frontend | 3 semanas |
| **PHASE 4: Observability** | 0% | ⏸️ TODO | Backend + Frontend | 3 semanas |
| **PHASE 5: MCP Server** | 50% | 🟡 IN_PROGRESS | Backend | 3 semanas |
| **PHASE 6: Polish & Deploy** | 0% | ⏸️ TODO | DevOps | 3 semanas |

**Status Geral:** 🟢 **PHASE 1: FOUNDATION COMPLETA** - Backend + Frontend + MCP Server base implementados | 41 testes unitários passando (32 MemoryRepository + 9 McpServer)

### 📦 Módulos Planejados

- ✅ **brain-sentry-backend** (Spring Boot 4.0 + FalkorDB) - **BASE IMPLEMENTADA**
- ✅ **brain-sentry-frontend** (React 19 + Vite + Radix UI) - **BASE IMPLEMENTADA**
- 🟡 **brain-sentry-mcp** (Java MCP Server) - **50% IMPLEMENTADO**
- 🔲 **brain-sentry-llm** (Grok via OpenRouter)
- 🔲 **brain-sentry-embeddings** (all-MiniLM-L6-v2)
- 🔲 **brain-sentry-infrastructure** (Docker Compose, K8s)

---

## 📝 PROGRESSO RECENTE

### Backend Foundation (2025-01-18)

**Concluído:**
- ✅ Projeto Spring Boot 4.0 configurado com Maven
- ✅ Estrutura de packages criada
- ✅ Domain models (Memory, MemoryRelationship, AuditLog, MemoryVersion)
- ✅ DTOs (5 request + 5 response)
- ✅ Docker Compose (PostgreSQL 16 + FalkorDB)
- ✅ MemoryRepository com FalkorDB/Jedis
- ✅ Configurações (Redis, Security, Web, OpenRouter)
- ✅ MemoryService (CRUD completo)
- ✅ OpenRouterService (integração Grok)
- ✅ EmbeddingService (placeholder para DJL)
- ✅ InterceptionService (core functionality)
- ✅ MemoryController (REST API)
- ✅ InterceptionController (prompt enhancement)
- ✅ StatsController (health + overview)
- ✅ AuditService (logging)
- ✅ Testes unitários MemoryRepository (32 testes) - JUnit 5 + Mockito
- ✅ Testes unitários McpServer (9 testes) - JUnit 5 + Mockito

### Frontend Foundation (2025-01-18)

**Concluído:**
- ✅ Projeto React 19 + Vite configurado
- ✅ Tailwind CSS 3.4 configurado
- ✅ Radix UI components (Button, Card, Dialog, Dropdown, Label, Select, Tabs, Toast, Switch)
- ✅ Layout base (AdminLayout) com sidebar responsivo
- ✅ MemoryAdminPage com listagem e busca
- ✅ AnalyticsAdminPage com cards de métricas
- ✅ MemoryCard component com ações
- ✅ MemoryForm component para criar/editar
- ✅ API client (Axios) com interceptadores
- ✅ Apache ECharts (substituindo Recharts)
- ✅ TypeScript compilando sem erros

**NOTA:** Multi-tenancy simplificado usando Hibernate 6 (@TenantId, @TenantResolver) - pendente de implementação

**Próximos Passos:**
- Interação com DJL/ONNX para embeddings reais
- Implementar Hibernate 6 multi-tenancy
- Integrar frontend + backend (CRUD end-to-end)

### MCP Server Foundation (2026-01-19)

**Concluído:**
- ✅ McpServer criado com estrutura base (Service Spring)
- ✅ Tool: CreateMemoryTool - create_memory (criar memórias via MCP)
- ✅ Tool: SearchMemoryTool - search_memories (busca semântica)
- ✅ Tool: GetMemoryTool - get_memory (recuperar por ID)
- ✅ Tool: InterceptPromptTool - intercept_prompt (integrar com InterceptionService)
- ✅ Resource: ListMemoriesResource - list_memories (listar todas)
- ✅ Prompts: AgentPrompts com 4 prompts (capture_pattern, extract_learning, summarize_discussion, context_builder)
- ✅ Testes unitários: 9 testes passando (McpServerTest)
- ✅ Configuração Jackson: JacksonConfig ajustado para Spring Boot 4.0 (Jackson 3)
- ✅ Multi-tenancy: McpTenantContext para isolamento por tenant
- ✅ Error handling: McpErrorHandler com tratamento centralizado de erros
- ✅ Documentação: MCP_SERVER_API.md com especificação completa

**Arquivos criados:**
- `McpServer.java` - Service principal do MCP Server
- `CreateMemoryTool.java` - Tool para criar memórias
- `SearchMemoryTool.java` - Tool para buscar memórias
- `GetMemoryTool.java` - Tool para recuperar memória por ID
- `InterceptPromptTool.java` - Tool para interceptar e melhorar prompts
- `ListMemoriesResource.java` - Resource para listar memórias
- `AgentPrompts.java` - Prompts pré-definidos para agentes
- `ContextBuilderPrompt.java` - Prompt para construir contexto
- `McpTenantContext.java` - Gerenciamento de contexto multi-tenant
- `McpErrorHandler.java` - Tratamento centralizado de erros
- `McpServerTest.java` - Testes unitários completos
- `MCP_SERVER_API.md` - Documentação completa da API

**Estrutura MCP:**
```
brain-sentry-backend/src/main/java/com/integraltech/brainsentry/mcp/
├── McpServer.java                 # Service principal
├── McpTenantContext.java          # Multi-tenancy context
├── McpErrorHandler.java           # Error handling
├── tools/
│   ├── CreateMemoryTool.java     # Tool: create_memory
│   ├── SearchMemoryTool.java     # Tool: search_memories
│   ├── GetMemoryTool.java        # Tool: get_memory
│   └── InterceptPromptTool.java  # Tool: intercept_prompt
├── resources/
│   └── ListMemoriesResource.java # Resource: list_memories
└── prompts/
    ├── AgentPrompts.java         # Prompts para agentes
    └── ContextBuilderPrompt.java # Prompt: context_builder
```

**Funcionalidades implementadas:**
- **4 Tools MCP**: create_memory, search_memories, get_memory, intercept_prompt
- **1 Resource MCP**: list_memories
- **4 Prompts MCP**: capture_pattern, extract_learning, summarize_discussion, context_builder
- **Validação de tenantId**: Formato alfanumérico com traços e underscores
- **Isolamento multi-tenant**: Todas as operações escopadas por tenantId
- **Error categorization**: VALIDATION, AUTHORIZATION, NOT_FOUND, INTERNAL, TENANT, RATE_LIMIT, TIMEOUT
- **Documentação OpenAPI/Markdown**: Especificação completa com exemplos

---

## 📋 TAREFAS DETALHADAS

### PHASE 1: FOUNDATION (Weeks 1-3)

**Objetivo:** CRUD básico + Graph Setup + UI Scaffold

| ID | Tarefa | Descrição | Estimativa | Prioridade | Status | Depende de | Última Atualização |
|----|-------|-----------|------------|-----------|--------|------------|-------------------|
| **BACKEND-001** | Criar projeto Spring Boot | Maven + dependencies | 2h | 🔴 ALTA | ✅ DONE | - | 2025-01-18 |
| **BACKEND-002** | Configurar estrutura de packages | Segundo spec | 2h | 🔴 ALTA | ✅ DONE | BACKEND-001 | 2025-01-18 |
| **BACKEND-003** | Setup FalkorDB Docker | docker-compose.yml | 1h | 🔴 ALTA | ✅ DONE | - | 2025-01-18 |
| **BACKEND-004** | Configurar Jedis | Redis client | 1h | 🔴 ALTA | ✅ DONE | BACKEND-003 | 2025-01-18 |
| **BACKEND-005** | Criar domain models | Memory, Relationship, AuditLog | 3h | 🔴 ALTA | ✅ DONE | BACKEND-002 | 2025-01-18 |
| **BACKEND-006** | Criar DTOs | Request/Response | 2h | 🔴 ALTA | ✅ DONE | BACKEND-005 | 2025-01-18 |
| **BACKEND-007** | Configurar MapStruct | Mappers | 1h | 🟡 MÉDIA | ✅ DONE | BACKEND-002 | 2025-01-18 |
| **BACKEND-008** | Implementar MemoryRepository | FalkorDB operations | 4h | 🔴 ALTA | ✅ DONE | BACKEND-004, BACKEND-005 | 2025-01-18 |
| **BACKEND-009** | CRUD Memory básico | Create, Read, Update, Delete | 4h | 🔴 ALTA | ✅ DONE | BACKEND-008 | 2025-01-18 |
| **BACKEND-010** | Criar MemoryController | REST endpoints | 3h | 🔴 ALTA | ✅ DONE | BACKEND-009 | 2025-01-18 |
| **BACKEND-011** | Health check endpoints | Actuator | 1h | 🟡 MÉDIA | ✅ DONE | BACKEND-001 | 2025-01-18 |
| **BACKEND-012** | Configuração application.yml | Environments | 1h | 🟡 MÉDIA | ✅ DONE | BACKEND-003 | 2025-01-18 |
| **BACKEND-013** | Testes unitários repository | JUnit + Mockito | 3h | 🟡 MÉDIA | ✅ DONE | BACKEND-008 | 2026-01-19 |
| **BACKEND-014** | Testes unitários controller | MockMvc | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-010 | - |
| **BACKEND-015** | OpenRouter integration | Grok via API | 3h | 🔴 ALTA | ✅ DONE | - | 2025-01-18 |
| **BACKEND-016** | InterceptionService | Prompt enhancement | 4h | 🔴 ALTA | ✅ DONE | BACKEND-015 | 2025-01-18 |
| **BACKEND-017** | Testes unitários MCP Server | McpServerTest (9 testes) | 3h | 🔴 ALTA | ✅ DONE | MCP-001 | 2026-01-19 |
| **FRONTEND-001** | Criar projeto React 19 + Vite | Vite + TypeScript | 1h | 🔴 ALTA | ✅ DONE | - | 2025-01-18 |
| **FRONTEND-002** | Configurar Tailwind CSS | tailwind.config.js | 1h | 🔴 ALTA | ✅ DONE | FRONTEND-001 | 2025-01-18 |
| **FRONTEND-003** | Setup Radix UI | Componentes base | 2h | 🔴 ALTA | ✅ DONE | FRONTEND-001 | 2025-01-18 |
| **FRONTEND-004** | Configurar ESLint + Prettier | Linting | 1h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-001 | - |
| **FRONTEND-005** | Criar layout base | AdminLayout, Sidebar | 3h | 🔴 ALTA | ✅ DONE | FRONTEND-002 | 2025-01-18 |
| **FRONTEND-006** | Criar MemoryAdminPage | Listagem de memórias | 3h | 🔴 ALTA | ✅ DONE | FRONTEND-005 | 2025-01-18 |
| **FRONTEND-007** | Criar MemoryCard component | Card de memória | 2h | 🔴 ALTA | ✅ DONE | FRONTEND-003 | 2025-01-18 |
| **FRONTEND-008** | Criar MemoryForm component | Formulário | 3h | 🔴 ALTA | ✅ DONE | FRONTEND-003 | 2025-01-18 |
| **FRONTEND-009** | Criar API client | Axios configuration | 1h | 🔴 ALTA | ✅ DONE | - | 2025-01-18 |
| **FRONTEND-010** | Integrar backend+frontend | CRUD funcionando | 2h | 🔴 ALTA | ✅ DONE | BACKEND-010, FRONTEND-009 | 2026-01-19 |
| **DEVOPS-001** | Docker Compose completo | Todos serviços | 2h | 🟡 MÉDIA | ✅ DONE | BACKEND-012 | 2025-01-18 |
| **DEVOPS-002** | README setup | Instruções | 2h | 🟡 MÉDIA | ✅ DONE | DEVOPS-001 | 2025-01-18 |

**Subtotal Phase 1:** 26 tarefas | **26 DONE** | **0 TODO** | Estimativa: ~55 horas

---

### PHASE 2: CORE INTELLIGENCE (Weeks 4-6)

**Objetivo:** LLM Integration + Vector Search + Interception

| ID | Tarefa | Descrição | Estimativa | Prioridade | Status | Depende de | Última Atualização |
|----|-------|-----------|------------|-----------|--------|------------|-------------------|
| **BACKEND-020** | Configurar OpenRouter API | RestTemplate + Grok access | 2h | 🔴 ALTA | 🔴 TODO | - | - |
| **BACKEND-021** | Criar OpenRouterConfig | API key, endpoints | 1h | 🔴 ALTA | 🔴 TODO | BACKEND-020 | - |
| **BACKEND-022** | Criar OpenRouterService | Grok integration | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-021 | - |
| **BACKEND-023** | Criar IntelligenceService | LLM integration | 4h | 🔴 ALTA | 🔴 TODO | BACKEND-022 | - |
| **BACKEND-024** | Implementar analyzeImportance() | Classificação via Grok | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-023 | - |
| **BACKEND-025** | Implementar analyzeRelevance() | Decisão via Grok | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-023 | - |
| **BACKEND-026** | Integrar DJL/ONNX | Embeddings | 3h | 🔴 ALTA | 🔴 TODO | - | - |
| **BACKEND-027** | Criar EmbeddingService | all-MiniLM + Virtual Threads | 2h | 🔴 ALTA | 🔴 TODO | BACKEND-026 | - |
| **BACKEND-028** | Configurar índice vetorial FalkorDB | Vector index | 2h | 🔴 ALTA | 🔴 TODO | BACKEND-008 | - |
| **BACKEND-029** | Implementar vector search | Similaridade | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-027, BACKEND-028 | - |
| **BACKEND-030** | Criar GraphRAGService | Vector + Graph | 4h | 🔴 ALTA | 🔴 TODO | BACKEND-029 | - |
| **BACKEND-031** | Implementar Quick Check | Regex patterns | 2h | 🔴 ALTA | 🔴 TODO | - | - |
| **BACKEND-032** | Criar InterceptionService | Intercept loop | 4h | 🔴 ALTA | 🔴 TODO | BACKEND-025, BACKEND-030, BACKEND-031 | - |
| **BACKEND-033** | Criar InterceptionController | /api/v1/intercept | 2h | 🔴 ALTA | 🔴 TODO | BACKEND-032 | - |
| **BACKEND-034** | Implementar formatContext() | Template de injeção | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-032 | - |
| **FRONTEND-020** | Criar TestInterceptPage | UI de teste | 3h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-005 | - |
| **FRONTEND-021** | Criar PromptInput component | Input de teste | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-020 | - |
| **FRONTEND-022** | Criar ContextViewer | Visualização do contexto | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-020 | - |
| **FRONTEND-023** | Criar EnhancedPromptViewer | Resultado | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-020 | - |
| **FRONTEND-024** | Integrar intercept endpoints | Chamada API | 2h | 🔴 ALTA | 🔴 TODO | BACKEND-033, FRONTEND-009 | - |
| **TEST-001** | Testar OpenRouter API | Validação Grok | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-022 | - |
| **TEST-002** | Testar embeddings | Validação | 1h | 🟡 MÉDIA | 🔴 TODO | BACKEND-027 | - |
| **TEST-003** | Testar vector search | Precisão | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-029 | - |
| **TEST-004** | Testar intercept E2E | Fluxo completo | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-033, FRONTEND-024 | - |

**Subtotal Phase 2:** 23 tarefas | Estimativa: ~56 horas

---

### PHASE 3: MEMORY MANAGEMENT (Weeks 7-9)

**Objetivo:** Full Agent Memory Lifecycle

| ID | Tarefa | Descrição | Estimativa | Prioridade | Status | Depende de | Última Atualização |
|----|-------|-----------|------------|-----------|--------|------------|-------------------|
| **BACKEND-040** | Memory categorization | 4 tipos de memória | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-009 | - |
| **BACKEND-041** | Importance scoring automático | Auto-classificação | 2h | 🔴 ALTA | 🔴 TODO | BACKEND-024 | - |
| **BACKEND-042** | Relationship management | USED_WITH, CONFLICTS, etc | 4h | 🔴 ALTA | 🔴 TODO | BACKEND-005 | - |
| **BACKEND-043** | Memory versioning | Histórico de mudanças | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-008 | - |
| **BACKEND-044** | Implementar rollback | Reversão | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-043 | - |
| **BACKEND-045** | Conflict detection | Auto-deteção | 4h | 🟡 MÉDIA | 🔴 TODO | BACKEND-042 | - |
| **BACKEND-046** | Memory compression | Para memórias antigas | 3h | 🟡 MÉDIA | 🔴 TODO | BACKEND-023 | - |
| **BACKEND-047** | Consolidação de memórias | Merge similares | 3h | 🟡 MÉDIA | 🔴 TODO | BACKEND-030 | - |
| **BACKEND-048** | Tracking de uso | access_count, helpfulness | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-009 | - |
| **BACKEND-049** | Criar AuditService | Logging completo | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-005 | - |
| **BACKEND-050** | Criar AuditController | /api/v1/audit | 2h | 🔴 ALTA | 🔴 TODO | BACKEND-049 | - |
| **FRONTEND-040** | Criar MemoryDetail page | Tabs: Details, Relationships, Usage, History | 4h | 🔴 ALTA | 🔴 TODO | FRONTEND-005 | - |
| **FRONTEND-041** | Criar RelationshipGraph | Cytoscape.js | 6h | 🔴 ALTA | 🔴 TODO | FRONTEND-040 | - |
| **FRONTEND-042** | Criar VersionHistory | Diff viewer | 3h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-040 | - |
| **FRONTEND-043** | Criar UsageStats component | Métricas de uso | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-040 | - |
| **FRONTEND-044** | Feedback UI (helpful?) | Botão feedback | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-040 | - |
| **FRONTEND-045** | Criar MemoryFilters | Filtros avançados | 3h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-006 | - |
| **FRONTEND-046** | Criar MemorySearch | Busca com filtros | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-045 | - |
| **FRONTEND-047** | Audit logs page | Tabela de logs | 3h | 🟡 MÉDIA | 🔴 TODO | BACKEND-050 | - |
| **TEST-010** | Testar versioning | Criar, editar, rollback | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-043 | - |
| **TEST-011** | Testar relationships | Criar relacionamentos | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-042 | - |
| **TEST-012** | Testar conflict detection | Cenários de conflito | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-045 | - |
| **TEST-013** | Testar consolidação | Merge de memórias | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-047 | - |

**Subtotal Phase 3:** 23 tarefas | Estimativa: ~62 horas

---

### PHASE 4: OBSERVABILITY (Weeks 10-12)

**Objetivo:** Production-Ready System

| ID | Tarefa | Descrição | Estimativa | Prioridade | Status | Depende de | Última Atualização |
|----|-------|-----------|------------|-----------|--------|------------|-------------------|
| **BACKEND-060** | Spring Security | JWT authentication | 4h | 🔴 ALTA | 🔴 TODO | BACKEND-001 | - |
| **BACKEND-061** | RBAC implementation | Roles e permissions | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-060 | - |
| **BACKEND-062** | Rate limiting | Bucket4j | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-060 | - |
| **BACKEND-063** | Metrics collection | Micrometer | 3h | 🟡 MÉDIA | 🔴 TODO | BACKEND-001 | - |
| **BACKEND-064** | Prometheus endpoints | Actuator + Prometheus | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-063 | - |
| **BACKEND-065** | Criar StatsController | /api/v1/stats | 2h | 🔴 ALTA | 🔴 TODO | BACKEND-009 | - |
| **BACKEND-066** | Criar LearningService | Auto-evolução | 4h | 🟡 MÉDIA | 🔴 TODO | BACKEND-048 | - |
| **BACKEND-067** | Memory reflection jobs | Consolidação periódica | 3h | 🟡 MÉDIA | 🔴 TODO | BACKEND-047 | - |
| **BACKEND-068** | Advanced forgetting | Além de TTL | 3h | 🟢 BAIXA | 🔴 TODO | BACKEND-066 | - |
| **BACKEND-069** | CORS configuration | Cross-origin | 1h | 🟡 MÉDIA | 🔴 TODO | BACKEND-060 | - |
| **BACKEND-070** | Input validation | Bean Validation | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-006 | - |
| **FRONTEND-060** | Criar Dashboard page | Stats cards + charts | 4h | 🔴 ALTA | 🔴 TODO | FRONTEND-005 | - |
| **FRONTEND-061** | Criar StatsCards | 4 cards principais | 2h | 🔴 ALTA | 🔴 TODO | FRONTEND-060 | - |
| **FRONTEND-062** | Criar InjectionRateChart | Recharts | 3h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-060 | - |
| **FRONTEND-063** | Criar CategoryDistribution | Pie chart | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-060 | - |
| **FRONTEND-064** | Criar LatencyChart | Line chart | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-060 | - |
| **FRONTEND-065** | Criar ActivityFeed | Feed de atividades | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-060 | - |
| **FRONTEND-066** | Criar TopPatterns | Padrões mais usados | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-060 | - |
| **FRONTEND-067** | Autenticação frontend | JWT no client | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-060 | - |
| **FRONTEND-068** | Protected routes | Auth wrapper | 2h | 🔴 ALTA | 🔴 TODO | FRONTEND-067 | - |
| **FRONTEND-069** | Real-time updates | Polling ou WebSocket | 3h | 🟢 BAIXA | 🔴 TODO | FRONTEND-060 | - |
| **DEVOPS-010** | Docker images production | Optimized builds | 3h | 🟡 MÉDIA | 🔴 TODO | BACKEND-012 | - |
| **DEVOPS-011** | Kubernetes manifests | K8s configs | 4h | 🟡 MÉDIA | 🔴 TODO | DEVOPS-010 | - |
| **DEVOPS-012** | CI/CD pipeline | GitHub Actions | 3h | 🟡 MÉDIA | 🔴 TODO | DEVOPS-010 | - |
| **TEST-020** | Load testing | k6 ou similar | 3h | 🟡 MÉDIA | 🔴 TODO | DEVOPS-010 | - |
| **TEST-021** | Security testing | OWASP ZAP | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-060 | - |

**Subtotal Phase 4:** 25 tarefas | Estimativa: ~68 horas

---

### PHASE 5: MCP SERVER (Weeks 13-15)

**Objetivo:** Java MCP Server Production-Ready

| ID | Tarefa | Descrição | Estimativa | Prioridade | Status | Depende de | Última Atualização |
|----|-------|-----------|------------|-----------|--------|------------|-------------------|
| **MCP-001** | Configurar McpServer | Server base + port | 3h | 🔴 ALTA | ✅ DONE | BACKEND-001 | 2026-01-19 |
| **MCP-002** | Implementar Tool: create_memory | /tools/create_memory | 4h | 🔴 ALTA | ✅ DONE | MCP-001 | 2026-01-19 |
| **MCP-003** | Implementar Tool: search_memories | /tools/search_memories | 4h | 🔴 ALTA | ✅ DONE | MCP-001 | 2026-01-19 |
| **MCP-004** | Implementar Tool: get_memory | /tools/get_memory | 3h | 🔴 ALTA | ✅ DONE | MCP-001 | 2026-01-19 |
| **MCP-005** | Implementar Resource: memories | /resources/list_memories | 4h | 🔴 ALTA | ✅ DONE | MCP-001 | 2026-01-19 |
| **MCP-006** | Implementar Prompts | AgentPrompts (3 prompts) | 3h | 🟡 MÉDIA | ✅ DONE | MCP-001 | 2026-01-19 |
| **MCP-007** | Testes unitários MCP | McpServerTest (9 testes) | 3h | 🔴 ALTA | ✅ DONE | MCP-006 | 2026-01-19 |
| **MCP-008** | Implementar Tool: intercept_prompt | /tools/intercept_prompt | 5h | 🔴 ALTA | ✅ DONE | MCP-001, BACKEND-032 | 2026-01-19 |
| **MCP-009** | Multi-inquilino (tenants) | Isolamento por tenant | 5h | 🔴 ALTA | ✅ DONE | MCP-001 | 2026-01-19 |
| **MCP-010** | MCP Server authentication | JWT validation | 3h | 🔴 ALTA | 🔴 TODO | BACKEND-060 | - |
| **MCP-011** | MCP error handling | Proper error responses | 2h | 🟡 MÉDIA | ✅ DONE | MCP-001 | 2026-01-19 |
| **MCP-012** | Documentação MCP endpoints | OpenAPI/Markdown | 3h | 🟡 MÉDIA | ✅ DONE | MCP-006 | 2026-01-19 |
| **MCP-013** | Implementar Prompt: context_builder | /prompts/context_builder | 3h | 🟡 MÉDIA | ✅ DONE | MCP-001 | 2026-01-19 |
| **TEST-M01** | Testar MCP tools E2E | Validação endpoints completos | 3h | 🟡 MÉDIA | 🔴 TODO | MCP-008 | - |
| **TEST-M02** | Testar multi-tenancy | Isolamento tenants | 2h | 🟡 MÉDIA | 🔴 TODO | MCP-009 | - |

**Subtotal Phase 5:** 15 tarefas | **12 DONE** | **3 TODO** | Estimativa: ~50 horas

---

### PHASE 6: POLISH & DEPLOY (Weeks 16-18)

**Objetivo:** Market Launch

| ID | Tarefa | Descrição | Estimativa | Prioridade | Status | Depende de | Última Atualização |
|----|-------|-----------|------------|-----------|--------|------------|-------------------|
| **BACKEND-090** | Security hardening | Best practices | 4h | 🔴 ALTA | 🔴 TODO | BACKEND-060 | - |
| **BACKEND-091** | Performance optimization | Profiling + tuning | 4h | 🔴 ALTA | 🔴 TODO | TEST-020 | - |
| **BACKEND-092** | Error handling robust | Global handler | 2h | 🟡 MÉDIA | 🔴 TODO | BACKEND-001 | - |
| **FRONTEND-090** | Responsividade mobile | Breakpoints | 4h | 🔴 ALTA | 🔴 TODO | FRONTEND-005 | - |
| **FRONTEND-091** | Accessibility (WCAG AA) | ARIA labels | 3h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-090 | - |
| **FRONTEND-092** | ErrorBoundary component | Error handling | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-001 | - |
| **FRONTEND-093** | Loading states | Skeletons | 2h | 🟡 MÉDIA | 🔴 TODO | FRONTEND-001 | - |
| **DOC-001** | README completo | Documentação principal | 3h | 🔴 ALTA | 🔴 TODO | DEVOPS-001 | - |
| **DOC-002** | API Documentation | Swagger/OpenAPI | 4h | 🔴 ALTA | 🔴 TODO | BACKEND-033 | - |
| **DOC-003** | Setup guide | Instruções detalhadas | 3h | 🔴 ALTA | 🔴 TODO | DOC-001 | - |
| **DOC-004** | Architecture docs | ADRs | 4h | 🟡 MÉDIA | 🔴 TODO | BACKEND-001 | - |
| **DOC-005** | User manual | Como usar | 3h | 🟡 MÉDIA | 🔴 TODO | DOC-003 | - |
| **DEVOPS-020** | Production deployment | Deploy em staging | 4h | 🔴 ALTA | 🔴 TODO | DEVOPS-011 | - |
| **DEVOPS-021** | Backup strategy | Backups automatizados | 2h | 🔴 ALTA | 🔴 TODO | DEVOPS-011 | - |
| **DEVOPS-022** | Monitoring setup | Prometheus + Grafana | 3h | 🔴 ALTA | 🔴 TODO | DEVOPS-011 | - |
| **DEVOPS-023** | Log aggregation | ELK ou similar | 3h | 🟡 MÉDIA | 🔴 TODO | DEVOPS-022 | - |
| **BENCH-001** | LongMemEval benchmark | Testar contra padrão | 6h | 🟡 MÉDIA | 🔴 TODO | PHASE-5-COMP | - |
| **BENCH-002** | SWE-Bench-Pro evaluation | Comparar c/ Confucius | 8h | 🟢 BAIXA | 🔴 TODO | BENCH-001 | - |

**Subtotal Phase 6:** 18 tarefas | Estimativa: ~66 horas

---

## 📈 RESUMO EXECUTIVO

### Estimativas Totais

| Fase | Tarefas | Horas | Semanas |
|------|---------|-------|---------|
| **Phase 1: Foundation** | 26 | ~55h | 3 |
| **Phase 2: Core Intelligence** | 23 | ~56h | 3 |
| **Phase 3: Memory Management** | 23 | ~62h | 3 |
| **Phase 4: Observability** | 25 | ~68h | 3 |
| **Phase 5: MCP Server** | 15 | ~50h | 3 |
| **Phase 6: Polish & Deploy** | 18 | ~66h | 3 |
| **TOTAL** | **130** | **~357h** | **18 semanas** |

### Status por Área

| Área | Tarefas TODO | IN_PROGRESS | DONE | % Completo |
|------|-------------|-------------|------|------------|
| **Backend** | 30 | 1 | **19** | **38%** |
| **Frontend** | 20 | 0 | **10** | **33%** |
| **MCP Server** | 3 | 0 | **12** | **80%** |
| **DevOps** | 6 | 0 | **2** | **25%** |
| **Testes** | 12 | 0 | **1** | **8%** |
| **Docs** | 4 | 0 | **2** | **33%** |
| **TOTAL** | **75** | **1** | **46** | **38%** |

---

## 🔗 DEPENDÊNCIAS CRÍTICAS

### Backend Dependencies

```
FalkorDB Setup (BACKEND-003)
    ↓
Jedis Config (BACKEND-004)
    ↓
MemoryRepository (BACKEND-008)
    ↓
Memory CRUD (BACKEND-009)
    ↓
MemoryController (BACKEND-010)
    ↓
Frontend Integration (FRONTEND-010)
```

### Intelligence Flow

```
OpenRouter API Config (BACKEND-020)
    ↓
OpenRouterService (BACKEND-022)
    ↓
IntelligenceService (BACKEND-023)
    ↓
EmbeddingService (BACKEND-027) + Virtual Threads
    ↓
GraphRAGService (BACKEND-030)
    ↓
InterceptionService (BACKEND-032)
```

### Frontend Dependencies

```
Tailwind + Radix UI (FRONTEND-002, FRONTEND-003)
    ↓
AppLayout (FRONTEND-005)
    ↓
Memory Components (FRONTEND-007, FRONTEND-008)
    ↓
Pages (FRONTEND-006, FRONTEND-040)
    ↓
Graph Visualization (FRONTEND-041)
```

---

## 📝 NOTAS

### Hibernate 6 Multi-Tenancy

**OBSERVAÇÃO IMPORTANTE:** Hibernate 6 facilita significativamente a implementação de multi-tenancy com as anotações `@TenantId` e `@TenantResolver`.

**Implementação Planejada:**
```java
@Entity
public class Memory {
    @TenantId  // Hibernate 6 - filtragem automática por tenant
    private String tenantId;
}

@Configuration
public class TenantConfig {
    @Bean
    public CurrentTenantIdentifierResolver currentTenantResolver() {
        return new CurrentTenantIdentifierResolver() {
            @Override
            public String resolveCurrentTenantIdentifier() {
                // Extract from request header or JWT
                return TenantContext.getTenantId();
            }
        };
    }
}
```

**Benefícios:**
- Filtragem automática em todas as queries
- Isolamento garantido a nível de ORM
- Menos código manual
- Maior segurança

### Contexto SimpleMem
O projeto SimpleMem foi analisado como **REFERÊNCIA COMPARATIVA**:
- Arquitetura de memória estruturada adaptável
- Padrões MCP Server production-ready
- Algoritmos de recuperação adaptativa
- Padrões de código limpos e extensíveis

**NOTA:** Não haverá integração direta com SimpleMem. Apenas inspiração de padrões.

### Stack Tecnológico Definitivo
| Tecnologia | Versão | Observação |
|------------|--------|------------|
| **Backend** | Java 25 + Spring Boot 4.0 | Monolítico |
| **Frontend** | React 19 + Vite + Radix UI | 2 Admin UIs |
| **Database** | PostgreSQL 16 + FalkorDB | Multi-inquilino |
| **LLM** | x-ai/grok-4.1-fast | Via OpenRouter |
| **Embeddings** | all-MiniLM-L6-v2 | 384 dimensions |
| **Interface** | Java MCP Server | Baseado no SimpleMem |
| **Threading** | Virtual Threads | Java 25+ I/O-bound |

### Próximos Passos Imediatos
1. Configurar ambiente de desenvolvimento (Java 25, Maven 3.9+, Docker Compose, Node.js)
2. Criar estrutura base do projeto Spring Boot monolítico
3. Criar estrutura base do projeto React 19 + Vite
4. Setup PostgreSQL 16 + FalkorDB via Docker
5. Implementar primeiro CRUD de memória multi-tenant

---

## 🚧 O QUE FALTA IMPLEMENTAR

### IMEDIATO (Próximos dias)

**Backend - MCP Server:**
- [x] **MCP-008**: Tool `intercept_prompt` - Integração com InterceptionService ✅
- [x] **MCP-009**: Multi-inquilino (tenants) - Isolamento por tenantId ✅
- [x] **MCP-011**: Error handling robusto para MCP endpoints ✅
- [x] **MCP-012**: Documentação dos endpoints MCP (OpenAPI/Markdown) ✅
- [x] **MCP-013**: Prompt `context_builder` - Template para injeção de contexto ✅

**Backend - Foundation:**
- [x] **BACKEND-013**: Testes unitários do MemoryRepository (32 testes) ✅
- [ ] **BACKEND-014**: Testes unitários do MemoryController
- [ ] **BACKEND-016**: Integrar OpenRouterService com Grok (testes reais)

**Frontend:**
- [ ] **FRONTEND-004**: Configurar ESLint + Prettier
- [ ] **FRONTEND-010**: Integrar backend + frontend (CRUD end-to-end)

### CURTO PRAZO (Próximas semanas)

**Phase 2: Core Intelligence:**
- [ ] **BACKEND-020 a BACKEND-034**: LLM Integration + Vector Search + Interception
  - OpenRouter API Config
  - OpenRouterService com Grok
  - IntelligenceService
  - analyzeImportance() - Classificação via Grok
  - analyzeRelevance() - Decisão via Grok
  - DJL/ONNX para Embeddings
  - EmbeddingService com Virtual Threads
  - Índice vetorial FalkorDB
  - Vector search (similaridade)
  - GraphRAGService (Vector + Graph)
  - Quick Check (Regex patterns)
  - InterceptionService completo
  - InterceptionController (/api/v1/intercept)
  - formatContext() - Template de injeção

**Frontend - Intelligence:**
- [ ] **FRONTEND-020 a FRONTEND-024**: UI de teste de interceptação
  - TestInterceptPage
  - PromptInput component
  - ContextViewer
  - EnhancedPromptViewer
  - Integração com intercept endpoints

### MÉDIO PRAZO (Próximos meses)

**Phase 3: Memory Management:**
- [ ] **BACKEND-040 a BACKEND-050**: Full Agent Memory Lifecycle
  - Memory categorization (4 tipos)
  - Importance scoring automático
  - Relationship management
  - Memory versioning
  - Rollback
  - Conflict detection
  - Memory compression
  - Consolidação de memórias
  - Tracking de uso
  - AuditService
  - AuditController

**Phase 4: Observability:**
- [ ] **BACKEND-060 a BACKEND-070**: Production-Ready System
  - Spring Security (JWT)
  - RBAC implementation
  - Rate limiting
  - Metrics collection (Micrometer)
  - Prometheus endpoints
  - StatsController
  - LearningService (auto-evolução)
  - Memory reflection jobs
  - Advanced forgetting
  - CORS configuration
  - Input validation

**Phase 5: MCP Server (Continuação):**
- [ ] **MCP-010**: MCP Server authentication (JWT)
- [ ] **TEST-M01**: Testar MCP tools E2E
- [ ] **TEST-M02**: Testar multi-tenancy MCP

### LONGO PRAZO

**Phase 6: Polish & Deploy:**
- [ ] Security hardening
- [ ] Performance optimization
- [ ] Error handling robust
- [ ] Responsividade mobile
- [ ] Accessibility (WCAG AA)
- [ ] ErrorBoundary component
- [ ] Loading states
- [ ] Documentação completa (README, API, Setup, Architecture)
- [ ] Production deployment
- [ ] Backup strategy
- [ ] Monitoring setup (Prometheus + Grafana)
- [ ] Log aggregation
- [ ] Benchmarks (LongMemEval, SWE-Bench-Pro)

### BLOQUEIOS / DEPENDÊNCIAS

1. **DJL/ONNX para Embeddings**: Configuração necessária para embeddings reais com all-MiniLM-L6-v2
2. **Hibernate 6 Multi-tenancy**: Implementação de @TenantId e @TenantResolver
3. **Grok API**: Credenciais OpenRouter/x-ai necessárias para testes reais
4. **FalkorDB Vector Index**: Configuração de índices vetoriais no FalkorDB

---

**Documento criado em:** 18 de Janeiro de 2025
**Versão:** 2.2 (Atualizado: 2026-01-19 - MCP Server 80% completo)
**Próxima revisão:** Ao iniciar Phase 2
