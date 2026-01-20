import { ReactNode } from "react";
import { LandingHeader } from "./LandingHeader";
import { LandingFooter } from "./LandingFooter";
import { LanguageProvider } from "../../contexts/LanguageContext";

interface LandingLayoutProps {
  children: ReactNode;
}

export function LandingLayout({ children }: LandingLayoutProps) {
  return (
    <LanguageProvider>
      <div className="min-h-screen bg-background">
        <LandingHeader />
        <main className="pt-16">{children}</main>
        <LandingFooter />
      </div>
    </LanguageProvider>
  );
}
