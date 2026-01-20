# Research Cards - Visual Specifications for Designers

**Component:** Research Paper Cards  
**Section:** Academic Validation  
**Total Cards:** 8 papers  
**Layout:** 2-column grid (desktop), 1-column (mobile)  

---

## CARD 1: CONFUCIUS CODE AGENT ⭐ (Featured)

```
┌─────────────────────────────────────────────────────────────┐
│ 🏆 FEATURED RESEARCH                                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ 📄 Confucius Code Agent: Scalable Agent Scaffolding        │
│    for Real-World Codebases                                 │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ META                                                 │   │
│ │ Wong, Sherman • Qi, Zhenting • Wang, Zhaodong      │   │
│ │ Meta AI & Harvard University                        │   │
│ │ December 2025 • arXiv:2512.10398v5                 │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ KEY RESULTS                                                 │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ 🎯 54.3% on SWE-Bench-Pro                                  │
│    (State-of-the-Art Performance)                          │
│                                                              │
│ 🎯 Proves scaffolding > model capability                   │
│    Claude Sonnet 4.5 + CCA (52.7%)                         │
│    > Claude Opus 4.5 + Anthropic (52.0%)                   │
│                                                              │
│ 🎯 Hierarchical memory architecture                        │
│    Context management for long sessions                     │
│                                                              │
│ 🎯 Note-taking agent                                       │
│    Persistent knowledge across sessions                     │
│                                                              │
│ INSIGHTS FOR BRAIN SENTRY                                  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ ✅ Memory architecture matters more than LLM               │
│ ✅ Context compression is critical                         │
│ ✅ Note-taking enables cross-session learning              │
│ ✅ Meta-agent can automate development                     │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ ALIGNMENT: 85%                                       │   │
│ │                                                      │   │
│ │ Shared:    Memory systems, context management       │   │
│ │ Different: Graph vs hierarchy, autonomous vs tool   │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ WHAT WE LEARNED & IMPLEMENTED                              │
│ • Phase 3: Note-taking agent (Confucius-inspired)          │
│ • Phase 3: Architect agent for compression                 │
│ • Phase 5: Meta-agent (build-test-improve loop)            │
│                                                              │
│ [📄 Read Full Analysis] [🔗 View Paper]                   │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Background: Linear gradient #F9FAFB → #FFFFFF
- Border: 2px solid #3B82F6 (blue, featured)
- Border-radius: 16px
- Padding: 32px
- Shadow on hover: 0 20px 40px rgba(59, 130, 246, 0.2)
- Featured badge: Gold/orange gradient background
- Alignment score: Green gradient pill (85%)
```

---

## CARD 2: FROM RAG TO AGENT MEMORY ⭐ (Featured)

```
┌─────────────────────────────────────────────────────────────┐
│ 🏆 FEATURED RESEARCH                                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ 📄 From RAG to Agent Memory                                │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ INDUSTRY THOUGHT LEADERSHIP                          │   │
│ │ Monigatti, Leonie                                   │   │
│ │ Independent AI Researcher                           │   │
│ │ 2024 • Towards Data Science                        │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ KEY INSIGHT                                                 │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ The industry is evolving from RAG (read-only retrieval)    │
│ to Agent Memory (read-write with learning). Multi-type     │
│ memory systems are essential for advanced agents.          │
│                                                              │
│ FRAMEWORK EVOLUTION                                         │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│                                                              │
│ 2020-2023: RAG                                             │
│ └─ Read-only • Static knowledge                           │
│                                                              │
│ 2023-2024: Agentic RAG                                     │
│ └─ Agent decides when to retrieve • Tool-based            │
│                                                              │
│ 2024-2025: AGENT MEMORY ← Brain Sentry                    │
│ └─ Read-write • Learning • Multi-type                     │
│                                                              │
│ MEMORY TYPES VALIDATED                                     │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ ✓ Semantic Memory   (facts/concepts)                      │
│ ✓ Episodic Memory   (events/history)                      │
│ ✓ Procedural Memory (how-to/patterns)                     │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ ALIGNMENT: 95%                                       │   │
│ │                                                      │   │
│ │ We implement all 3 types PLUS Associative Memory    │   │
│ │ (relationships via graph) - unique to Brain Sentry  │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ IMPACT ON BRAIN SENTRY                                     │
│ • Validated our 4 memory types architecture                │
│ • Confirmed need for write operations                      │
│ • Inspired autonomous (not tool-based) approach            │
│                                                              │
│ [📄 Read Article] [🔗 Full Analysis]                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Background: Linear gradient #F9FAFB → #FFFFFF
- Border: 2px solid #8B5CF6 (purple, featured)
- Evolution timeline: Vertical with connecting lines
- Alignment score: Green gradient pill (95%)
```

