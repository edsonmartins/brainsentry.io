import { Check } from "lucide-react";
import { cn } from "@/lib/utils";

interface TrustBadgeProps {
  children: React.ReactNode;
  className?: string;
}

export function TrustBadge({ children, className }: TrustBadgeProps) {
  return (
    <div
      className={cn(
        "flex items-center gap-2 text-sm text-muted-foreground",
        className
      )}
    >
      <Check className="w-4 h-4 text-brain-success" />
      <span>{children}</span>
    </div>
  );
}
