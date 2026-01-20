import { ArrowRight } from "lucide-react";
import { SectionHeader } from "../ui/SectionHeader";
import { useLanguage } from "../../contexts/LanguageContext";

export function HowItWorksSection() {
  const { t } = useLanguage();

  const steps = [
    {
      number: "1",
      title: t("how.step1.title"),
      description: t("how.step1.desc"),
      icon: "💬",
    },
    {
      number: "2",
      title: t("how.step2.title"),
      description: t("how.step2.desc"),
      icon: "🔍",
    },
    {
      number: "3",
      title: t("how.step3.title"),
      description: t("how.step3.desc"),
      icon: "🧠",
    },
    {
      number: "4",
      title: t("how.step4.title"),
      description: t("how.step4.desc"),
      icon: "✨",
    },
  ];

  return (
    <section className="py-24 bg-muted/30 dark:bg-muted/10">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge={t("how.badge")}
          title={t("how.title")}
          description={t("how.description")}
        />

        {/* Steps Flow */}
        <div className="max-w-5xl mx-auto">
          {/* Desktop: Horizontal flow */}
          <div className="hidden md:grid grid-cols-4 gap-4">
            {steps.map((step, index) => (
              <div key={index} className="relative">
                <div className="bg-card dark:bg-card/90 rounded-xl border-2 border-border dark:border-border/50 p-6 text-center h-full hover:border-brain-primary/50 dark:hover:border-brain-primary/40 transition-colors">
                  <div className="text-4xl mb-3">{step.icon}</div>
                  <div className="text-xs font-bold text-brain-primary dark:text-primary-400 mb-2">{step.number}</div>
                  <h3 className="font-bold mb-2 dark:text-white">{step.title}</h3>
                  <p className="text-sm text-muted-foreground dark:text-gray-400">{step.description}</p>
                </div>
                {index < steps.length - 1 && (
                  <ArrowRight className="absolute -right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-brain-accent dark:text-accent-400" />
                )}
              </div>
            ))}
          </div>

          {/* Mobile: Vertical flow */}
          <div className="md:hidden space-y-4">
            {steps.map((step, index) => (
              <div key={index} className="flex items-center gap-4">
                <div className="flex-shrink-0 w-16 h-16 bg-brain-primary/10 dark:bg-brain-primary/20 rounded-full flex items-center justify-center text-2xl">
                  {step.icon}
                </div>
                <div className="flex-1">
                  <div className="text-xs font-bold text-brain-primary dark:text-primary-400 mb-1">
                    {step.number}
                  </div>
                  <h3 className="font-bold dark:text-white">{step.title}</h3>
                  <p className="text-sm text-muted-foreground dark:text-gray-400">{step.description}</p>
                </div>
              </div>
            ))}
          </div>

          {/* Code Example */}
          <div className="mt-12 bg-card dark:bg-card/90 rounded-xl border-2 border-border dark:border-border/50 p-6 max-w-2xl mx-auto">
            <div className="text-xs font-mono text-muted-foreground dark:text-gray-500 mb-2">{t("how.code1.comment")}</div>
            <pre className="text-sm font-mono overflow-x-auto bg-muted/30 dark:bg-muted/20 p-3 rounded-lg mb-4">
              <code>
                <span className="text-amber-600 dark:text-accent-400">if</span> <span className="text-orange-500 dark:text-primary-300">(needsContext)</span> {"{"}
                <br />
                {"  "}<span className="text-brain-primary dark:text-primary-400">const</span> memories = <span className="text-brain-primary dark:text-primary-400">await</span> searchMemory(query);
                <br />
                <span className="text-gray-500 dark:text-gray-600">// Might forget!</span>
                <br />
                {"}"}
              </code>
            </pre>

            <div className="text-xs font-mono text-muted-foreground dark:text-gray-500 mt-4 mb-2">{t("how.code2.comment")}</div>
            <pre className="text-sm font-mono overflow-x-auto bg-muted/30 dark:bg-muted/20 p-3 rounded-lg">
              <code>
                <span className="text-brain-primary dark:text-primary-400">const</span> enriched = <span className="text-brain-primary dark:text-primary-400">await</span> brainSentry.
                <span className="text-yellow-500 dark:text-yellow-400">intercept</span>(request);
                <br />
                <span className="text-gray-500 dark:text-gray-600">{t("how.code2.note")}</span>
              </code>
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}
