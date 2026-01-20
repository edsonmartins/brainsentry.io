import { LandingLayout } from "../components/layout/LandingLayout";
import { HeroSection } from "../components/hero/HeroSection";
import { ProblemSection } from "../components/problem/ProblemSection";
import { SolutionSection } from "../components/solution/SolutionSection";
import { HowItWorksSection } from "../components/how-it-works/HowItWorksSection";
import { FeaturesSection } from "../components/features/FeaturesSection";
import { ComparisonSection } from "../components/comparison/ComparisonSection";
import { UseCasesSection } from "../components/use-cases/UseCasesSection";
import { ArchitectureSection } from "../components/architecture/ArchitectureSection";
import { ResearchCardsSection } from "../components/research/ResearchCardsSection";
import { EvolutionTimeline } from "../components/research/EvolutionTimeline";
import { CredibilityBadges } from "../components/research/CredibilityBadges";
import { RoadmapSection } from "../components/roadmap/RoadmapSection";
import { OpenSourceSection } from "../components/open-source/OpenSourceSection";
import { FaqSection } from "../components/faq/FaqSection";
import { CTASection } from "../components/cta/CTASection";

export function LandingPage() {
  return (
    <LandingLayout>
      <HeroSection />
      <ProblemSection />
      <SolutionSection />
      <HowItWorksSection />
      <FeaturesSection />
      <ComparisonSection />
      <UseCasesSection />
      <ArchitectureSection />
      <ResearchCardsSection />
      <CredibilityBadges />
      <EvolutionTimeline />
      <RoadmapSection />
      <OpenSourceSection />
      <FaqSection />
      <CTASection />
    </LandingLayout>
  );
}
