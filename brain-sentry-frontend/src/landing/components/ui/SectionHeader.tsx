import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface SectionHeaderProps {
  badge?: string;
  title: string | ReactNode;
  description?: string;
  className?: string;
  center?: boolean;
}

export function SectionHeader({
  badge,
  title,
  description,
  className,
  center = false,
}: SectionHeaderProps) {
  return (
    <div className={cn("mb-12", center && "text-center", className)}>
      {badge && (
        <span className="inline-block px-3 py-1 mb-4 text-xs font-semibold tracking-wider text-brain-primary uppercase bg-orange-50 dark:bg-primary-950/50 dark:text-primary-300 rounded-full border border-orange-100 dark:border-primary-900">
          {badge}
        </span>
      )}
      <h2 className="text-3xl font-bold tracking-tight sm:text-4xl lg:text-5xl dark:text-white">
        {title}
      </h2>
      {description && (
        <p className={cn(
          "mt-4 text-lg text-muted-foreground dark:text-gray-400 max-w-3xl",
          center && "mx-auto"
        )}>
          {description}
        </p>
      )}
    </div>
  );
}
