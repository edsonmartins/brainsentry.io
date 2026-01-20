import { Github, MessageSquare, Users } from "lucide-react";
import { SectionHeader } from "../ui/SectionHeader";
import { PillBadge } from "../ui/PillBadge";
import { useLanguage } from "../../contexts/LanguageContext";

export function OpenSourceSection() {
  const { t } = useLanguage();

  return (
    <section className="py-24 bg-muted/30 dark:bg-muted/10">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center">
          <SectionHeader
            center
            badge={t("opensource.badge")}
            title={t("opensource.title")}
          />

          {/* Main Badge */}
          <div className="inline-flex items-center gap-2 mb-8">
            <span className="text-4xl">🌟</span>
            <span className="text-2xl font-bold dark:text-white">OPEN SOURCE & COMMUNITY-DRIVEN</span>
          </div>

          {/* License */}
          <div className="mb-8">
            <p className="text-muted-foreground dark:text-gray-400 mb-4">
              {t("opensource.license")}{" "}
              <span className="font-semibold text-foreground dark:text-gray-200">Apache 2.0</span>
            </p>
            <p className="text-sm text-muted-foreground dark:text-gray-400">
              {t("opensource.repo")}{" "}
              <a
                href="https://github.com/edsonmartins/brainsentry.io"
                target="_blank"
                rel="noopener noreferrer"
                className="text-brain-primary dark:text-primary-400 hover:underline"
              >
                github.com/edsonmartins/brainsentry.io
              </a>
            </p>
          </div>

          {/* Values */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            <div className="flex flex-col items-center p-4 bg-card dark:bg-card/80 rounded-lg border-2 border-border dark:border-border/50">
              <div className="w-10 h-10 rounded-full bg-orange-100 dark:bg-primary-900/30 flex items-center justify-center mb-2">
                <span className="text-xl">📖</span>
              </div>
              <span className="text-sm font-medium dark:text-gray-200">{t("opensource.value1")}</span>
            </div>
            <div className="flex flex-col items-center p-4 bg-card dark:bg-card/80 rounded-lg border-2 border-border dark:border-border/50">
              <div className="w-10 h-10 rounded-full bg-amber-100 dark:bg-accent-900/30 flex items-center justify-center mb-2">
                <span className="text-xl">🤝</span>
              </div>
              <span className="text-sm font-medium dark:text-gray-200">{t("opensource.value2")}</span>
            </div>
            <div className="flex flex-col items-center p-4 bg-card dark:bg-card/80 rounded-lg border-2 border-border dark:border-border/50">
              <div className="w-10 h-10 rounded-full bg-yellow-100 dark:bg-gold-900/30 flex items-center justify-center mb-2">
                <span className="text-xl">🔧</span>
              </div>
              <span className="text-sm font-medium dark:text-gray-200">{t("opensource.value3")}</span>
            </div>
            <div className="flex flex-col items-center p-4 bg-card dark:bg-card/80 rounded-lg border-2 border-border dark:border-border/50">
              <div className="w-10 h-10 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center mb-2">
                <span className="text-xl">🔓</span>
              </div>
              <span className="text-sm font-medium dark:text-gray-200">{t("opensource.value4")}</span>
            </div>
          </div>

          {/* CTAs */}
          <div className="flex flex-wrap justify-center gap-4">
            <a
              href="https://github.com/edsonmartins/brainsentry.io"
              target="_blank"
              rel="noopener noreferrer"
            >
              <PillBadge variant="primary" className="cursor-pointer hover:opacity-80 dark:border-border/60">
                <Github className="w-4 h-4 inline mr-2" />
                {t("opensource.github")}
              </PillBadge>
            </a>
            <PillBadge variant="accent" className="cursor-pointer dark:border-border/60">
              <MessageSquare className="w-4 h-4 inline mr-2" />
              {t("opensource.discord")}
            </PillBadge>
            <PillBadge variant="gold" className="cursor-pointer dark:border-border/60">
              <Users className="w-4 h-4 inline mr-2" />
              {t("opensource.docs")}
            </PillBadge>
          </div>
        </div>
      </div>
    </section>
  );
}
