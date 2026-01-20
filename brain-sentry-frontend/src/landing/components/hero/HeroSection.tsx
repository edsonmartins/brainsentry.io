import { ArrowRight, Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { GradientText } from "../ui/GradientText";
import { TrustBadge } from "../ui/TrustBadge";
import { PillBadge } from "../ui/PillBadge";
import { useLanguage } from "../../contexts/LanguageContext";

export function HeroSection() {
  const { t } = useLanguage();

  const scrollToCTA = () => {
    document.getElementById("cta")?.scrollIntoView({ behavior: "smooth" });
  };

  return (
    <section className="relative overflow-hidden bg-gradient-to-b from-orange-50/50 via-background to-background dark:from-orange-950/30 dark:via-background dark:to-background">
      {/* Background decorative elements */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-orange-500/10 dark:bg-primary-500/20 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-amber-500/10 dark:bg-accent-500/20 rounded-full blur-3xl" />
      </div>

      <div className="container mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-32 relative">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Left Column - Text Content */}
          <div className="text-left">
            {/* Badge */}
            <div className="flex justify-start mb-8 animate-fade-in">
              <PillBadge variant="primary">
                {t("hero.tagline")}
              </PillBadge>
            </div>

            {/* Headline */}
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight mb-6 animate-slide-up dark:text-white">
              {t("hero.title")}{" "}
              <GradientText from="#E67E50" to="#F59E0B">
                Second Brain
              </GradientText>
            </h1>

            {/* Subheadline */}
            <p className="text-xl sm:text-2xl text-muted-foreground dark:text-gray-300 mb-6 max-w-xl animate-slide-up" style={{ animationDelay: "0.1s" }}>
              {t("hero.subtitle1")}
            </p>

            <p className="text-lg text-muted-foreground dark:text-gray-400 mb-10 max-w-xl animate-slide-up" style={{ animationDelay: "0.2s" }}>
              {t("hero.subtitle2")}
              <br className="hidden sm:inline" />
              {" "}{t("hero.subtitle3")}
            </p>

            {/* CTAs */}
            <div className="flex flex-col sm:flex-row items-start gap-4 mb-12 animate-slide-up" style={{ animationDelay: "0.3s" }}>
              <Button size="lg" className="h-12 px-8 text-base bg-brain-gold hover:bg-brain-gold/90" onClick={scrollToCTA}>
                {t("hero.cta.primary")}
                <ArrowRight className="ml-2 w-4 h-4" />
              </Button>
              <Button size="lg" variant="outline" className="h-12 px-8 text-base dark:border-border/60 dark:hover:bg-muted/20" asChild>
                <a href="https://github.com/edsonmartins/brainsentry.io" target="_blank" rel="noopener noreferrer">
                  <Github className="mr-2 w-4 h-4" />
                  {t("hero.cta.secondary")}
                </a>
              </Button>
            </div>

            {/* Trust Badges */}
            <div className="flex flex-wrap items-start gap-6 animate-slide-up" style={{ animationDelay: "0.4s" }}>
              <TrustBadge>{t("hero.trust1")}</TrustBadge>
              <TrustBadge>{t("hero.trust2")}</TrustBadge>
              <TrustBadge>{t("hero.trust3")}</TrustBadge>
              <TrustBadge>{t("hero.trust4")}</TrustBadge>
            </div>

            {/* Stats */}
            <div className="mt-12 grid grid-cols-3 gap-6 max-w-lg animate-slide-up" style={{ animationDelay: "0.5s" }}>
              <div className="text-left">
                <div className="text-3xl sm:text-4xl font-bold text-brain-primary dark:text-primary-400">4</div>
                <div className="text-sm text-muted-foreground dark:text-gray-400 mt-1">{t("hero.stat1.label")}</div>
              </div>
              <div className="text-left">
                <div className="text-3xl sm:text-4xl font-bold text-brain-accent dark:text-accent-400">100%</div>
                <div className="text-sm text-muted-foreground dark:text-gray-400 mt-1">{t("hero.stat2.label")}</div>
              </div>
              <div className="text-left">
                <div className="text-3xl sm:text-4xl font-bold text-brain-success dark:text-emerald-400">&lt;500ms</div>
                <div className="text-sm text-muted-foreground dark:text-gray-400 mt-1">{t("hero.stat3.label")}</div>
              </div>
            </div>
          </div>

          {/* Right Column - Video */}
          <div className="flex justify-center lg:justify-end animate-fade-in">
            <div className="bg-gray-900 dark:bg-gray-950 rounded-2xl p-3 shadow-2xl">
              <video
                src="/videos/brainsentry.mp4"
                autoPlay
                loop
                muted
                playsInline
                className="w-full max-w-md h-auto rounded-xl"
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
