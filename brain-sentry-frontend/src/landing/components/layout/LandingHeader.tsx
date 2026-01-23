import { useState, useEffect } from "react";
import { Menu, X, Sun, Moon, Globe } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { useLanguage } from "../../contexts/LanguageContext";

export function LandingHeader() {
  const navigate = useNavigate();
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isDark, setIsDark] = useState(false);
  const { language, setLanguage, t } = useLanguage();

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 10);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    const isDarkMode = document.documentElement.classList.contains("dark");
    setIsDark(isDarkMode);
  }, []);

  const toggleTheme = () => {
    const newIsDark = !isDark;
    setIsDark(newIsDark);
    if (newIsDark) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  };

  const languages = [
    { code: "en" as const, label: "EN", flag: "🇺🇸" },
    { code: "pt" as const, label: "PT", flag: "🇧🇷" },
    { code: "es" as const, label: "ES", flag: "🇪🇸" },
  ];

  const navLinks = [
    { name: t("nav.features"), href: "#features" },
    { name: t("nav.research"), href: "#research" },
    { name: t("nav.comparison"), href: "#comparison" },
    { name: t("nav.faq"), href: "#faq" },
  ];

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        isScrolled
          ? "bg-background/95 dark:bg-background/90 backdrop-blur-lg border-b border-border dark:border-border/60 shadow-sm"
          : "bg-transparent"
      }`}
    >
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <div className="flex items-center">
            <a href="#" className="flex items-center gap-2">
              <div className="flex items-center">
                <img src="/images/brainsentry_logo_branco.png" alt="Brain Sentry" className="h-10 w-auto dark:hidden rounded-xl" />
                <div className="hidden dark:flex bg-gray-950 rounded-xl p-2 items-center shadow-md border border-gray-800">
                  <img src="/images/brainsentry_logo.png" alt="Brain Sentry" className="h-10 w-auto" />
                </div>
              </div>
              <span className="font-bold text-xl text-gray-900 dark:text-white">Brain Sentry</span>
            </a>
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center gap-8">
            {navLinks.map((link) => (
              <a
                key={link.name}
                href={link.href}
                className="text-sm font-medium text-muted-foreground dark:text-gray-300 hover:text-foreground dark:hover:text-white transition-colors"
              >
                {link.name}
              </a>
            ))}
          </nav>

          {/* Desktop Controls */}
          <div className="hidden md:flex items-center gap-3">
            {/* Language Switcher */}
            <div className="flex items-center gap-1 bg-muted/50 dark:bg-muted/20 rounded-lg p-1 border border-border dark:border-border/40">
              {languages.map((lang) => (
                <button
                  key={lang.code}
                  onClick={() => setLanguage(lang.code)}
                  className={`px-2 py-1 rounded-md text-xs font-medium transition-all ${
                    language === lang.code
                      ? "bg-background dark:bg-background shadow-sm text-foreground dark:text-white"
                      : "text-muted-foreground dark:text-gray-400 hover:text-foreground dark:hover:text-white"
                  }`}
                  title={lang.label}
                >
                  {lang.flag} {lang.label}
                </button>
              ))}
            </div>

            {/* Theme Toggle */}
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg hover:bg-muted/50 dark:hover:bg-muted/20 transition-colors border border-border dark:border-border/40"
              title={isDark ? "Switch to Light Mode" : "Switch to Dark Mode"}
            >
              {isDark ? (
                <Sun className="w-4 h-4 text-yellow-500" />
              ) : (
                <Moon className="w-4 h-4 text-slate-600" />
              )}
            </button>

            <Button variant="ghost" size="sm" className="dark:text-gray-300 dark:hover:text-white" onClick={() => navigate("/login")}>
              {t("nav.signIn")}
            </Button>
            <Button size="sm" className="bg-brain-primary hover:bg-brain-primary/90" onClick={() => navigate("/login")}>
              {t("nav.getStarted")}
            </Button>
          </div>

          {/* Mobile Menu Button */}
          <button
            className="md:hidden p-2 dark:text-white"
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
          >
            {isMobileMenuOpen ? (
              <X className="w-6 h-6" />
            ) : (
              <Menu className="w-6 h-6" />
            )}
          </button>
        </div>

        {/* Mobile Menu */}
        {isMobileMenuOpen && (
          <div className="md:hidden py-4 border-t border-border dark:border-border/60 bg-background/95 dark:bg-background/90 backdrop-blur-lg">
            <nav className="flex flex-col gap-4">
              {navLinks.map((link) => (
                <a
                  key={link.name}
                  href={link.href}
                  className="text-sm font-medium text-muted-foreground dark:text-gray-300 hover:text-foreground dark:hover:text-white transition-colors"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  {link.name}
                </a>
              ))}

              {/* Mobile Language & Theme Controls */}
              <div className="flex items-center justify-between py-3 border-t border-border dark:border-border/40 mt-2">
                {/* Language Switcher */}
                <div className="flex items-center gap-2">
                  <Globe className="w-4 h-4 text-muted-foreground dark:text-gray-400" />
                  <div className="flex items-center gap-1 bg-muted/50 dark:bg-muted/20 rounded-lg p-1 border border-border dark:border-border/40">
                    {languages.map((lang) => (
                      <button
                        key={lang.code}
                        onClick={() => setLanguage(lang.code)}
                        className={`px-2 py-1 rounded-md text-xs font-medium transition-all ${
                          language === lang.code
                            ? "bg-background dark:bg-background shadow-sm text-foreground dark:text-white"
                            : "text-muted-foreground dark:text-gray-400"
                        }`}
                      >
                        {lang.label}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Theme Toggle */}
                <button
                  onClick={toggleTheme}
                  className="p-2 rounded-lg hover:bg-muted/50 dark:hover:bg-muted/20 transition-colors border border-border dark:border-border/40"
                >
                  {isDark ? (
                    <Sun className="w-4 h-4 text-yellow-500" />
                  ) : (
                    <Moon className="w-4 h-4 text-slate-600" />
                  )}
                </button>
              </div>

              <div className="flex flex-col gap-2 mt-4">
                <Button variant="ghost" size="sm" className="justify-start dark:text-gray-300 dark:hover:text-white" onClick={() => navigate("/login")}>
                  {t("nav.signIn")}
                </Button>
                <Button size="sm" className="bg-brain-primary hover:bg-brain-primary/90" onClick={() => navigate("/login")}>
                  {t("nav.getStarted")}
                </Button>
              </div>
            </nav>
          </div>
        )}
      </div>
    </header>
  );
}
