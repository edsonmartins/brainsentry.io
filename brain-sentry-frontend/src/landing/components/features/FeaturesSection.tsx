import {
  Network,
  Zap,
  Brain,
  FileSearch,
  CircleDot,
  Code,
} from "lucide-react";
import { SectionHeader } from "../ui/SectionHeader";
import { features, type Feature } from "../../data/features";
import { useLanguage } from "../../contexts/LanguageContext";

const iconMap: Record<string, React.ElementType> = {
  Network,
  Zap,
  Brain,
  FileSearch,
  CircleDot,
  Code,
};

export function FeaturesSection() {
  const { t } = useLanguage();

  return (
    <section className="py-24 dark:bg-muted/5" id="feature-cards">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge={t("features.badge")}
          title={t("features.title")}
        />

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 max-w-6xl mx-auto">
          {features.map((feature, index) => {
            const Icon = iconMap[feature.icon] || Network;
            return (
              <div
                key={index}
                className="group bg-card dark:bg-card/80 rounded-xl border-2 border-border dark:border-border/50 p-6 hover:shadow-lg dark:hover:shadow-lg/50 hover:border-brain-primary/50 dark:hover:border-brain-primary/40 transition-all duration-300 hover:-translate-y-1"
              >
                <div className="w-12 h-12 rounded-lg bg-brain-primary/10 dark:bg-brain-primary/20 flex items-center justify-center mb-4 group-hover:bg-brain-primary/20 dark:group-hover:bg-brain-primary/30 transition-colors">
                  <Icon className="w-6 h-6 text-brain-primary dark:text-primary-400" />
                </div>
                <h3 className="text-lg font-bold mb-2 dark:text-white">{feature.headline}</h3>
                <p className="text-sm text-muted-foreground dark:text-gray-400 mb-4">{feature.description}</p>
                {feature.tech && feature.tech.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {feature.tech.map((tech, i) => (
                      <span
                        key={i}
                        className="text-xs px-2 py-1 rounded-full bg-muted dark:bg-muted/40 text-muted-foreground dark:text-gray-400 border border-border dark:border-border/40"
                      >
                        {tech}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
