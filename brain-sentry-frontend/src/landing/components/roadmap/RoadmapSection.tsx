import { SectionHeader } from "../ui/SectionHeader";
import { roadmap, type RoadmapPhase } from "../../data/roadmap";
import { CheckCircle, Clock, Rocket } from "lucide-react";
import { useLanguage } from "../../contexts/LanguageContext";

export function RoadmapSection() {
  const { t } = useLanguage();

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "completed":
        return <CheckCircle className="w-5 h-5 text-brain-success dark:text-green-400" />;
      case "in-progress":
        return <Clock className="w-5 h-5 text-brain-gold dark:text-gold-400" />;
      case "planned":
        return <Rocket className="w-5 h-5 text-muted-foreground dark:text-gray-500" />;
      default:
        return null;
    }
  };

  return (
    <section className="py-24 dark:bg-muted/5" id="roadmap">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge={t("roadmap.badge")}
          title={t("roadmap.title")}
        />

        <div className="max-w-4xl mx-auto">
          {/* Timeline */}
          <div className="relative">
            {/* Horizontal Line */}
            <div className="absolute top-8 left-0 right-0 h-1 bg-muted dark:bg-muted/40 hidden md:block" />

            <div className="grid md:grid-cols-3 gap-8">
              {roadmap.map((phase: RoadmapPhase, index: number) => (
                <div key={index} className="relative">
                  {/* Dot */}
                  <div className="hidden md:flex justify-center mb-4">
                    <div
                      className={`w-6 h-6 rounded-full border-4 border-background dark:border-gray-900 flex items-center justify-center ${
                        phase.status === "completed"
                          ? "bg-brain-success dark:bg-green-600"
                          : phase.status === "in-progress"
                          ? "bg-brain-gold dark:bg-gold-500 animate-pulse"
                          : "bg-muted dark:bg-gray-700"
                      }`}
                    >
                      {getStatusIcon(phase.status)}
                    </div>
                  </div>

                  {/* Mobile Year Badge */}
                  <div className="md:hidden flex items-center gap-2 mb-4">
                    <div
                      className={`w-6 h-6 rounded-full border-4 border-background dark:border-gray-900 flex items-center justify-center ${
                        phase.status === "completed"
                          ? "bg-brain-success dark:bg-green-600"
                          : phase.status === "in-progress"
                          ? "bg-brain-gold dark:bg-gold-500"
                          : "bg-muted dark:bg-gray-700"
                      }`}
                    >
                      {getStatusIcon(phase.status)}
                    </div>
                    <span className="text-sm font-bold text-muted-foreground dark:text-gray-400">
                      {phase.quarter} {phase.year}
                    </span>
                  </div>

                  {/* Content */}
                  <div
                    className={`bg-card dark:bg-card/80 rounded-xl border-2 border-border dark:border-border/50 p-6 h-full ${
                      phase.status === "in-progress" ? "border-brain-gold dark:border-gold-500/60 shadow-lg dark:shadow-amber-500/20" : ""
                    }`}
                  >
                    <div className="md:hidden text-xs font-bold text-muted-foreground dark:text-gray-400 mb-2">
                      {phase.quarter} {phase.year}
                    </div>

                    <ul className="space-y-2">
                      {phase.phases.map((item: string, i: number) => (
                        <li key={i} className="text-sm flex items-start gap-2">
                          {item.startsWith("✅") ? (
                            <CheckCircle className="w-4 h-4 text-brain-success dark:text-green-400 flex-shrink-0 mt-0.5" />
                          ) : item.startsWith("🔄") ? (
                            <Clock className="w-4 h-4 text-brain-gold dark:text-gold-400 flex-shrink-0 mt-0.5" />
                          ) : item.startsWith("📊") ? (
                            <span className="flex-shrink-0 mt-0.5">📊</span>
                          ) : (
                            <span className="w-1.5 h-1.5 rounded-full bg-brain-primary dark:bg-primary-400 flex-shrink-0 mt-1.5" />
                          )}
                          <span className="text-muted-foreground dark:text-gray-400">{item.slice(2)}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Current Status */}
          <div className="mt-12 text-center">
            <div className="inline-flex items-center gap-3 px-6 py-3 bg-brain-primary/10 dark:bg-brain-primary/20 rounded-full border-2 border-brain-primary/20 dark:border-brain-primary/40">
              <Clock className="w-5 h-5 text-brain-gold dark:text-gold-400" />
              <span className="font-semibold text-brain-primary dark:text-primary-300">
                {t("roadmap.current")}: Phase 1 (Q1 2025) • {t("roadmap.mvp")}: March 2025
              </span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
