export interface Feature {
  icon: string;
  title: string;
  headline: string;
  description: string;
  tech?: string[];
}

export const features: Feature[] = [
  {
    icon: "Network",
    title: "Graph-Native",
    headline: "Relationships Are Everything",
    description: "FalkorDB provides graph + vector search in one database. Query relationships natively, detect conflicts automatically, and visualize your knowledge graph interactively.",
    tech: ["FalkorDB", "Native Cypher", "GraphRAG"],
  },
  {
    icon: "Zap",
    title: "Autonomous",
    headline: "Never Miss Context",
    description: "Traditional agent memory requires the AI to explicitly search. This means it can forget to check, waste tool calls, and miss relevant context. Brain Sentry always analyzes. Every query. No exceptions.",
    tech: [],
  },
  {
    icon: "Brain",
    title: "Typed Memory",
    headline: "Four Types, Optimized Retrieval",
    description: "Semantic: Technical facts. Episodic: Past events. Procedural: Patterns and how-tos. Associative: Relationships between components. The right memory, retrieved the right way, every time.",
    tech: [],
  },
  {
    icon: "FileSearch",
    title: "Full Audit",
    headline: "Production-Ready from Day 1",
    description: "Every memory has version history (rollback to any point), provenance tracking (who created, when, why), impact analysis (what depends on this), and conflict detection.",
    tech: [],
  },
  {
    icon: "CircleDot",
    title: "Graph Visualization",
    headline: "See Your Knowledge",
    description: "Interactive graph visualization powered by Cytoscape.js. Explore 10,000+ nodes without lag, advanced layout algorithms, highlight connected components, dark mode support.",
    tech: ["Cytoscape.js", "Dark mode"],
  },
  {
    icon: "Code",
    title: "Developer-Focused",
    headline: "Built for Code, Not Just Chat",
    description: "Brain Sentry understands code patterns and antipatterns, architectural decisions and trade-offs, integration points and dependencies, bug histories and resolutions.",
    tech: [],
  },
];
