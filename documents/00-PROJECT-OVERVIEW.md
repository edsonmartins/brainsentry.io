# Brain Sentry - Project Overview

**Project Name:** Brain Sentry  
**Version:** 1.0.0  
**Date:** January 2025  
**Lead Developer:** EDSON (IntegrAllTech)  

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Architecture](#system-architecture)
3. [Technology Stack](#technology-stack)
4. [Project Phases](#project-phases)
5. [Core Features](#core-features)
6. [Team & Resources](#team--resources)
7. [Success Metrics](#success-metrics)

---

## 1. Executive Summary

### 1.1 What is Brain Sentry?

Brain Sentry é um sistema inteligente de gerenciamento de contexto que funciona como um "segundo cérebro" para aplicações de IA. Ele intercepta requisições, analisa relevância, busca contexto histórico em um graph database, e injeta automaticamente informações relevantes.

**Analogia:** Sistema Límbico do cérebro humano - traz memórias relevantes automaticamente quando necessário.

### 1.2 Problem Statement

**Problema:**
- Modelos de IA (Claude Code, ChatGPT, etc) esquecem contexto de conversas anteriores
- Decisões arquiteturais e padrões não são seguidos consistentemente
- Conhecimento do projeto se perde ao longo do tempo
- Onboarding de novos desenvolvedores é lento

**Solução:**
Brain Sentry mantém memória estruturada do projeto e injeta contexto automaticamente quando relevante.

### 1.3 Target Users

**Primário:**
- Desenvolvedores usando AI coding assistants (Claude Code, Cursor, etc)
- Times de desenvolvimento que precisam manter consistência

**Secundário (VendaX.ai):**
- Agentes de IA conversando com clientes
- Sistema de vendas que precisa de contexto de cliente/histórico

### 1.4 Key Value Propositions

1. **Memória Automática** - Sistema decide o que memorizar
2. **Contexto Inteligente** - Injeta contexto relevante automaticamente
3. **Graph-based** - Relacionamentos ricos entre memórias
4. **Auditável** - Tudo rastreável e corrigível
5. **Local-First** - Dados não saem do servidor

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Web Dashboard (Next.js 15 + Radix UI)          │   │
│  │  - Memory Management UI                          │   │
│  │  - Audit & Analytics                             │   │
│  │  - System Monitoring                             │   │
│  └─────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────┘
                           │ HTTPS/REST
                           ↓
┌─────────────────────────────────────────────────────────┐
│                   API GATEWAY LAYER                      │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Spring Boot REST API                            │   │
│  │  - Authentication (JWT)                          │   │
│  │  - Rate Limiting                                 │   │
│  │  - Request Validation                            │   │
│  └─────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ↓
┌─────────────────────────────────────────────────────────┐
│                   BUSINESS LOGIC LAYER                   │
│                                                           │
│  ┌────────────────────────────────────────────────┐    │
│  │  Brain Sentry Core                             │    │
│  │  - Request Interception                         │    │
│  │  - Quick Check (regex)                          │    │
│  │  - Deep Analysis (LLM)                          │    │
│  │  - Context Injection                            │    │
│  └────────────────────────────────────────────────┘    │
│                                                           │
│  ┌────────────────────────────────────────────────┐    │
│  │  Intelligence Layer                             │    │
│  │  - LLM Integration (Qwen 2.5-7B)               │    │
│  │  - Importance Analysis                          │    │
│  │  - Pattern Detection                            │    │
│  └────────────────────────────────────────────────┘    │
│                                                           │
│  ┌────────────────────────────────────────────────┐    │
│  │  Memory Management                              │    │
│  │  - CRUD Operations                              │    │
│  │  - Search & Retrieval                           │    │
│  │  - Relationship Management                      │    │
│  └────────────────────────────────────────────────┘    │
│                                                           │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ↓
┌─────────────────────────────────────────────────────────┐
│                    DATA LAYER                            │
│                                                           │
│  ┌──────────────────┐  ┌──────────────────────────┐    │
│  │  FalkorDB        │  │  PostgreSQL               │    │
│  │  (Graph + Vector)│  │  (Audit Logs, Users)     │    │
│  │  - Memories      │  │  - Authentication         │    │
│  │  - Relationships │  │  - Analytics              │    │
│  │  - Embeddings    │  │  - Configuration          │    │
│  └──────────────────┘  └──────────────────────────┘    │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### 2.2 Component Diagram

```
Backend Components (Spring Boot):

┌─────────────────────────────────────────┐
│  @RestController Layer                  │
│  - InterceptController                  │
│  - MemoryController                     │
│  - AnalyticsController                  │
│  - AuditController                      │
└──────────────┬──────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────┐
│  @Service Layer                         │
│  - BrainSentryService                   │
│  - IntelligenceService                  │
│  - MemoryService                        │
│  - EmbeddingService                     │
│  - AuditService                         │
└──────────────┬──────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────┐
│  @Repository Layer                      │
│  - FalkorDBRepository                   │
│  - AuditLogRepository (JPA)             │
│  - UserRepository (JPA)                 │
└─────────────────────────────────────────┘
```

### 2.3 Data Flow

```
User Request Flow:

1. Client → POST /api/v1/intercept
   {
     "prompt": "Add method to OrderAgent",
     "userId": "edson",
     "context": {...}
   }

2. Controller → Service (validation)

3. BrainSentryService:
   a. Quick Check (regex patterns)
   b. If relevant → Deep Analysis (LLM)
   c. If needs context → Search FalkorDB
   d. Format & Inject context
   e. Return enhanced prompt

4. Response → Client
   {
     "enhanced": true,
     "originalPrompt": "...",
     "enhancedPrompt": "...",
     "memoriesUsed": ["mem_001", "mem_042"],
     "latencyMs": 342
   }

5. Async: Log to Audit (PostgreSQL)
```

---

## 3. Technology Stack

### 3.1 Backend Stack

```yaml
Core:
  Language: Java 17
  Framework: Spring Boot 3.2.1
  Build Tool: Maven 3.9

Databases:
  Primary: FalkorDB (Graph + Vector)
  Relational: PostgreSQL 16
  Cache: Redis (built into FalkorDB)

AI/ML:
  LLM: Qwen 2.5-7B (llama.cpp)
  Embeddings: sentence-transformers
  Integration: JNI / ProcessBuilder

Key Dependencies:
  - Spring Web
  - Spring Data JPA
  - Spring Security
  - Jedis (Redis/FalkorDB client)
  - Lombok
  - MapStruct
  - Micrometer (metrics)
  - Logback

Testing:
  - JUnit 5
  - Mockito
  - TestContainers
  - REST Assured
```

### 3.2 Frontend Stack

```yaml
Core:
  Framework: Next.js 15 (App Router)
  Language: TypeScript 5.3
  Styling: Tailwind CSS 3.4
  UI Library: Radix UI

Key Libraries:
  - React 19
  - TanStack Query (React Query)
  - Zustand (state management)
  - Zod (validation)
  - React Hook Form
  - date-fns
  - recharts (charts)
  
Development:
  - ESLint
  - Prettier
  - TypeScript strict mode
```

### 3.3 Infrastructure

```yaml
Development:
  - Docker Compose
  - FalkorDB container
  - PostgreSQL container
  - Adminer (DB UI)

Production:
  - Kubernetes
  - Helm charts
  - Prometheus + Grafana
  - ELK Stack (logs)

CI/CD:
  - GitHub Actions
  - Maven for backend
  - npm/pnpm for frontend
  - Docker build & push
```

---

## 4. Project Phases

### Phase 1: Foundation (Weeks 1-3)

**Backend:**
- ✅ Project setup (Spring Boot + Maven)
- ✅ Database setup (FalkorDB + PostgreSQL)
- ✅ Core domain models
- ✅ Repository layer (FalkorDB + JPA)
- ✅ Basic CRUD for memories
- ✅ Health check endpoints

**Frontend:**
- ✅ Project setup (Next.js + TypeScript)
- ✅ UI component library setup (Radix)
- ✅ Layout & navigation
- ✅ Authentication pages
- ✅ Memory list view (basic)

**Deliverable:** Basic system with CRUD operations working

---

### Phase 2: Core Intelligence (Weeks 4-6)

**Backend:**
- ✅ LLM integration (Qwen 2.5)
- ✅ Embedding generation
- ✅ Quick Check logic (regex patterns)
- ✅ Deep Analysis service
- ✅ Brain Sentry core algorithm
- ✅ Vector search in FalkorDB

**Frontend:**
- ✅ Intercept testing page
- ✅ Real-time feedback display
- ✅ Memory inspector (detail view)
- ✅ Search functionality

**Deliverable:** System can intercept, analyze, and inject context

---

### Phase 3: Memory Management (Weeks 7-9)

**Backend:**
- ✅ Memory categorization
- ✅ Importance scoring
- ✅ Relationship management (USED_WITH, CONFLICTS_WITH)
- ✅ Memory versioning
- ✅ Conflict detection
- ✅ Consolidation logic

**Frontend:**
- ✅ Memory editor (create/edit/delete)
- ✅ Relationship visualizer (graph view)
- ✅ Conflict resolution UI
- ✅ Version history viewer
- ✅ Bulk operations

**Deliverable:** Full memory lifecycle management

---

### Phase 4: Observability & Audit (Weeks 10-12)

**Backend:**
- ✅ Comprehensive audit logging
- ✅ Metrics collection (Micrometer)
- ✅ Performance monitoring
- ✅ Usage analytics
- ✅ Alert system

**Frontend:**
- ✅ Analytics dashboard
- ✅ Audit log viewer
- ✅ Memory usage charts
- ✅ Performance metrics
- ✅ Alert notifications

**Deliverable:** Full observability and governance

---

### Phase 5: Advanced Features (Weeks 13-15)

**Backend:**
- ✅ A/B testing framework
- ✅ Pattern detection (auto-learning)
- ✅ Memory consolidation (automated)
- ✅ Export/Import functionality
- ✅ Backup/Restore

**Frontend:**
- ✅ A/B test management UI
- ✅ Pattern discovery view
- ✅ Automated actions config
- ✅ Export/Import UI
- ✅ Settings & configuration

**Deliverable:** Production-ready with advanced features

---

### Phase 6: Polish & Deploy (Weeks 16-18)

**Backend:**
- ✅ Security hardening
- ✅ Performance optimization
- ✅ Load testing
- ✅ Documentation
- ✅ Deployment scripts

**Frontend:**
- ✅ Responsive design polish
- ✅ Accessibility (WCAG 2.1)
- ✅ Error handling
- ✅ Loading states
- ✅ User documentation

**Deliverable:** Production deployment

---

## 5. Core Features

### 5.1 MVP Features (Phase 1-3)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Memory CRUD** | Create, Read, Update, Delete memories | P0 |
| **Vector Search** | Semantic search in FalkorDB | P0 |
| **Context Injection** | Automatic context enhancement | P0 |
| **Web Dashboard** | Basic UI for management | P0 |
| **Audit Logging** | Track all actions | P1 |
| **Authentication** | User login/logout | P1 |

### 5.2 V1.0 Features (Phase 4-6)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Analytics Dashboard** | Usage metrics and insights | P1 |
| **Relationship Graph** | Visualize memory connections | P1 |
| **Memory Versioning** | Track changes over time | P1 |
| **Conflict Detection** | Find contradicting memories | P2 |
| **Pattern Learning** | Auto-detect emerging patterns | P2 |
| **A/B Testing** | Test memory effectiveness | P2 |
| **Export/Import** | Backup and restore data | P2 |

### 5.3 Future Features (V2.0+)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Multi-Tenancy** | Support multiple projects | P3 |
| **Team Collaboration** | Shared memories + permissions | P3 |
| **Real-time Sync** | WebSocket updates | P3 |
| **Mobile App** | iOS/Android client | P3 |
| **Plugin System** | Extensibility for custom logic | P3 |
| **Cloud SaaS** | Hosted version | P3 |

---

## 6. Team & Resources

### 6.1 Team Structure

```
Project Lead & Full-Stack Developer:
- EDSON (IntegrAllTech)
  • Backend development (Java/Spring Boot)
  • Architecture & design
  • FalkorDB integration
  • Frontend development (Next.js)
  • DevOps & deployment

External Resources (if needed):
- UI/UX Designer (contract)
- QA Engineer (part-time)
```

### 6.2 Hardware Requirements

**Development:**
- CPU: Intel i7 / Ryzen 7 or better
- RAM: 32GB minimum
- GPU: RTX 3060 (12GB VRAM) for LLM
- Storage: 500GB SSD

**Production (Initial):**
- Server: 16 cores, 64GB RAM
- GPU: RTX 3060 or cloud GPU instance
- Storage: 1TB SSD
- Network: 1Gbps

### 6.3 Budget Estimate

```
Development (18 weeks):
- EDSON time: In-house
- Tools & licenses: ~$500
- Cloud services (dev): ~$300/month = $1,350
- Total: ~$1,850

Production (First Year):
- Server/GPU: $15,000 (hardware) or $300/month (cloud)
- Monitoring tools: $100/month = $1,200
- Backup/Storage: $50/month = $600
- Total: ~$17,000 or ~$5,400/year (cloud)
```

---

## 7. Success Metrics

### 7.1 Technical Metrics

```yaml
Performance:
  - Latency p95: < 500ms
  - Latency p99: < 1000ms
  - Throughput: > 100 req/sec
  - Uptime: > 99.5%

Quality:
  - Context relevance: > 85%
  - False positive rate: < 15%
  - Test coverage: > 80%
  - Code quality: SonarQube A rating

Scalability:
  - Memories: Support 100k+
  - Concurrent users: 50+
  - Database size: < 10GB for 100k memories
```

### 7.2 Business Metrics

```yaml
Adoption:
  - Daily active users: 10+ (VendaX team)
  - Memories created: 1000+ in first 3 months
  - Queries per day: 500+

Effectiveness:
  - Context injection accuracy: > 85%
  - User satisfaction: > 4.0/5.0
  - Time saved per developer: > 2 hours/week

Growth (if product):
  - Beta users: 50 in 6 months
  - Paying customers: 10 in 12 months
  - MRR: $2,000+ in 12 months
```

### 7.3 KPIs Dashboard

```
Weekly Review:
- Number of memories added
- Average query latency
- Context injection success rate
- User feedback score
- System uptime

Monthly Review:
- Feature completion rate
- Bug count and resolution time
- Memory database growth
- Infrastructure costs
- User retention
```

---

## 8. Risk Management

### 8.1 Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| FalkorDB performance issues | High | Medium | Benchmark early, plan fallback to Qdrant |
| LLM inference too slow | High | Medium | Optimize batch processing, consider GPU upgrade |
| Graph queries complex | Medium | High | Study Cypher deeply, prototype early |
| Memory growth unbounded | Medium | Medium | Implement archival, set retention policies |

### 8.2 Project Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Scope creep | High | High | Strict phase gates, prioritize ruthlessly |
| Single developer bottleneck | High | Medium | Document everything, modular architecture |
| Timeline delays | Medium | Medium | Buffer time in each phase, MVP first |
| Technology learning curve | Medium | Low | Allocate learning time, use familiar tools where possible |

### 8.3 Contingency Plans

```
Plan A (Ideal):
- Java + Spring Boot + FalkorDB
- 18 weeks to V1.0

Plan B (If FalkorDB Issues):
- Switch to Qdrant for vector
- Keep PostgreSQL for graph (manual joins)
- Add 2 weeks to timeline

Plan C (If LLM Too Slow):
- Simplify to heuristic-based analysis
- Use LLM only for critical decisions
- Maintain functionality, reduce intelligence

Plan D (If Timeline Pressure):
- Cut A/B testing (Phase 5)
- Cut analytics dashboard (Phase 4)
- Focus on core MVP (Phase 1-3)
- Ship V0.5 instead of V1.0
```

---

## 9. Documentation Plan

### 9.1 Technical Documentation

```
Architecture:
- System design document (this)
- API documentation (OpenAPI/Swagger)
- Database schema
- Deployment guide

Development:
- Backend specification
- Frontend specification
- API contracts
- Coding standards
- Git workflow
```

### 9.2 User Documentation

```
User Guides:
- Getting started
- Memory management
- Dashboard usage
- Best practices
- Troubleshooting

API Documentation:
- REST API reference
- Integration guide
- Example requests
- Error codes
```

### 9.3 Operational Documentation

```
Operations:
- Deployment procedures
- Monitoring setup
- Backup/Restore
- Incident response
- Performance tuning
```

---

## 10. Next Steps

### 10.1 Immediate Actions (Week 1)

1. **Setup Development Environment**
   - Install Java 17, Maven, Docker
   - Setup IDE (IntelliJ IDEA)
   - Clone repositories

2. **Create Project Structure**
   - Initialize Spring Boot project
   - Initialize Next.js project
   - Setup Git repository

3. **Infrastructure Setup**
   - Docker Compose file
   - FalkorDB container
   - PostgreSQL container

4. **Documentation**
   - Review all specification documents
   - Set up project management (Jira/Trello)
   - Create task breakdown

### 10.2 Week 1 Deliverables

- ✅ All documentation complete
- ✅ Development environment ready
- ✅ Backend "Hello World" running
- ✅ Frontend "Hello World" running
- ✅ Docker containers running
- ✅ First commit to Git

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| **Memory** | A piece of stored knowledge (pattern, decision, fact) |
| **Brain Sentry** | The core intelligence component that manages context |
| **Quick Check** | Fast regex-based relevance detection |
| **Deep Analysis** | LLM-powered detailed analysis |
| **Context Injection** | Adding relevant memories to a prompt |
| **FalkorDB** | Graph + Vector database (Redis-based) |
| **GraphRAG** | Retrieval Augmented Generation using graph queries |

---

## Appendix B: Reference Links

**Technologies:**
- Spring Boot: https://spring.io/projects/spring-boot
- FalkorDB: https://www.falkordb.com/
- Next.js: https://nextjs.org/
- Radix UI: https://www.radix-ui.com/

**Related Projects:**
- Cursor: https://cursor.sh/
- Continue.dev: https://continue.dev/
- Mem0: https://mem0.ai/

---

**Document Version:** 1.0  
**Last Updated:** January 17, 2025  
**Next Review:** Start of Phase 2  

---

**Status:** ✅ Ready for Development

This overview serves as the master reference for the Brain Sentry project. All other documents (Backend, Frontend, API, Deployment) extend from this foundation.
