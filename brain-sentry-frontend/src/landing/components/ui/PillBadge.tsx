import { cn } from "@/lib/utils";

interface PillBadgeProps {
  children: React.ReactNode;
  variant?: "primary" | "accent" | "success" | "gold" | "gray";
  className?: string;
}

export function PillBadge({
  children,
  variant = "primary",
  className,
}: PillBadgeProps) {
  const variants = {
    primary: "bg-orange-50 text-orange-700 border-orange-200 dark:bg-orange-950/40 dark:text-orange-300 dark:border-orange-800/60",
    accent: "bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800/60",
    success: "bg-green-50 text-green-700 border-green-200 dark:bg-green-950/40 dark:text-green-300 dark:border-green-800/60",
    gold: "bg-gradient-to-r from-amber-500 to-red-500 text-white border-transparent",
    gray: "bg-gray-100 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center px-3 py-1 rounded-full text-sm font-medium border",
        variants[variant],
        className
      )}
    >
      {children}
    </span>
  );
}
