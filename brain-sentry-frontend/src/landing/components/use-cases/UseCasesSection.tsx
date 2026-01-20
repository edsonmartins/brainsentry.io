import { Quote } from "lucide-react";
import { SectionHeader } from "../ui/SectionHeader";
import { useCases } from "../../data/useCases";
import { useLanguage } from "../../contexts/LanguageContext";

export function UseCasesSection() {
  const { t } = useLanguage();

  return (
    <section className="py-24 dark:bg-muted/5" id="use-cases">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge={t("usecases.badge")}
          title={t("usecases.title")}
          description={t("usecases.description")}
        />

        <div className="grid md:grid-cols-2 gap-8 max-w-6xl mx-auto">
          {useCases.map((useCase, index) => (
            <div
              key={index}
              className="bg-card dark:bg-card/80 rounded-xl border-2 border-border dark:border-border/50 p-6 hover:shadow-lg dark:hover:shadow-lg/50 transition-all duration-300"
            >
              {/* Title & Result Badge */}
              <div className="flex items-start justify-between mb-4">
                <h3 className="text-xl font-bold dark:text-white">{useCase.title}</h3>
                <div className="text-xs px-2 py-1 rounded-full bg-emerald-100 dark:bg-emerald-900/30 dark:text-emerald-400 font-medium border border-emerald-200 dark:border-emerald-800">
                  {useCase.result}
                </div>
              </div>

              {/* Challenge */}
              <div className="mb-4">
                <div className="text-xs font-semibold text-red-500 dark:text-red-400 uppercase mb-1">
                  {t("usecases.challenge")}
                </div>
                <p className="text-sm text-muted-foreground dark:text-gray-400">{useCase.challenge}</p>
              </div>

              {/* Solution */}
              <div className="mb-4">
                <div className="text-xs font-semibold text-brain-primary dark:text-primary-400 uppercase mb-1">
                  {t("usecases.solution")}
                </div>
                <p className="text-sm text-muted-foreground dark:text-gray-400">{useCase.solution}</p>
              </div>

              {/* Quote */}
              <div className="relative bg-muted/50 dark:bg-muted/30 rounded-lg p-4 border-l-4 border-brain-accent dark:border-brain-accent/60">
                <Quote className="absolute -top-2 -left-2 w-5 h-5 text-brain-accent dark:text-accent-400 bg-background dark:bg-gray-900 rounded-full p-0.5" />
                <p className="text-sm italic text-muted-foreground dark:text-gray-400 mb-2">
                  &quot;{useCase.quote}&quot;
                </p>
                <p className="text-xs text-muted-foreground dark:text-gray-500">— {useCase.author}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
