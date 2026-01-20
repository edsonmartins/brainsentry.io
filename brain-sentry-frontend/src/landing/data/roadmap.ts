export interface RoadmapPhase {
  quarter: string;
  year: string;
  status: "completed" | "in-progress" | "planned";
  phases: string[];
}

export const roadmap: RoadmapPhase[] = [
  {
    quarter: "Q1",
    year: "2025",
    status: "completed",
    phases: [
      "✅ Foundation - Core infrastructure",
      "✅ Graph Memory - FalkorDB integration",
      "✅ Autonomous Interception",
      "✅ Basic Dashboard",
    ],
  },
  {
    quarter: "Q2",
    year: "2025",
    status: "in-progress",
    phases: [
      "🔄 Note-taking agent",
      "🔄 Architect agent",
      "🔄 Meta-agent",
      "🔄 Advanced visualizations",
    ],
  },
  {
    quarter: "Q3",
    year: "2025",
    status: "planned",
    phases: [
      "📊 Multi-tenant SaaS",
      "📊 Cloud deployment",
      "📊 Advanced analytics",
      "📊 Team collaboration",
    ],
  },
];
