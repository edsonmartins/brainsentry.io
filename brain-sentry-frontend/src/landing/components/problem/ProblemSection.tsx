import { SectionHeader } from "../ui/SectionHeader";
import { X, Check } from "lucide-react";
import { useLanguage } from "../../contexts/LanguageContext";

export function ProblemSection() {
  const { t } = useLanguage();

  const problems = [
    { feature: t("problem.feature1"), without: t("problem.feature1.without"), with: t("problem.feature1.with") },
    { feature: t("problem.feature2"), without: t("problem.feature2.without"), with: t("problem.feature2.with") },
    { feature: t("problem.feature3"), without: t("problem.feature3.without"), with: t("problem.feature3.with") },
    { feature: t("problem.feature4"), without: t("problem.feature4.without"), with: t("problem.feature4.with") },
    { feature: t("problem.feature5"), without: t("problem.feature5.without"), with: t("problem.feature5.with") },
  ];

  return (
    <section className="py-24 bg-muted/30 dark:bg-muted/10">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          title={t("problem.title")}
          description={t("problem.description")}
        />

        <div className="max-w-4xl mx-auto">
          {/* Comparison Table */}
          <div className="bg-card dark:bg-card/95 rounded-xl border-2 border-border dark:border-border/60 shadow-lg overflow-hidden">
            <div className="grid grid-cols-3">
              <div className="p-4 bg-muted/80 dark:bg-muted/30 border-b-2 border-border dark:border-border/50 text-center font-semibold text-sm">
                {t("problem.table.feature")}
              </div>
              <div className="p-4 bg-muted/80 dark:bg-muted/30 border-b-2 border-border dark:border-border/50 text-center font-semibold text-sm text-destructive">
                {t("problem.table.without")}
              </div>
              <div className="p-4 bg-brain-primary/20 dark:bg-brain-primary/10 border-b-2 border-brain-primary/30 dark:border-brain-primary/20 text-center font-semibold text-sm text-brain-primary dark:text-brain-primary">
                {t("problem.table.with")}
              </div>
            </div>

            {problems.map((item, index) => (
              <div
                key={index}
                className="grid grid-cols-3 border-b border-border dark:border-border/50 last:border-b-0 hover:bg-muted/30 dark:hover:bg-muted/20 transition-colors"
              >
                <div className="p-4 border-r border-border dark:border-border/40 font-medium text-sm text-foreground dark:text-foreground">
                  {item.feature}
                </div>
                <div className="p-4 border-r border-border dark:border-border/40 flex items-center gap-2 text-muted-foreground dark:text-muted-foreground">
                  <X className="w-5 h-5 text-destructive flex-shrink-0" />
                  <span className="text-sm">{item.without}</span>
                </div>
                <div className="p-4 flex items-center gap-2">
                  <Check className="w-5 h-5 text-brain-success flex-shrink-0" />
                  <span className="text-sm font-medium text-foreground dark:text-foreground">{item.with}</span>
                </div>
              </div>
            ))}
          </div>

          {/* Additional Context */}
          <p className="mt-8 text-center text-muted-foreground dark:text-muted-foreground max-w-2xl mx-auto">
            {t("problem.footer")}
            <br />
            <span className="font-semibold text-foreground dark:text-foreground">{t("problem.footer.highlight")}</span>
          </p>
        </div>
      </div>
    </section>
  );
}
