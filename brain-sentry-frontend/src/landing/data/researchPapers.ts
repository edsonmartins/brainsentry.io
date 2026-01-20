export interface ResearchPaper {
  id: string;
  title: string;
  authors: string[];
  institution: string;
  date: string;
  source: string;
  arxiv?: string;
  arxivUrl?: string;
  featured?: boolean;
  badge?: string;
  borderColor?: string;
  keyResults: string[];
  alignmentScore: number;
  insights: string[];
  implemented?: string[];
  benchmark?: {
    confucius: string;
    anthropic: string;
    liveSWE: string;
    brainSentryTarget: string;
  };
  gaps?: {
    mostSystemsLack: string;
    fewHave: string;
    limited: string;
    weak: string;
  };
  brainSentryFills?: string[];
  limitationsIdentified?: string[];
  brainSentryEvolution?: string[];
}

export const researchPapers: ResearchPaper[] = [
  {
    id: "confucius",
    title: "Confucius Code Agent: Scalable Agent Scaffolding for Real-World Codebases",
    authors: ["Wong, Sherman", "Qi, Zhenting", "Wang, Zhaodong"],
    institution: "Meta AI & Harvard University",
    date: "December 2025",
    source: "arXiv",
    arxiv: "2512.10398v5",
    arxivUrl: "https://arxiv.org/abs/2512.10398v5",
    featured: true,
    badge: "META AI & HARVARD",
    borderColor: "#3B82F6",
    keyResults: [
      "🎯 54.3% on SWE-Bench-Pro (State-of-the-Art)",
      "🎯 Proves scaffolding > model capability",
      "🎯 Hierarchical memory architecture",
      "🎯 Note-taking agent for cross-session learning",
    ],
    alignmentScore: 85,
    insights: [
      "Memory architecture matters more than LLM",
      "Context compression is critical",
      "Note-taking enables cross-session learning",
    ],
    implemented: [
      "Phase 3: Note-taking agent (Confucius-inspired)",
      "Phase 3: Architect agent for compression",
      "Phase 5: Meta-agent (build-test-improve loop)",
    ],
  },
  {
    id: "rag-to-agent-memory",
    title: "From RAG to Agent Memory",
    authors: ["Monigatti, Leonie"],
    institution: "Independent AI Researcher",
    date: "2024",
    source: "Towards Data Science",
    featured: true,
    badge: "INDUSTRY THOUGHT LEADERSHIP",
    borderColor: "#8B5CF6",
    keyResults: [
      "Industry evolving from RAG to Agent Memory",
      "Multi-type memory systems essential for advanced agents",
    ],
    alignmentScore: 95,
    insights: [
      "We implement all 3 types PLUS Associative Memory (relationships via graph) - unique to Brain Sentry",
    ],
    implemented: [
      "Validated our 4 memory types architecture",
      "Confirmed need for write operations",
      "Inspired autonomous (not tool-based) approach",
    ],
  },
  {
    id: "coala",
    title: "CoALA: Cognitive Architectures for Language Agents",
    authors: ["Sumers, T.", "Yao, S.", "Narasimhan, K.", "Griffiths, T."],
    institution: "Princeton • Stanford • DeepMind",
    date: "2024",
    source: "arXiv",
    featured: false,
    keyResults: ["Agents need structured cognitive architectures that separate different types of memory and knowledge."],
    alignmentScore: 100,
    insights: [
      "Memory type separation is scientifically validated",
      "Working memory needs hierarchical structure",
      "Procedural memory crucial for code patterns",
    ],
    implemented: [],
  },
  {
    id: "graphrag",
    title: "GraphRAG: Knowledge Graph-Based RAG",
    authors: ["Microsoft Research"],
    institution: "Microsoft Research",
    date: "2024",
    source: "arXiv",
    featured: false,
    keyResults: [
      "📊 2-3x improvement over vector-only RAG",
      "📊 Multi-hop reasoning enabled",
      "📊 Community detection finds implicit patterns",
    ],
    alignmentScore: 100,
    insights: [
      "FalkorDB provides native Graph + Vector in one DB",
      "We implement GraphRAG from day one",
    ],
    implemented: [
      "Validated our FalkorDB choice",
      "Confirmed graph-first architecture",
      "Inspired GraphRAG implementation",
    ],
  },
  {
    id: "memgpt",
    title: "MemGPT: Towards LLMs as Operating Systems",
    authors: ["Packer, C.", "Wooders, S.", "Lin, K.", "et al."],
    institution: "UC Berkeley",
    date: "2023",
    source: "arXiv",
    featured: false,
    keyResults: ["LLMs need hierarchical memory management inspired by operating systems."],
    alignmentScore: 75,
    insights: [
      "Hierarchical concept ✓ Graph-native storage ✓",
      "Different: Graph vs flat text chunks",
    ],
    implemented: [
      "Memory lifecycle management strategies",
      "Importance of forgetting mechanisms",
      "Context window compression techniques",
    ],
  },
  {
    id: "swe-bench",
    title: "SWE-Bench Pro: Long-Horizon Software Engineering",
    authors: ["Deng, X.", "Da, J.", "Pan, E.", "et al."],
    institution: "Princeton University • Scale AI",
    date: "2025",
    source: "arXiv",
    featured: false,
    keyResults: ["Industry-standard benchmark for evaluating AI agents on real-world GitHub issues."],
    benchmark: {
      confucius: "54.3%",
      anthropic: "52.0%",
      liveSWE: "45.8%",
      brainSentryTarget: ">55%",
    },
    insights: [
      "Proves memory architecture = performance",
      "Validates long-context capabilities",
      "Industry-standard benchmark",
    ],
    alignmentScore: 100,
    implemented: [],
  },
  {
    id: "agent-memory-survey",
    title: "Survey: Memory in LLM-Based Agents",
    authors: ["Multiple Authors (Survey Paper)"],
    institution: "arXiv",
    date: "2024",
    source: "arXiv",
    featured: false,
    keyResults: ["Categorizes existing memory systems and identifies critical gaps in current research."],
    alignmentScore: 100,
    gaps: {
      mostSystemsLack: "graph relationships",
      fewHave: "full auditability",
      limited: "production deployments",
      weak: "developer-specific features",
    },
    brainSentryFills: [
      "✅ Graph-native relationships",
      "✅ Full audit trail",
      "✅ Production-ready architecture",
      "✅ Developer-focused memory types",
    ],
    insights: [],
    implemented: [],
  },
  {
    id: "rag-foundational",
    title: "Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks",
    authors: ["Lewis, P.", "Perez, E.", "Piktus, A.", "et al."],
    institution: "Meta AI • University College London",
    date: "2020",
    source: "NeurIPS",
    featured: false,
    badge: "FOUNDATIONAL PAPER",
    keyResults: ["The foundational RAG paper that started it all. Showed retrieval before generation reduces hallucinations."],
    alignmentScore: 100,
    limitationsIdentified: [
      "❌ Read-only (can't learn)",
      "❌ No write operations",
      "❌ No relationships between documents",
      "❌ Static knowledge base",
    ],
    brainSentryEvolution: [
      "✅ Read-write operations",
      "✅ Graph relationships",
      "✅ Continuous learning",
      "✅ Multi-type memory",
    ],
    insights: [
      "Standing on the shoulders of this giant!",
    ],
    implemented: [],
  },
];
