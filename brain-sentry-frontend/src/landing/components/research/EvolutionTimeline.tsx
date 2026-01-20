import { SectionHeader } from "../ui/SectionHeader";
import { useLanguage } from "../../contexts/LanguageContext";

interface TimelineEvent {
  year: string;
  isGroup?: boolean;
  current?: boolean;
  title?: string;
  subtitle?: string;
  description?: string;
  detail?: string;
  events?: SubEvent[];
}

interface SubEvent {
  title: string;
  subtitle: string;
  description: string;
  featured?: boolean;
  isBrainSentry?: boolean;
}

const timelineEvents: TimelineEvent[] = [
  {
    year: "2020",
    title: "RAG",
    subtitle: "Lewis et al., Meta AI",
    description: "Read-only retrieval",
    detail: "Vector search",
  },
  {
    year: "2023",
    title: "MemGPT",
    subtitle: "UC Berkeley",
    description: "Hierarchical memory",
    detail: "OS-inspired architecture",
  },
  {
    year: "2024",
    isGroup: true,
    events: [
      {
        title: "GraphRAG",
        subtitle: "Microsoft",
        description: "Graph + Vector hybrid",
      },
      {
        title: "CoALA",
        subtitle: "Princeton/Stanford",
        description: "Cognitive framework",
      },
      {
        title: "Agent Memory",
        subtitle: "Monigatti",
        description: "Read-write operations",
      },
    ],
  },
  {
    year: "2025",
    isGroup: true,
    current: true,
    events: [
      {
        title: "Confucius",
        subtitle: "Meta/Harvard",
        description: "🏆 54.3% SWE-Bench-Pro",
        featured: true,
      },
      {
        title: "BRAIN SENTRY",
        subtitle: "IntegrAllTech",
        description: "🌟 Graph-Native + Autonomous",
        isBrainSentry: true,
      },
    ],
  },
];

export function EvolutionTimeline() {
  const { t } = useLanguage();

  return (
    <section className="py-24 bg-muted/30 dark:bg-muted/10">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge="EVOLUTION"
          title="The Evolution of Agent Memory"
          description="From read-only RAG to full Agent Memory with graph-native architecture"
        />

        <div className="max-w-3xl mx-auto">
          {/* Vertical line */}
          <div className="relative">
            <div className="absolute left-8 top-0 bottom-0 w-0.5 bg-gradient-to-b from-brain-primary via-brain-accent to-brain-primary hidden md:block"></div>

            <div className="space-y-12">
              {timelineEvents.map((event: TimelineEvent, index: number) => (
                <div key={index} className="relative md:pl-20">
                  {/* Dot */}
                  <div
                    className={`absolute left-6 md:left-5.5 top-2 w-5 h-5 rounded-full border-4 border-background dark:border-gray-900 ${
                      event.current || event.isGroup
                        ? "bg-brain-gold shadow-lg shadow-brain-gold/50 animate-pulse-slow"
                        : "bg-brain-primary dark:bg-primary-500"
                    }`}
                  />

                  {/* Year */}
                  <div className="md:hidden mb-4">
                    <div className="text-sm font-bold text-brain-primary dark:text-primary-400">
                      {event.year} ●━━━━━━━━━━━━━━━━━━━━━
                    </div>
                  </div>

                  {/* Desktop Year */}
                  <div className="hidden md:block absolute -left-12 top-2 text-sm font-bold text-brain-primary dark:text-primary-400">
                    {event.year}
                  </div>

                  {/* Content */}
                  {event.isGroup && event.events ? (
                    <div className="space-y-6">
                      {event.events.map((subEvent: SubEvent, i: number) => (
                        <div
                          key={i}
                          className={`bg-card dark:bg-card/80 rounded-xl border-2 border-border dark:border-border/50 p-6 ${
                            subEvent.isBrainSentry
                              ? "border-2 border-brain-accent shadow-lg dark:shadow-brain-accent/30"
                              : subEvent.featured
                              ? "border-brain-primary dark:border-primary-900"
                              : "border-border dark:border-border/60"
                          }`}
                        >
                          <div className="flex items-start justify-between mb-2">
                            <div>
                              <h3 className="font-bold text-lg dark:text-white">{subEvent.title}</h3>
                              <p className="text-sm text-muted-foreground dark:text-gray-400">
                                {subEvent.subtitle}
                              </p>
                            </div>
                            {subEvent.featured && (
                              <span className="text-2xl">🏆</span>
                            )}
                            {subEvent.isBrainSentry && (
                              <span className="text-2xl">🌟</span>
                            )}
                          </div>
                          <p className="text-sm text-muted-foreground dark:text-gray-400">
                            {subEvent.description}
                          </p>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="bg-card dark:bg-card/80 rounded-xl border-2 border-border dark:border-border/50 p-6">
                      <h3 className="font-bold text-lg mb-1 dark:text-white">{event.title}</h3>
                      <p className="text-sm text-muted-foreground dark:text-gray-400 mb-2">
                        {event.subtitle}
                      </p>
                      <p className="text-sm dark:text-gray-300">{event.description}</p>
                      {event.detail && (
                        <p className="text-xs text-muted-foreground dark:text-gray-500 mt-1">
                          └─ {event.detail}
                        </p>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
