import { Brain, Zap, Network } from "lucide-react";
import { SectionHeader } from "../ui/SectionHeader";
import { GradientText } from "../ui/GradientText";
import { useLanguage } from "../../contexts/LanguageContext";

export function SolutionSection() {
  const { t } = useLanguage();

  const pillars = [
    {
      icon: Brain,
      title: t("solution.pillar1.title"),
      description: t("solution.pillar1.desc"),
      types: [
        { name: t("solution.pillar1.type1"), desc: t("solution.pillar1.type1.desc") },
        { name: t("solution.pillar1.type2"), desc: t("solution.pillar1.type2.desc") },
        { name: t("solution.pillar1.type3"), desc: t("solution.pillar1.type3.desc") },
        { name: t("solution.pillar1.type4"), desc: t("solution.pillar1.type4.desc") },
      ],
    },
    {
      icon: Zap,
      title: t("solution.pillar2.title"),
      description: t("solution.pillar2.desc"),
      types: [],
      highlight: t("solution.pillar2.highlight"),
    },
    {
      icon: Network,
      title: t("solution.pillar3.title"),
      description: t("solution.pillar3.desc"),
      types: [
        { name: t("solution.pillar3.item1"), desc: t("solution.pillar3.item1.desc") },
        { name: t("solution.pillar3.item2"), desc: t("solution.pillar3.item2.desc") },
        { name: t("solution.pillar3.item3"), desc: t("solution.pillar3.item3.desc") },
      ],
      highlight: t("solution.pillar3.highlight"),
    },
  ];

  return (
    <section className="py-24 dark:bg-muted/5" id="features">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge={t("solution.badge")}
          title={
            <>
              {t("solution.title")}{" "}
              <GradientText>Developers</GradientText>
            </>
          }
        />

        <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          {pillars.map((pillar, index) => (
            <div
              key={index}
              className="group bg-card dark:bg-card/80 rounded-xl border-2 border-border dark:border-border/50 p-6 hover:shadow-lg dark:hover:shadow-lg/50 hover:border-brain-primary/50 dark:hover:border-brain-primary/40 transition-all duration-300 hover:-translate-y-1"
            >
              <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-brain-primary to-brain-accent flex items-center justify-center mb-4">
                <pillar.icon className="w-6 h-6 text-white" />
              </div>
              <h3 className="text-xl font-bold mb-3 dark:text-white">{pillar.title}</h3>
              <p className="text-muted-foreground dark:text-gray-400 text-sm mb-4">{pillar.description}</p>

              {pillar.types.length > 0 && (
                <ul className="space-y-2">
                  {pillar.types.map((type, i) => (
                    <li key={i} className="flex items-start gap-2 text-sm">
                      <span className="text-brain-primary mt-0.5">•</span>
                      <div>
                        <span className="font-medium dark:text-gray-200">{type.name}</span>
                        {type.desc && <span className="text-muted-foreground dark:text-gray-400">: {type.desc}</span>}
                      </div>
                    </li>
                  ))}
                </ul>
              )}

              {pillar.highlight && (
                <p className="text-sm font-medium text-brain-accent dark:text-brain-accent/80 mt-4 italic">
                  {pillar.highlight}
                </p>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
