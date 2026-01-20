import { cn } from "@/lib/utils";

interface GradientTextProps {
  children: React.ReactNode;
  from?: string;
  to?: string;
  className?: string;
}

export function GradientText({
  children,
  from = "#E67E50", // Laranja principal
  to = "#F59E0B",   // Laranja vibrante
  className,
}: GradientTextProps) {
  return (
    <span
      className={cn("inline-block", className)}
      style={{
        background: `linear-gradient(135deg, ${from} 0%, ${to} 100%)`,
        WebkitBackgroundClip: "text",
        WebkitTextFillColor: "transparent",
        backgroundClip: "text",
      }}
    >
      {children}
    </span>
  );
}
