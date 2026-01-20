import { SectionHeader } from "../ui/SectionHeader";
import { PillBadge } from "../ui/PillBadge";
import { Database, Cpu, HardDrive } from "lucide-react";
import { useLanguage } from "../../contexts/LanguageContext";

const layers = [
  {
    name: "CLIENT LAYER",
    icon: Cpu,
    items: ["React Dashboard", "Graph Visualization (Cytoscape.js)", "Admin UI"],
    color: "from-blue-500 to-blue-600",
  },
  {
    name: "BRAIN SENTRY CORE",
    icon: Cpu,
    items: [
      "Autonomous Interception Engine",
      "Memory Management (CRUD + Write ops)",
      "Intelligence Layer (LLM + Patterns)",
    ],
    color: "from-purple-500 to-purple-600",
  },
  {
    name: "DATA LAYER",
    icon: Database,
    items: ["FalkorDB (Graph+Vector)", "PostgreSQL (Audit)", "Redis (Cache)"],
    color: "from-green-500 to-green-600",
  },
];

const techStack = [
  { category: "Backend", items: ["Java 17", "Spring Boot 3.2", "FalkorDB"] },
  { category: "Frontend", items: ["React 19", "Radix UI", "Cytoscape.js"] },
  { category: "AI/ML", items: ["Qwen 2.5-7B", "all-MiniLM-L6-v2"] },
];

export function ArchitectureSection() {
  const { t } = useLanguage();

  return (
    <section className="py-24 bg-muted/30 dark:bg-muted/10">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge={t("architecture.badge")}
          title={t("architecture.title")}
        />

        {/* Architecture Diagram */}
        <div className="max-w-3xl mx-auto mb-16">
          <div className="space-y-4">
            {layers.map((layer, index) => (
              <div key={index} className="relative">
                <div className={`bg-gradient-to-r ${layer.color} rounded-xl p-4 text-white shadow-lg dark:shadow-xl/20`}>
                  <div className="flex items-center gap-3 mb-3">
                    <layer.icon className="w-5 h-5" />
                    <h3 className="font-bold">{layer.name}</h3>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {layer.items.map((item, i) => (
                      <span
                        key={i}
                        className="text-xs px-2 py-1 rounded-full bg-white/20 dark:bg-white/10"
                      >
                        {item}
                      </span>
                    ))}
                  </div>
                </div>
                {index < layers.length - 1 && (
                  <div className="flex justify-center py-2">
                    <div className="w-0.5 h-6 bg-border dark:border-border/60"></div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Tech Stack */}
        <div className="max-w-4xl mx-auto">
          <h3 className="text-center font-bold mb-8 dark:text-white">{t("architecture.stack")}</h3>
          <div className="grid md:grid-cols-3 gap-6">
            {techStack.map((category, index) => (
              <div key={index} className="bg-card dark:bg-card/80 rounded-xl border-2 border-border dark:border-border/50 p-6">
                <h4 className="font-bold mb-3 text-brain-primary dark:text-primary-400">{category.category}</h4>
                <div className="space-y-2">
                  {category.items.map((item, i) => (
                    <div key={i} className="text-sm flex items-center gap-2">
                      <div className="w-1.5 h-1.5 rounded-full bg-brain-accent dark:bg-accent-400" />
                      <span className="text-muted-foreground dark:text-gray-400">{item}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Badges */}
        <div className="mt-12 flex flex-wrap justify-center gap-4">
          <PillBadge variant="primary" className="dark:border-border/60">Self-Hosted</PillBadge>
          <PillBadge variant="accent" className="dark:border-border/60">Open Source</PillBadge>
          <PillBadge variant="success" className="dark:border-border/60">Production-Ready</PillBadge>
          <PillBadge variant="gray" className="dark:border-border/60">API-First</PillBadge>
        </div>
      </div>
    </section>
  );
}