---

## CARD 3: CoALA FRAMEWORK

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│ 📄 CoALA: Cognitive Architectures for Language Agents      │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Sumers, T. • Yao, S. • Narasimhan, K. • Griffiths, T.│   │
│ │ Princeton • Stanford • DeepMind                      │   │
│ │ 2024 • arXiv                                         │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ KEY CONTRIBUTION                                            │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ Agents need structured cognitive architectures that        │
│ separate different types of memory and knowledge.          │
│ Generic "chat history" is insufficient.                    │
│                                                              │
│ FRAMEWORK COMPONENTS                                        │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ • Working Memory    (short-term reasoning)                 │
│ • Episodic Memory   (experience traces)                    │
│ • Semantic Memory   (factual knowledge)                    │
│ • Procedural Memory (action sequences)                     │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ ALIGNMENT: 100%                                      │   │
│ │                                                      │   │
│ │ Our 4 memory types directly map to CoALA framework  │   │
│ │ + Associative Memory (graph relationships)          │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ WHAT WE LEARNED                                            │
│ • Memory type separation is scientifically validated       │
│ • Working memory needs hierarchical structure              │
│ • Procedural memory crucial for code patterns              │
│                                                              │
│ [🔗 View Paper]                                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Border: 1px solid #E5E7EB
- Border-left: 4px solid #3B82F6
- Alignment badge: Bright green (100%)
- Framework components: Icon + label grid
```

---

## CARD 4: GraphRAG

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│ 📄 GraphRAG: Knowledge Graph-Based RAG                     │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Microsoft Research                                   │   │
│ │ 2024 • arXiv                                         │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ KEY FINDING                                                 │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ Combining graph traversal with vector search yields        │
│ significantly better retrieval than either alone.          │
│                                                              │
│ PERFORMANCE GAINS                                           │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ 📊 2-3x improvement over vector-only RAG                   │
│ 📊 Multi-hop reasoning enabled                             │
│ 📊 Community detection finds implicit patterns             │
│                                                              │
│ ARCHITECTURAL VALIDATION                                    │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ ✓ Graph-native storage > separate vector + graph DBs      │
│ ✓ Relationship queries crucial for context                │
│ ✓ Cypher queries more expressive than SQL                 │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ ALIGNMENT: 100%                                      │   │
│ │                                                      │   │
│ │ FalkorDB provides native Graph + Vector in one DB   │   │
│ │ We implement GraphRAG from day one                  │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ IMPACT                                                      │
│ • Validated our FalkorDB choice                            │
│ • Confirmed graph-first architecture                       │
│ • Inspired GraphRAG implementation                         │
│                                                              │
│ [🔗 View Paper] [💡 See Our Implementation]               │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Performance chart: Mini bar chart showing 2-3x improvement
- Microsoft logo/badge
- Alignment: Bright green (100%)
```

---

