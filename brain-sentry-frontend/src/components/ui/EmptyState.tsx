import { type LucideIcon } from "lucide-react";
import { Button } from "./button";

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description: string;
  action?: {
    label: string;
    onClick: () => void;
  };
  className?: string;
}

export function EmptyState({ icon: Icon, title, description, action, className = "" }: EmptyStateProps) {
  return (
    <div className={`flex flex-col items-center justify-center py-12 px-4 ${className}`}>
      <div className="p-4 rounded-full bg-muted/50 mb-4">
        <Icon className="h-10 w-10 text-muted-foreground/40" />
      </div>
      <h3 className="text-sm font-semibold text-foreground mb-1">{title}</h3>
      <p className="text-xs text-muted-foreground text-center max-w-xs mb-4">{description}</p>
      {action && (
        <Button
          size="sm"
          className="bg-gradient-to-r from-brain-primary to-brain-accent text-white"
          onClick={action.onClick}
        >
          {action.label}
        </Button>
      )}
    </div>
  );
}
