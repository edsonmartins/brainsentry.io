import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const spinnerVariants = cva(
  "inline-block animate-spin rounded-full border-solid",
  {
    variants: {
      size: {
        xs: "h-3 w-3 border-2",
        sm: "h-4 w-4 border-2",
        md: "h-6 w-6 border-2",
        lg: "h-8 w-8 border-3",
        xl: "h-12 w-12 border-4",
      },
      variant: {
        default: "border-primary border-t-transparent",
        muted: "border-muted-foreground/30 border-t-transparent",
        white: "border-white/30 border-t-transparent",
      },
    },
    defaultVariants: {
      size: "md",
      variant: "default",
    },
  }
);

export interface SpinnerProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof spinnerVariants> {
  label?: string;
}

const Spinner = React.forwardRef<HTMLDivElement, SpinnerProps>(
  ({ className, size, variant, label, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn("flex flex-col items-center justify-center gap-2", className)}
        {...props}
      >
        <div className={cn(spinnerVariants({ size, variant }))} />
        {label && <p className="text-sm text-muted-foreground">{label}</p>}
      </div>
    );
  }
);
Spinner.displayName = "Spinner";

// Full page loading overlay
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
      )
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

const skeletonVariants = cva(
  "animate-pulse bg-muted",
  {
    variants: {
      variant: {
        text: "h-4 w-full rounded",
        circular: "rounded-full",
        rectangular: "rounded-md",
        rounded: "rounded-lg",
      },
    },
    defaultVariants: {
      variant: "rectangular",
    },
  }
);

function Skeleton({
  variant,
  width,
  height,
  className,
  ...props
}: SkeletonProps) {
  return (
    <div
      className={cn(skeletonVariants({ variant }), className)}
      style={{ width, height }}
      {...props}
    />
  );
}

// Card skeleton for loading cards
function CardSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("rounded-lg border bg-card p-6 space-y-4", className)}>
      <Skeleton variant="text" width="60%" />
      <Skeleton variant="text" className="h-16" />
      <div className="flex gap-2">
        <Skeleton variant="text" width="20%" />
        <Skeleton variant="text" width="20%" />
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
          <Skeleton key={i} variant="text" width={`${100 / columns}%`} />
        ))}
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex gap-4 p-2">
          {Array.from({ length: columns }).map((_, j) => (
            <Skeleton key={j} variant="text" width={`${100 / columns}%`} />
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
