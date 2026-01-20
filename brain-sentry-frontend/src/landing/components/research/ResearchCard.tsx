import { BookOpen, ExternalLink } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ResearchPaper } from "../../data/researchPapers";
import { useLanguage } from "../../contexts/LanguageContext";

interface ResearchCardProps {
  paper: ResearchPaper;
  featured?: boolean;
}

const getAlignmentColor = (score: number) => {
  if (score >= 90) return "from-green-500 to-green-600";
  if (score >= 75) return "from-amber-500 to-amber-600";
  return "from-gray-400 to-gray-500";
};

const getAlignmentLabel = (score: number) => {
  if (score >= 90) return "EXCELLENT";
  if (score >= 75) return "GOOD";
  return "MODERATE";
};

export function ResearchCard({ paper, featured = false }: ResearchCardProps) {
  const { t } = useLanguage();

  return (
    <div
      className={cn(
        "relative bg-gradient-to-br from-gray-50 to-white dark:from-gray-900/80 dark:to-gray-800/80 rounded-2xl border-2 border-border dark:border-border/50 p-6 transition-all duration-300 hover:shadow-xl dark:hover:shadow-xl/50 hover:-translate-y-1",
        featured
          ? "border-brain-primary dark:border-primary-500/60 shadow-lg dark:shadow-primary-500/20"
          : ""
      )}
    >
      {/* Featured gradient top bar */}
      {featured && (
        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-brain-primary to-brain-accent rounded-t-2xl" />
      )}

      {/* Badge */}
      {paper.badge && (
        <div
          className={cn(
            "inline-block px-3 py-1 text-xs font-bold rounded-full mb-4 border border-transparent",
            featured
              ? "bg-gradient-to-r from-amber-500 to-red-500 text-white"
              : "bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 border-gray-200 dark:border-gray-700"
          )}
        >
          {paper.badge}
        </div>
      )}

      {/* Title */}
      <h3 className="text-lg font-bold mb-2 line-clamp-2 dark:text-white">{paper.title}</h3>

      {/* Meta */}
      <div className="mb-4 p-3 bg-muted/50 dark:bg-muted/30 rounded-lg border border-border dark:border-border/40">
        <div className="text-xs font-semibold text-muted-foreground dark:text-gray-400 mb-1">
          {paper.institution}
        </div>
        <div className="text-xs text-muted-foreground dark:text-gray-500">
          {paper.authors.join(" • ")}
        </div>
        <div className="text-xs text-muted-foreground dark:text-gray-500 mt-1">
          {paper.date} • {paper.source}
          {paper.arxiv && (
            <span className="ml-2 font-mono text-brain-primary dark:text-primary-400">arXiv:{paper.arxiv}</span>
          )}
        </div>
      </div>

      {/* Key Results */}
      <div className="mb-4">
        <div className="text-xs font-semibold text-muted-foreground dark:text-gray-400 mb-2">
          {t("research.keyFindings")}
        </div>
        <ul className="space-y-1">
          {paper.keyResults.slice(0, 3).map((result, index) => (
            <li key={index} className="text-xs flex items-start gap-2">
              <span className="text-brain-primary dark:text-primary-400 mt-0.5">•</span>
              <span className="text-muted-foreground dark:text-gray-400">{result}</span>
            </li>
          ))}
        </ul>
      </div>

      {/* Alignment Score */}
      <div className="mb-4">
        <div
          className={cn(
            "inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-white text-sm font-bold",
            "bg-gradient-to-r",
            getAlignmentColor(paper.alignmentScore)
          )}
        >
          <span>{t("research.alignment")}: {paper.alignmentScore}%</span>
        </div>
      </div>

      {/* What We Learned */}
      {paper.insights.length > 0 && (
        <div className="mb-4">
          <div className="text-xs font-semibold text-muted-foreground dark:text-gray-400 mb-2">
            {t("research.impact")}
          </div>
          <ul className="space-y-1">
            {paper.insights.map((insight, index) => (
              <li key={index} className="text-xs flex items-start gap-2">
                <span className="text-green-500 dark:text-green-400 mt-0.5">✓</span>
                <span className="text-muted-foreground dark:text-gray-400">{insight}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Implemented Features */}
      {paper.implemented && paper.implemented.length > 0 && (
        <div className="mb-4">
          <div className="text-xs font-semibold text-muted-foreground dark:text-gray-400 mb-2">
            {t("research.implemented")}
          </div>
          <ul className="space-y-1">
            {paper.implemented.map((item, index) => (
              <li key={index} className="text-xs flex items-start gap-2">
                <span className="text-brain-accent dark:text-accent-400 mt-0.5">•</span>
                <span className="text-muted-foreground dark:text-gray-400">{item}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Benchmark */}
      {paper.benchmark && (
        <div className="mb-4 p-3 bg-brain-primary/5 dark:bg-brain-primary/10 rounded-lg border-2 border-brain-primary/20 dark:border-brain-primary/40">
          <div className="text-xs font-semibold text-brain-primary dark:text-primary-300 mb-2">
            {t("research.benchmark")}
          </div>
          <div className="space-y-1">
            <div className="flex justify-between text-xs">
              <span className="text-muted-foreground dark:text-gray-400">{t("research.benchmark.confucius")}:</span>
              <span className="font-bold dark:text-white">{paper.benchmark.confucius}</span>
            </div>
            <div className="flex justify-between text-xs">
              <span className="text-muted-foreground dark:text-gray-400">{t("research.benchmark.anthropic")}:</span>
              <span className="font-bold dark:text-white">{paper.benchmark.anthropic}</span>
            </div>
            <div className="flex justify-between text-xs">
              <span className="text-muted-foreground dark:text-gray-400">{t("research.benchmark.target")}:</span>
              <span className="font-bold text-brain-gold dark:text-gold-400">
                {paper.benchmark.brainSentryTarget} 🎯
              </span>
            </div>
          </div>
        </div>
      )}

      {/* Brain Sentry Fills Gaps */}
      {paper.brainSentryFills && (
        <div className="mb-4">
          <div className="text-xs font-semibold text-red-500 dark:text-red-400 mb-2">
            {t("research.gaps")}
          </div>
          <ul className="space-y-1 mb-3">
            <li className="text-xs text-muted-foreground dark:text-gray-400">
              ❌ {paper.gaps?.mostSystemsLack}
            </li>
            <li className="text-xs text-muted-foreground dark:text-gray-400">
              ❌ {paper.gaps?.fewHave}
            </li>
          </ul>
          <div className="text-xs font-semibold text-green-600 dark:text-green-400 mb-2">
            {t("research.fills")}
          </div>
          <ul className="space-y-1">
            {paper.brainSentryFills.map((item, index) => (
              <li key={index} className="text-xs flex items-start gap-2">
                <span className="text-green-500 dark:text-green-400">✅</span>
                <span className="text-muted-foreground dark:text-gray-400">{item}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Limitations & Evolution */}
      {paper.limitationsIdentified && (
        <div className="mb-4">
          <div className="text-xs font-semibold text-amber-600 dark:text-gold-400 mb-2">
            {t("research.limitations")}
          </div>
          <ul className="space-y-1 mb-3">
            {paper.limitationsIdentified.map((limitation, index) => (
              <li key={index} className="text-xs text-muted-foreground dark:text-gray-400">
                {limitation}
              </li>
            ))}
          </ul>
          <div className="text-xs font-semibold text-brain-primary dark:text-primary-300 mb-2">
            {t("research.evolution")}
          </div>
          <ul className="space-y-1">
            {paper.brainSentryEvolution!.map((item, index) => (
              <li key={index} className="text-xs flex items-start gap-2">
                <span className="text-brain-primary dark:text-primary-400">•</span>
                <span className="text-muted-foreground dark:text-gray-400">{item}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Link */}
      {paper.arxivUrl && (
        <a
          href={paper.arxivUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-2 text-sm text-brain-primary dark:text-primary-400 hover:underline mt-2"
        >
          <BookOpen className="w-4 h-4" />
          {t("research.viewPaper")}
          <ExternalLink className="w-3 h-3" />
        </a>
      )}
    </div>
  );
}