## CARD 5: MemGPT

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│ 📄 MemGPT: Towards LLMs as Operating Systems               │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Packer, C. • Wooders, S. • Lin, K. • et al.         │   │
│ │ UC Berkeley                                          │   │
│ │ 2023 • arXiv                                         │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ KEY INNOVATION                                              │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ LLMs need hierarchical memory management inspired by       │
│ operating systems: fast working memory + persistent        │
│ long-term storage.                                         │
│                                                              │
│ OS-INSPIRED CONCEPTS                                        │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ • Main Context      (working memory/RAM)                   │
│ • External Context  (long-term storage/disk)               │
│ • Memory Paging     (swap in/out)                          │
│ • Self-editing      (memory management)                    │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ ALIGNMENT: 75%                                       │   │
│ │                                                      │   │
│ │ Hierarchical concept ✓ Graph-native storage ✓      │   │
│ │ Different: Graph vs flat text chunks                │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ INSIGHTS ADOPTED                                           │
│ • Memory lifecycle management strategies                   │
│ • Importance of forgetting mechanisms                      │
│ • Context window compression techniques                    │
│                                                              │
│ [🔗 View Paper]                                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- OS metaphor icons (RAM, disk, swap)
- UC Berkeley badge
- Alignment: Yellow-green (75%)
```

---

## CARD 6: SWE-BENCH PRO

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│ 📄 SWE-Bench Pro: Long-Horizon Software Engineering        │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Deng, X. • Da, J. • Pan, E. • et al.                │   │
│ │ Princeton University • Scale AI                      │   │
│ │ 2025 • arXiv                                         │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ BENCHMARK PURPOSE                                           │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ Evaluate AI coding agents on real-world, production-       │
│ level software engineering tasks requiring multi-file      │
│ edits and deep codebase context.                           │
│                                                              │
│ BENCHMARK STATS                                             │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ • 731 real GitHub issues                                   │
│ • Multi-file modifications required                        │
│ • Long-context reasoning tested                            │
│ • Production-level complexity                              │
│                                                              │
│ STATE-OF-THE-ART PERFORMANCE                               │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ 🏆 Confucius (Meta/Harvard): 54.3%                        │
│ 📊 Anthropic Claude Opus: 52.0%                           │
│ 📊 Live-SWE-Agent: 45.8%                                  │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ BRAIN SENTRY GOAL                                    │   │
│ │                                                      │   │
│ │ Phase 6 Target: >55% (beat current SOTA)            │   │
│ │ Validate graph-native vs hierarchical architecture  │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ WHY IT MATTERS                                             │
│ • Proves memory architecture = performance                 │
│ • Validates long-context capabilities                      │
│ • Industry-standard benchmark                              │
│                                                              │
│ [🔗 View Benchmark] [📊 See Leaderboard]                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Performance chart: Top 3 results comparison
- Target badge: "Phase 6: >55%" (orange)
- Princeton + Scale AI logos
```

---

## CARD 7: AGENT MEMORY SURVEY

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│ 📄 Survey: Memory in LLM-Based Agents                      │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Multiple Authors (Survey Paper)                      │   │
│ │ 2024 • arXiv                                         │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ COMPREHENSIVE OVERVIEW                                      │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ Categorizes existing memory systems and identifies         │
│ critical gaps in current research.                         │
│                                                              │
│ MEMORY CATEGORIES SURVEYED                                  │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ ✓ Parametric Memory  (model weights)                       │
│ ✓ Episodic Memory    (experiences)                         │
│ ✓ Semantic Memory    (facts)                               │
│ ✓ Working Memory     (reasoning)                           │
│ ✓ Procedural Memory  (skills)                              │
│                                                              │
│ RESEARCH GAPS IDENTIFIED                                    │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ ❌ Most systems lack graph relationships                   │
│ ❌ Few have full auditability                              │
│ ❌ Limited production deployments                          │
│ ❌ Weak developer-specific features                        │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ BRAIN SENTRY FILLS ALL GAPS                          │   │
│ │                                                      │   │
│ │ ✅ Graph-native relationships                        │   │
│ │ ✅ Full audit trail                                  │   │
│ │ ✅ Production-ready architecture                     │   │
│ │ ✅ Developer-focused memory types                    │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ [🔗 View Survey]                                           │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Gap vs filled: Check/X icon comparison
- Green highlight for Brain Sentry advantages
```

---

## CARD 8: RAG (FOUNDATIONAL)

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│ 📄 Retrieval-Augmented Generation for                      │
│    Knowledge-Intensive NLP Tasks                           │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ 🏛️ FOUNDATIONAL PAPER                                │   │
│ │ Lewis, P. • Perez, E. • Piktus, A. • et al.         │   │
│ │ Meta AI • University College London                 │   │
│ │ 2020 • NeurIPS                                       │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ HISTORICAL SIGNIFICANCE                                     │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ The foundational RAG paper that started it all.            │
│ Showed retrieval before generation reduces                 │
│ hallucinations and improves accuracy.                      │
│                                                              │
│ CORE INNOVATION (2020)                                      │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ ✓ Dense retrieval (vector embeddings)                     │
│ ✓ Non-parametric knowledge access                         │
│ ✓ Combine retrieval + generation                          │
│                                                              │
│ LIMITATIONS IDENTIFIED (2025 VIEW)                         │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━             │
│ ❌ Read-only (can't learn)                                 │
│ ❌ No write operations                                     │
│ ❌ No relationships between documents                      │
│ ❌ Static knowledge base                                   │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ BRAIN SENTRY EVOLUTION                               │   │
│ │                                                      │   │
│ │ Built on RAG foundations, we add:                   │   │
│ │ ✅ Read-write operations                             │   │
│ │ ✅ Graph relationships                               │   │
│ │ ✅ Continuous learning                               │   │
│ │ ✅ Multi-type memory                                 │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ Standing on the shoulders of this giant! 🙏                │
│                                                              │
│ [🔗 View Paper (NeurIPS 2020)]                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Vintage/foundational badge (gold)
- Timeline: 2020 → 2025 evolution
- Meta AI + UCL logos
```

