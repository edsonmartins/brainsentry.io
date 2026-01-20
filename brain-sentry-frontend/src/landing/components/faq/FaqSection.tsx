import { ChevronDown } from "lucide-react";
import { SectionHeader } from "../ui/SectionHeader";
import { faqItems } from "../../data/faq";
import { useState } from "react";
import { useLanguage } from "../../contexts/LanguageContext";

export function FaqSection() {
  const [openIndex, setOpenIndex] = useState<number | null>(null);
  const { t } = useLanguage();

  return (
    <section className="py-24 dark:bg-muted/5" id="faq">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          center
          badge={t("faq.badge")}
          title={t("faq.title")}
        />

        <div className="max-w-3xl mx-auto space-y-4">
          {faqItems.map((item, index) => (
            <div
              key={index}
              className="bg-card dark:bg-card/80 rounded-xl border-2 border-border dark:border-border/50 overflow-hidden"
            >
              <button
                className="w-full flex items-center justify-between p-6 text-left hover:bg-muted/50 dark:hover:bg-muted/30 transition-colors"
                onClick={() => setOpenIndex(openIndex === index ? null : index)}
              >
                <span className="font-semibold pr-4 dark:text-gray-200">{item.q}</span>
                <ChevronDown
                  className={`w-5 h-5 text-muted-foreground dark:text-gray-400 transition-transform flex-shrink-0 ${
                    openIndex === index ? "rotate-180" : ""
                  }`}
                />
              </button>
              {openIndex === index && (
                <div className="px-6 pb-6 pt-0 text-muted-foreground dark:text-gray-400 border-t border-border dark:border-border/40">
                  <p className="mt-4">{item.a}</p>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
