# Brain Sentry - Project Overview

**Version:** 1.0  
**Date:** Janeiro 2025  
**Project Owner:** EDSON - IntegrAllTech/VendaX.ai  
**Status:** Planning Phase  

---

## Executive Summary

Brain Sentry é um sistema inteligente de gerenciamento de contexto para assistentes de IA (LLMs), projetado para resolver o problema fundamental de **memória e consistência** em desenvolvimento de software e automação de vendas.

O sistema funciona como um "segundo cérebro" que:
- Intercepta requisições para LLMs
- Decide autonomamente quando contexto histórico é relevante
- Injeta automaticamente memórias e padrões no prompt
- Aprende e evolui continuamente com o uso

---

## Problem Statement

### Contexto
Ao usar assistentes de IA (Claude Code, Cursor, etc.) para desenvolvimento ou vendas, existe um padrão recorrente:

**Segunda-feira:** Decisão arquitetural - "Usar Spring Events para comunicação entre agentes"  
**Terça-feira:** Implementação seguindo o padrão  
**Quinta-feira:** Nova feature - IA **esquece** o padrão e sugere chamadas REST diretas  
**Sexta-feira:** Refactoring manual para corrigir  

### Problema Fundamental
- ❌ **Models têm contexto limitado** e "esquecem" decisões passadas
- ❌ **Documentação existe mas é ignorada** (claude.md, README, etc)
- ❌ **Inconsistência crescente** à medida que o projeto evolui
- ❌ **Knowledge é perdido** quando desenvolvedores/vendedores saem

### Soluções Atuais (Insuficientes)
- Documentação estática (esquecida)
- RAG básico (sem contexto relacional)
- MCP tools (modelo precisa lembrar de chamar)

---

## Solution: Brain Sentry

### Conceito Central

```
┌─────────────────────────────────────────────┐
│  USUÁRIO                                     │
│  "Adicione método no OrderAgent"           │
└──────────────┬──────────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────────┐
│  BRAIN SENTRY (Autonomous)                  │
│  1. Intercepta TODA requisição              │
│  2. Decide: precisa de contexto?            │
│  3. Busca memórias relevantes (GraphRAG)    │
│  4. Injeta automaticamente                  │
└──────────────┬──────────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────────┐
│  LLM (Claude/GPT/etc)                       │
│  Recebe prompt JÁ ENRIQUECIDO               │
│  Não precisa "lembrar" de buscar            │
└─────────────────────────────────────────────┘
```

**Diferencial:** O modelo de trabalho **não decide** quando buscar contexto - o Brain Sentry **sempre analisa** e injeta quando relevante.

---

## Core Features

### 1. Autonomous Context Injection
- Interceptação transparente de requisições
- Análise inteligente de relevância (quick check + deep analysis)
- Busca semântica + grafo relacional (GraphRAG)
- Injeção automática no prompt

### 2. Intelligent Memory Management
- Memórias com embeddings vetoriais
- Relacionamentos nativos (graph database)
- Categorização automática (decision, pattern, antipattern, domain)
- Importância dinâmica (evolui com uso)

### 3. Full Auditability
- Audit log de todas as decisões
- Proveniência completa de memórias
- Version history com rollback
- Impact analysis antes de mudanças

### 4. Observability Dashboard
- Visualização de memórias e relacionamentos
- Métricas de efetividade
- Alertas automáticos
- Memory inspector detalhado

### 5. Continuous Learning
- Captura automática de novas memórias
- Detecção de patterns emergentes
- Promoção/rebaixamento de importância
- Consolidação e deduplicação

---

## Technology Stack

### Backend

```yaml
Language: Java 17
Framework: Spring Boot 3.2+
Build Tool: Maven ou Gradle

Core Dependencies:
  - Spring Web (REST API)
  - Spring Data Redis (FalkorDB integration)
  - Spring Boot Actuator (health, metrics)
  - Jedis (Redis client)
  
LLM Integration:
  - LlamaJava (llama.cpp Java bindings)
  - DJL (Deep Java Library) - embeddings
  - ONNX Runtime - alternativamente
  
Memory Store:
  - FalkorDB (Redis Graph + Vector search)
  - Redis (base)
  
Utilities:
  - Lombok (reduce boilerplate)
  - MapStruct (DTO mapping)
  - Jackson (JSON)
  - SLF4J + Logback (logging)
  
Testing:
  - JUnit 5
  - Mockito
  - TestContainers (integration tests)
```

### Frontend

