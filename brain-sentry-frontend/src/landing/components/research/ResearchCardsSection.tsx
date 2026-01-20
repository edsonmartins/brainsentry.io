import { Award, BookOpen } from "lucide-react";
import { SectionHeader } from "../ui/SectionHeader";
import { PillBadge } from "../ui/PillBadge";
import { researchPapers } from "../../data/researchPapers";
import { ResearchCard } from "./ResearchCard";
import { useLanguage } from "../../contexts/LanguageContext";

export function ResearchCardsSection() {
  const { t } = useLanguage();
  const featuredPapers = researchPapers.filter((p) => p.featured);
  const standardPapers = researchPapers.filter((p) => !p.featured);

  return (
    <section className="py-24 dark:bg-muted/5" id="research">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge="ACADEMIC VALIDATION"
          title={
            <>
              Research-Backed <span className="text-brain-primary dark:text-primary-400">Architecture</span>
            </>
          }
          description="Our design is validated by leading research from Meta, Harvard, Princeton, Stanford, UC Berkeley, and Microsoft Research."
        />

        {/* Featured Papers */}
        <div className="grid md:grid-cols-2 gap-8 max-w-6xl mx-auto mb-12">
          {featuredPapers.map((paper) => (
            <ResearchCard key={paper.id} paper={paper} featured />
          ))}
        </div>

        {/* Standard Papers */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 max-w-7xl mx-auto">
          {standardPapers.map((paper) => (
            <ResearchCard key={paper.id} paper={paper} featured={false} />
          ))}
        </div>

        {/* Research Summary */}
        <div className="mt-16 text-center">
          <div className="inline-flex items-center gap-2 px-6 py-3 bg-muted/50 dark:bg-muted/30 rounded-full border-2 border-border dark:border-border/60">
            <BookOpen className="w-5 h-5 text-brain-primary dark:text-primary-400" />
            <span className="font-semibold dark:text-gray-200">
              8 Research Papers • 4 Major Benchmarks
            </span>
            <Award className="w-5 h-5 text-brain-gold dark:text-orange-400" />
          </div>
        </div>
      </div>
    </section>
  );
}
