import { cn } from "@/lib/utils";
import { Loader2 } from "lucide-react";

interface SpinnerProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: "sm" | "md" | "lg";
  label?: string;
  className?: string;
}

const sizeClasses = {
  sm: "h-4 w-4",
  md: "h-6 w-6",
  lg: "h-8 w-8",
};

function Spinner({ size = "md", label, className }: SpinnerProps) {
  return (
    <div className="flex items-center gap-2">
      <Loader2 className={`animate-spin ${sizeClasses[size]} ${className}`} />
      {label && <span className="text-sm text-muted-foreground">{label}</span>}
    </div>
  );
}

interface LoadingOverlayProps {
  isLoading: boolean;
  label?: string;
  className?: string;
}

function LoadingOverlay({ isLoading, label = "Carregando...", className }: LoadingOverlayProps) {
  if (!isLoading) return null;

  return (
    <div
      className={cn(
        "fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm",
        className
      )}
    >
      <Spinner size="lg" label={label} />
    </div>
  );
}

// Inline loading for buttons or small areas
interface InlineLoadingProps {
  isLoading: boolean;
  children: React.ReactNode;
  className?: string;
}

function InlineLoading({ isLoading, children, className }: InlineLoadingProps) {
  return (
    <div className={cn("flex items-center gap-2", className)}>
      {isLoading ? <Spinner size="sm" /> : null}
      {children}
    </div>
  );
}

// Skeleton loading placeholder
interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "text" | "circular" | "rectangular" | "rounded";
  width?: string | number;
  height?: string | number;
  className?: string;
}

const skeletonVariants = {
  text: "h-4 w-full rounded",
  circular: "rounded-full",
  rectangular: "rounded-md",
  rounded: "rounded-lg",
};

function Skeleton({
  variant = "rectangular",
  width,
  height,
  className,
  ...props
}: SkeletonProps) {
  return (
    <div
      className={cn("animate-pulse bg-muted", skeletonVariants[variant], className)}
      style={{ width, height }}
      {...props}
    />
  );
}

// Card skeleton for loading cards
function CardSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("rounded-lg border bg-card p-6 space-y-4", className)}>
      <Skeleton width="60%" />
      <Skeleton className="h-16" />
      <div className="flex gap-2">
        <Skeleton width="20%" />
        <Skeleton width="20%" />
      </div>
    </div>
  );
}

// Table skeleton for loading tables
function TableSkeleton({ rows = 5, columns = 4, className }: { rows?: number; columns?: number; className?: string }) {
  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex gap-4 p-2 border-b">
        {Array.from({ length: columns }).map((_, i) => (
          <Skeleton key={i} width={`${100 / columns}%`} />
        ))}
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex gap-4 p-2">
          {Array.from({ length: columns }).map((_, j) => (
            <Skeleton key={j} width={`${100 / columns}%`} />
          ))}
        </div>
      ))}
    </div>
  );
}

export {
  Spinner,
  LoadingOverlay,
  InlineLoading,
  Skeleton,
  CardSkeleton,
  TableSkeleton,
};
