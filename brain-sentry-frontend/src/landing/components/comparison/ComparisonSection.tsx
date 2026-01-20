import { SectionHeader } from "../ui/SectionHeader";
import { Check, X, Minus, Star } from "lucide-react";
import { PillBadge } from "../ui/PillBadge";
import { useLanguage } from "../../contexts/LanguageContext";

type FeatureName =
  | "Semantic Memory"
  | "Episodic Memory"
  | "Procedural Memory"
  | "Associative Memory"
  | "Graph-Native"
  | "Autonomous"
  | "Full Audit"
  | "Dev-Focused"
  | "Open Source";

interface ComparisonFeature {
  name: FeatureName;
  mem0: "yes" | "no" | "partial";
  zep: "yes" | "no" | "partial";
  confucius: "yes" | "no" | "partial";
  brainSentry: "yes" | "no" | "partial" | "exclusive";
}

const comparisonFeatures: ComparisonFeature[] = [
  { name: "Semantic Memory", mem0: "yes", zep: "yes", confucius: "yes", brainSentry: "yes" },
  { name: "Episodic Memory", mem0: "yes", zep: "yes", confucius: "yes", brainSentry: "yes" },
  { name: "Procedural Memory", mem0: "no", zep: "no", confucius: "yes", brainSentry: "yes" },
  { name: "Associative Memory", mem0: "no", zep: "no", confucius: "no", brainSentry: "exclusive" },
  { name: "Graph-Native", mem0: "no", zep: "no", confucius: "no", brainSentry: "exclusive" },
  { name: "Autonomous", mem0: "no", zep: "no", confucius: "no", brainSentry: "exclusive" },
  { name: "Full Audit", mem0: "partial", zep: "partial", confucius: "no", brainSentry: "exclusive" },
  { name: "Dev-Focused", mem0: "no", zep: "no", confucius: "partial", brainSentry: "exclusive" },
  { name: "Open Source", mem0: "yes", zep: "partial", confucius: "yes", brainSentry: "yes" },
];

const StatusIcon = ({ status }: { status: "yes" | "no" | "partial" | "exclusive" }) => {
  if (status === "yes") {
    return <Check className="w-5 h-5 text-brain-success dark:text-green-400" />;
  }
  if (status === "no") {
    return <X className="w-5 h-5 text-muted-foreground/30 dark:text-gray-600" />;
  }
  if (status === "exclusive") {
    return (
      <div className="flex items-center gap-1">
        <Star className="w-5 h-5 text-brain-gold dark:text-orange-400 fill-brain-gold dark:fill-orange-400" />
        <Check className="w-5 h-5 text-brain-success dark:text-green-400" />
      </div>
    );
  }
  return <Minus className="w-5 h-5 text-amber-500 dark:text-gold-400" />;
};

export function ComparisonSection() {
  const { t } = useLanguage();

  return (
    <section className="py-24 bg-muted/30 dark:bg-muted/10" id="comparison">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge={t("comparison.badge")}
          title={t("comparison.title")}
        />

        <div className="max-w-5xl mx-auto overflow-x-auto">
          <table className="w-full bg-card dark:bg-card/90 rounded-xl border-2 border-border dark:border-border/50 overflow-hidden">
            <thead>
              <tr className="border-b-2 border-border dark:border-border/50 bg-muted/50 dark:bg-muted/30">
                <th className="p-4 text-left font-semibold dark:text-gray-200">{t("comparison.feature")}</th>
                <th className="p-4 text-center font-semibold dark:text-gray-200">Mem0</th>
                <th className="p-4 text-center font-semibold dark:text-gray-200">Zep</th>
                <th className="p-4 text-center font-semibold dark:text-gray-200">Confucius</th>
                <th className="p-4 text-center font-semibold text-brain-primary dark:text-primary-400 bg-brain-primary/5 dark:bg-brain-primary/10">
                  Brain Sentry
                </th>
              </tr>
            </thead>
            <tbody>
              {comparisonFeatures.map((feature, index) => (
                <tr key={index} className="border-b border-border dark:border-border/40 last:border-b-0 hover:bg-muted/20 dark:hover:bg-muted/10 transition-colors">
                  <td className="p-4 font-medium dark:text-gray-200">{feature.name}</td>
                  <td className="p-4 text-center">
                    <StatusIcon status={feature.mem0} />
                  </td>
                  <td className="p-4 text-center">
                    <StatusIcon status={feature.zep} />
                  </td>
                  <td className="p-4 text-center">
                    <StatusIcon status={feature.confucius} />
                  </td>
                  <td className="p-4 text-center">
                    <StatusIcon status={feature.brainSentry} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="mt-8 text-center">
          <PillBadge variant="purple">
            {t("comparison.exclusive")}
          </PillBadge>
        </div>

        {/* Key Differences */}
        <div className="mt-12 max-w-3xl mx-auto">
          <h3 className="text-lg font-bold mb-4 text-center dark:text-white">{t("comparison.difference")}</h3>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="flex items-start gap-3 p-4 bg-card dark:bg-card/80 rounded-lg border-2 border-border dark:border-border/50">
              <Check className="w-5 h-5 text-brain-primary dark:text-primary-400 flex-shrink-0 mt-0.5" />
              <div>
                <div className="font-semibold dark:text-gray-200">{t("comparison.diff1.title")}</div>
                <div className="text-sm text-muted-foreground dark:text-gray-400">
                  {t("comparison.diff1.desc")}
                </div>
              </div>
            </div>
            <div className="flex items-start gap-3 p-4 bg-card dark:bg-card/80 rounded-lg border-2 border-border dark:border-border/50">
              <Check className="w-5 h-5 text-brain-primary dark:text-primary-400 flex-shrink-0 mt-0.5" />
              <div>
                <div className="font-semibold dark:text-gray-200">{t("comparison.diff2.title")}</div>
                <div className="text-sm text-muted-foreground dark:text-gray-400">
                  {t("comparison.diff2.desc")}
                </div>
              </div>
            </div>
            <div className="flex items-start gap-3 p-4 bg-card dark:bg-card/80 rounded-lg border-2 border-border dark:border-border/50">
              <Check className="w-5 h-5 text-brain-primary dark:text-primary-400 flex-shrink-0 mt-0.5" />
              <div>
                <div className="font-semibold dark:text-gray-200">{t("comparison.diff3.title")}</div>
                <div className="text-sm text-muted-foreground dark:text-gray-400">
                  {t("comparison.diff3.desc")}
                </div>
              </div>
            </div>
            <div className="flex items-start gap-3 p-4 bg-card dark:bg-card/80 rounded-lg border-2 border-border dark:border-border/50">
              <Check className="w-5 h-5 text-brain-primary dark:text-primary-400 flex-shrink-0 mt-0.5" />
              <div>
                <div className="font-semibold dark:text-gray-200">{t("comparison.diff4.title")}</div>
                <div className="text-sm text-muted-foreground dark:text-gray-400">
                  {t("comparison.diff4.desc")}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
