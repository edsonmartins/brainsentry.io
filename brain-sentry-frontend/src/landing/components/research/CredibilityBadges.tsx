import { GraduationCap } from "lucide-react";
import { useLanguage } from "../../contexts/LanguageContext";

const institutions = [
  { name: "Meta AI", highlight: true },
  { name: "Harvard", highlight: true },
  { name: "Microsoft Research", highlight: true },
  { name: "Princeton", highlight: true },
  { name: "Stanford", highlight: true },
  { name: "UC Berkeley", highlight: true },
  { name: "DeepMind", highlight: true },
  { name: "UCL", highlight: true },
  { name: "Scale AI", highlight: true },
];

export function CredibilityBadges() {
  const { t } = useLanguage();

  return (
    <section className="py-16 border-t border-border dark:border-border/60 dark:bg-muted/5">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-8">
          <h3 className="text-lg font-bold text-muted-foreground dark:text-gray-400">
            {t("credibility.title")}
          </h3>
        </div>

        {/* Institution Grid */}
        <div className="grid grid-cols-3 md:grid-cols-9 gap-4 max-w-4xl mx-auto mb-8">
          {institutions.map((inst, index) => (
            <div
              key={index}
              className={`flex items-center justify-center p-4 rounded-lg border-2 border-border dark:border-border/50 text-sm font-medium transition-all hover:shadow-md ${
                inst.highlight
                  ? "bg-orange-50 dark:bg-primary-950/40 border-orange-200 dark:border-primary-800/60 text-orange-700 dark:text-primary-300"
                  : "bg-muted/50 dark:bg-muted/30 text-muted-foreground dark:text-gray-400"
              }`}
            >
              <GraduationCap className="w-4 h-4 mr-1" />
              {inst.name}
            </div>
          ))}
        </div>

        {/* Stats */}
        <div className="flex flex-wrap justify-center gap-6 text-sm text-muted-foreground dark:text-gray-400">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground dark:text-gray-300">📄</span>
            {t("credibility.papers")}
          </div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground dark:text-gray-300">📊</span>
            {t("credibility.benchmarks")}
          </div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground dark:text-gray-300">⭐</span>
            {t("credibility.reference")}
          </div>
        </div>
      </div>
    </section>
  );
}
