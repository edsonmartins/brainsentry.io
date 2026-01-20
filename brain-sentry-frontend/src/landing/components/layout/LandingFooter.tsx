import { Github, Twitter, MessageSquare } from "lucide-react";
import { useLanguage } from "../../contexts/LanguageContext";

export function LandingFooter() {
  const { t } = useLanguage();

  const productLinks = [
    { name: t("footer.link.features"), href: "#features" },
    { name: t("footer.link.research"), href: "#research" },
    { name: t("footer.link.roadmap"), href: "#" },
  ];

  const resourceLinks = [
    { name: t("footer.link.docs"), href: "#" },
    { name: t("footer.link.github"), href: "https://github.com/edsonmartins/brainsentry.io" },
    { name: t("footer.link.papers"), href: "#research" },
  ];

  const companyLinks = [
    { name: t("footer.link.about"), href: "#" },
    { name: t("footer.link.blog"), href: "#" },
    { name: t("footer.link.contact"), href: "#" },
  ];

  return (
    <footer className="border-t border-border dark:border-border/60 bg-muted/30 dark:bg-muted/10">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
          {/* Logo & Description */}
          <div className="col-span-2">
            <div className="flex items-center gap-2 mb-4">
              <img src="/images/brainsentry_logo_branco.png" alt="Brain Sentry" className="h-8 w-auto dark:hidden rounded-xl" />
              <div className="hidden dark:flex items-center gap-2">
                <div className="bg-gray-950 rounded-xl p-2 shadow-md border border-gray-800 inline-flex">
                  <img src="/images/brainsentry_logo.png" alt="Brain Sentry" className="h-8 w-auto" />
                </div>
                <span className="font-bold text-xl dark:text-white">Brain Sentry</span>
              </div>
              <span className="font-bold text-xl text-gray-900 dark:hidden">Brain Sentry</span>
            </div>
            <p className="text-sm text-muted-foreground dark:text-gray-400 mb-4 max-w-xs">
              {t("footer.description")}
            </p>
            <div className="flex gap-4">
              <a
                href="https://github.com/edsonmartins/brainsentry.io"
                className="text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
              >
                <Github className="w-5 h-5" />
              </a>
              <a
                href="#"
                className="text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
              >
                <Twitter className="w-5 h-5" />
              </a>
              <a
                href="#"
                className="text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
              >
                <MessageSquare className="w-5 h-5" />
              </a>
            </div>
          </div>

          {/* Product Links */}
          <div>
            <h3 className="font-semibold mb-4 dark:text-gray-200">{t("footer.product")}</h3>
            <ul className="space-y-2">
              {productLinks.map((link) => (
                <li key={link.name}>
                  <a
                    href={link.href}
                    className="text-sm text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
                  >
                    {link.name}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Resources & Company */}
          <div>
            <h3 className="font-semibold mb-4 dark:text-gray-200">{t("footer.resources")}</h3>
            <ul className="space-y-2">
              {resourceLinks.map((link) => (
                <li key={link.name}>
                  <a
                    href={link.href}
                    className="text-sm text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
                  >
                    {link.name}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h3 className="font-semibold mb-4 dark:text-gray-200">{t("footer.company")}</h3>
            <ul className="space-y-2">
              {companyLinks.map((link) => (
                <li key={link.name}>
                  <a
                    href={link.href}
                    className="text-sm text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
                  >
                    {link.name}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="mt-12 pt-8 border-t border-border dark:border-border/40">
          <div className="flex flex-col md:flex-row justify-between items-center gap-4">
            <p className="text-sm text-muted-foreground dark:text-gray-400">
              {t("footer.copyright")}
            </p>
            <div className="flex gap-6">
              <a
                href="#"
                className="text-sm text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
              >
                {t("footer.privacy")}
              </a>
              <a
                href="#"
                className="text-sm text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
              >
                {t("footer.terms")}
              </a>
              <a
                href="#"
                className="text-sm text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white transition-colors"
              >
                {t("footer.license")}
              </a>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}