```yaml
Framework: Next.js 15
Language: TypeScript
UI Library: Radix UI
Styling: Tailwind CSS

Core Dependencies:
  - React 18
  - Radix UI primitives
  - TanStack Query (data fetching)
  - Zustand (state management)
  - React Hook Form (forms)
  - Zod (validation)
  
Visualization:
  - Recharts (charts)
  - React Flow (graph visualization)
  - Lucide React (icons)
  
Development:
  - ESLint
  - Prettier
  - TypeScript strict mode
```

### Infrastructure

```yaml
Development:
  - Docker Compose
  - FalkorDB container
  - Hot reload (both backend/frontend)

Production:
  - Kubernetes (optional)
  - Redis Cluster
  - Load balancer
  - Monitoring (Prometheus + Grafana)
```

---

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Web UI     │  │  Claude Code │  │  VendaX.ai   │ │
│  │  (Next.js)   │  │   (Proxy)    │  │   Agents     │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
└─────────┼──────────────────┼──────────────────┼─────────┘
          │                  │                  │
          └──────────────────┴──────────────────┘
                             │
                     REST API (HTTP/JSON)
                             │
┌────────────────────────────┴─────────────────────────────┐
│                   APPLICATION LAYER                       │
│  ┌─────────────────────────────────────────────────────┐│
│  │         Brain Sentry Service (Spring Boot)          ││
│  │                                                       ││
│  │  ┌───────────────┐  ┌──────────────┐  ┌──────────┐ ││
│  │  │ Interception  │  │   Memory     │  │   API    │ ││
│  │  │    Engine     │  │ Intelligence │  │ Gateway  │ ││
│  │  └───────┬───────┘  └──────┬───────┘  └────┬─────┘ ││
│  │          │                  │                │       ││
│  │  ┌───────┴──────────────────┴────────────────┴────┐ ││
│  │  │           Core Business Logic                   │ ││
│  │  │  • Context Analysis                             │ ││
│  │  │  • Memory Management                            │ ││
│  │  │  • Learning & Evolution                         │ ││
│  │  └─────────────────────────────────────────────────┘ ││
│  └─────────────────────────────────────────────────────┘│
└────────────────────────────┬─────────────────────────────┘
                             │
┌────────────────────────────┴─────────────────────────────┐
│                    DATA LAYER                             │
│  ┌──────────────────┐  ┌────────────────┐  ┌──────────┐│
│  │    FalkorDB      │  │  LLM Service   │  │  Cache   ││
│  │  (Graph + Vec)   │  │  (Qwen 7B)     │  │  (Redis) ││
│  └──────────────────┘  └────────────────┘  └──────────┘│
└───────────────────────────────────────────────────────────┘
```

### Component Interaction Flow

```
User Request
    ↓
API Gateway
    ↓
Interception Engine
    ├─→ Quick Check (fast) ─→ Pass through?
    │                              ↓ No
    └─→ Deep Analysis (LLM) ─→ Needs context?
                                   ↓ Yes
Memory Intelligence
    ├─→ Generate Embedding
    ├─→ Search FalkorDB (GraphRAG)
    ├─→ Rank & Filter
    └─→ Format Context
         ↓
Context Injection
    ↓
Enhanced Prompt
    ↓
Return to Client
```

---

## Use Cases

### Primary: Software Development (VendaX.ai codebase)

**Scenario:** Developer working on VendaX.ai agents

```
Developer: "Add validation method to OrderAgent"

Brain Sentry:
1. Detects "OrderAgent" + "method"
2. Searches: patterns related to OrderAgent
3. Finds:
   - Critical: "Agents must validate with BeanValidator"
   - Important: "Use Spring Events for communication"
   - Important: "Error handling with BusinessException"
4. Injects context into prompt
5. LLM generates code ALREADY FOLLOWING patterns

Result: Consistent code from day 1
```

### Secondary: Sales Automation (VendaX.ai client interactions)

**Scenario:** AI agent interacting with customer

```
Customer: "Quero fazer um pedido" (WhatsApp: +55...)

Brain Sentry:
1. Identifies customer by phone (Bella Pasta)
2. Searches: customer history, preferences, negotiations
3. Finds:
   - Customer profile: Buys Penne Rigate bi-weekly
   - Discount: 5% current
   - Last order: 15 days ago (should be ordering now)
   - Preference: Fast delivery
4. Injects context
5. Agent responds: "Olá Marco! O Penne Rigate como sempre?"