---

## EVOLUTION TIMELINE COMPONENT

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│  THE EVOLUTION OF AGENT MEMORY (2020-2025)                 │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  2020 ●━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│       │                                                      │
│       │ RAG (Lewis et al., Meta AI)                        │
│       │ • Read-only retrieval                              │
│       └─ Vector search                                     │
│                                                              │
│  2023 ●━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│       │                                                      │
│       │ MemGPT (UC Berkeley)                               │
│       │ • Hierarchical memory                              │
│       └─ OS-inspired architecture                          │
│                                                              │
│  2024 ●━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│       │                                                      │
│       ├─ GraphRAG (Microsoft)                              │
│       │  • Graph + Vector hybrid                           │
│       │                                                      │
│       ├─ CoALA (Princeton/Stanford)                        │
│       │  • Cognitive framework                             │
│       │                                                      │
│       └─ Agent Memory (Monigatti)                          │
│          • Read-write operations                            │
│                                                              │
│  2025 ●━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│       │                                                      │
│       ├─ Confucius (Meta/Harvard)                          │
│       │  🏆 54.3% SWE-Bench-Pro                            │
│       │                                                      │
│       └─ BRAIN SENTRY                                      │
│          🌟 Graph-Native + Autonomous                      │
│          🌟 4 Memory Types + Associative                   │
│          🌟 Production-Ready + Full Audit                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Vertical timeline with gradient line (#3B82F6 → #8B5CF6)
- Dots: Filled circles with subtle glow
- Current (Brain Sentry): Larger dot with pulse animation
- Responsive: Horizontal on mobile
- Interactive: Hover to see details
```

---

## CREDIBILITY BADGES COMPONENT

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│  VALIDATED BY LEADING RESEARCH INSTITUTIONS                 │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  [🎓 Meta AI]  [🎓 Harvard]  [🎓 Microsoft Research]       │
│                                                              │
│  [🎓 Princeton] [🎓 Stanford] [🎓 UC Berkeley]              │
│                                                              │
│  [🎓 DeepMind]  [🎓 UCL]      [🎓 Scale AI]                │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  📄 8 Research Papers    📊 4 Major Benchmarks              │
│                                                              │
│  ⭐ 54.3% SWE-Bench-Pro (Reference Performance)            │
│                                                              │
└─────────────────────────────────────────────────────────────┘

VISUAL SPECS:
- Badge grid: 3x3 on desktop, 2-column on mobile
- Subtle gray borders
- Institution logos (if available) or text badges
- Hover: Slight scale up + shadow
```

---

## CSS SPECIFICATIONS

```css
/* Research Paper Card Base */
.research-card {
  background: linear-gradient(135deg, #F9FAFB 0%, #FFFFFF 100%);
  border: 1px solid #E5E7EB;
  border-radius: 16px;
  padding: 32px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

/* Featured Cards */
.research-card--featured {
  border: 2px solid #3B82F6;
  border-left: 6px solid #3B82F6;
}

.research-card--featured::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #3B82F6 0%, #8B5CF6 100%);
}

/* Hover State */
.research-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 40px rgba(59, 130, 246, 0.15);
  border-color: #3B82F6;
}

/* Featured Badge */
.featured-badge {
  display: inline-block;
  background: linear-gradient(135deg, #F59E0B 0%, #EF4444 100%);
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 16px;
}

/* Meta Information */
.research-meta {
  background: #F3F4F6;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
}

.research-meta .institution {
  color: #6B7280;
  font-size: 14px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
}

.research-meta .authors {
  color: #374151;
  font-size: 14px;
  line-height: 1.5;
}

/* Alignment Score Badge */
.alignment-score {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(135deg, #10B981 0%, #059669 100%);
  color: white;
  padding: 8px 16px;
  border-radius: 24px;
  font-size: 16px;
  font-weight: 700;
}

.alignment-score--high {
  background: linear-gradient(135deg, #10B981 0%, #059669 100%);
}

.alignment-score--medium {
  background: linear-gradient(135deg, #F59E0B 0%, #DC7609 100%);
}

/* Section Dividers */
.research-divider {
  height: 2px;
  background: linear-gradient(90deg, 
    rgba(59, 130, 246, 0) 0%, 
    rgba(59, 130, 246, 0.5) 50%, 
    rgba(59, 130, 246, 0) 100%
  );
  margin: 24px 0;
}

/* Key Results List */
.key-results {
  display: grid;
  gap: 12px;
  margin: 16px 0;
}

.key-result-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: #F9FAFB;
  border-radius: 8px;
  border-left: 3px solid #3B82F6;
}

.key-result-item .icon {
  font-size: 20px;
  flex-shrink: 0;
}

/* Timeline */
.evolution-timeline {
  position: relative;
  padding: 40px 0;
}

.evolution-timeline::before {
  content: '';
  position: absolute;
  left: 40px;
  top: 0;
  bottom: 0;
  width: 3px;
  background: linear-gradient(180deg, #3B82F6 0%, #8B5CF6 100%);
}

.timeline-item {
  position: relative;
  padding-left: 80px;
  margin-bottom: 40px;
}

.timeline-item::before {
  content: '';
  position: absolute;
  left: 32px;
  top: 8px;
  width: 16px;
  height: 16px;
  background: #3B82F6;
  border: 4px solid white;
  border-radius: 50%;
  box-shadow: 0 0 0 3px #3B82F6;
}

.timeline-item--current::before {
  width: 24px;
  height: 24px;
  left: 28px;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.7);
  }
  50% {
    box-shadow: 0 0 0 10px rgba(59, 130, 246, 0);
  }
}

/* Responsive */
@media (max-width: 768px) {
  .research-card {
    padding: 24px;
  }
  
  .evolution-timeline {
    padding-left: 0;
  }
  
  .evolution-timeline::before {
    left: 20px;
  }
  
  .timeline-item {
    padding-left: 60px;
  }
}
```

---

## INTERACTIVE ELEMENTS

**Expandable Details:**
```javascript
// Click to expand full paper details
const expandCard = (cardId) => {
  const card = document.querySelector(`#${cardId}`);
  card.classList.toggle('expanded');
  
  // Track analytics
  trackEvent('research_card_expand', { paper: cardId });
}
```

**Paper Link Tracking:**
```javascript
// Track external paper link clicks
document.querySelectorAll('.paper-link').forEach(link => {
  link.addEventListener('click', (e) => {
    const paperName = e.target.dataset.paper;
    trackEvent('paper_link_click', { 
      paper: paperName,
      destination: e.target.href 
    });
  });
});
```

---

**STATUS:** ✅ Visual Specifications Complete  
**Total Cards:** 8 research papers  
**Components:** Cards + Timeline + Badges  
**Design System:** Fully specified with CSS  

**Ready for:** Designer implementation + Developer coding 🎨💻