Result: Personalized, contextual conversation
```

---

## Success Metrics

### Technical Metrics
- **Latency p95:** < 500ms (end-to-end)
- **Context Injection Rate:** 30-40% of requests
- **Helpfulness Rate:** > 85% (based on feedback)
- **Memory Growth:** Sustainable (< 10k memories/month)
- **System Uptime:** 99.9%

### Business Metrics
- **Developer Onboarding:** 3x faster (3 weeks → 1 week)
- **Code Consistency:** 90%+ following patterns
- **Knowledge Retention:** 100% (nothing lost when dev leaves)
- **Sales Conversion:** +15% (better personalization)
- **Ticket Médio:** +20% (intelligent cross-sell)

---

## Target Users

### Primary: Development Teams
- Small to medium teams (5-50 developers)
- Using AI assistants (Claude Code, Cursor, Copilot)
- Codebases with > 6 months history
- Need consistency and knowledge retention

### Secondary: Sales Teams
- B2B sales with relationship focus
- High-value, recurring customers
- Need personalization at scale
- Using AI for customer interaction

---

## Competitive Landscape

### Direct Competitors
- **None identified** - No product exactly like Brain Sentry

### Adjacent Solutions
- **Cursor/Claude Code:** IDE integration, but no memory system
- **LangChain Memory:** Basic, not autonomous
- **Pinecone/Weaviate:** Vector stores, but not graph-aware
- **Notion AI, etc:** Document-based, not code-aware

### Differentiation
✅ **Autonomous decision** (model doesn't need to remember)  
✅ **Graph + Vector** (relationships matter)  
✅ **Continuous learning** (gets better with use)  
✅ **Full auditability** (enterprise-ready)  
✅ **Local-first** (data sovereignty)  

---

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| LLM accuracy insufficient | High | Medium | Extensive testing, fallback to heuristics |
| FalkorDB performance issues | Medium | Low | Benchmark early, consider Qdrant as backup |
| Adoption resistance | High | Medium | Transparent UX, clear value demonstration |
| Memory pollution | Medium | Medium | Smart filtering, human review for critical |
| Infrastructure costs | Low | Low | Local-first, cloud optional |

---

## Development Timeline

### Phase 1: MVP (8 weeks)
Core functionality:
- Context interception
- Basic memory management
- Simple dashboard

### Phase 2: Intelligence (4 weeks)
Advanced features:
- LLM-based analysis
- GraphRAG
- Learning system

### Phase 3: Production (4 weeks)
Enterprise features:
- Full auditability
- Advanced dashboard
- Performance optimization

### Phase 4: Scale (ongoing)
Scaling:
- Multi-tenancy
- High availability
- Advanced analytics

**Total to V1 Production:** ~16 weeks (4 months)

---

## Next Steps

1. ✅ Detailed phase planning → `DEVELOPMENT_PHASES.md`
2. ✅ Backend specification → `BACKEND_SPECIFICATION.md`
3. ✅ Frontend specification → `FRONTEND_SPECIFICATION.md`
4. ✅ Development setup → `SETUP_GUIDE.md`
5. ⏭️ Begin Phase 1 implementation

---

## Team & Resources

### Required Roles
- **Tech Lead / Architect:** EDSON
- **Backend Developer:** 1 (Java/Spring Boot)
- **Frontend Developer:** 1 (Next.js/React)
- **DevOps:** Part-time (Docker/K8s setup)

### Infrastructure
- **Development:**
  - Local machines (adequate)
  - Docker for FalkorDB
  - Git repository
  
- **Production:**
  - GPU server (for LLM - RTX 3060 sufficient)
  - Redis server (FalkorDB)
  - Application server (4 cores, 8GB RAM)

---

## Appendix

### Glossary
- **Brain Sentry:** Sistema principal de gerenciamento de contexto
- **Memory:** Unidade de conhecimento armazenada (decisão, pattern, etc)
- **Embedding:** Representação vetorial de texto para busca semântica
- **GraphRAG:** Retrieval Augmented Generation com graph database
- **Context Injection:** Processo de adicionar contexto relevante ao prompt

### References
- Project conceptual documentation: `project-brain-sentry-concept.md`
- FalkorDB documentation: https://docs.falkordb.com
- Spring Boot reference: https://spring.io/projects/spring-boot
- Next.js documentation: https://nextjs.org/docs

---

**Document Status:** ✅ Complete  
**Last Updated:** Janeiro 2025  
**Next Review:** Após Phase 1 completion
