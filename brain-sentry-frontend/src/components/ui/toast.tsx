import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from "lucide-react";
import { cn } from "@/lib/utils";

const toastVariants = cva(
  "group pointer-events-auto relative flex w-full items-center justify-between space-x-4 overflow-hidden rounded-md border p-4 pr-8 shadow-lg transition-all",
  {
    variants: {
      variant: {
        default: "border bg-background text-foreground",
        success: "border-green-200 bg-green-50 text-green-800 dark:border-green-800 dark:bg-green-900/20 dark:text-green-200",
        error: "border-red-200 bg-red-50 text-red-800 dark:border-red-800 dark:bg-red-900/20 dark:text-red-200",
        warning: "border-yellow-200 bg-yellow-50 text-yellow-800 dark:border-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200",
        info: "border-blue-200 bg-blue-50 text-blue-800 dark:border-blue-800 dark:bg-blue-900/20 dark:text-blue-200",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface ToastProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof toastVariants> {
  title?: string;
  description?: string;
  onClose?: () => void;
  duration?: number;
}

const Toast = React.forwardRef<HTMLDivElement, ToastProps>(
  ({ className, variant, title, description, onClose, duration = 5000, ...props }, ref) => {
    const [isVisible, setIsVisible] = React.useState(true);

    React.useEffect(() => {
      if (duration > 0) {
        const timer = setTimeout(() => {
          handleClose();
        }, duration);
        return () => clearTimeout(timer);
      }
    }, [duration]);

    const handleClose = () => {
      setIsVisible(false);
      setTimeout(() => {
        onClose?.();
      }, 300); // Wait for animation
    };

    if (!isVisible) return null;

    const icons = {
      default: Info,
      success: CheckCircle,
      error: AlertCircle,
      warning: AlertTriangle,
      info: Info,
    };

    const Icon = icons[variant as keyof typeof icons] || Info;

    return (
      <div
        ref={ref}
        className={cn(
          toastVariants({ variant }),
          "animate-in slide-in-from-top-full data-[swipe=move]:translate-x-[var(--radix-toast-swipe-move-x)] data-[swipe=cancel]:reverse-translate-x-[var(--radix-toast-swipe-end-x)] data-[swipe=end]:animate-out data-[swipe=cancel]:animate-in data-[swipe=cancel]:slide-in-from-left-full",
          !isVisible && "animate-out fade-out",
          className
        )}
        {...props}
      >
        <div className="flex items-start gap-3">
          <Icon className="h-5 w-5 flex-shrink-0" />
          <div className="flex-1">
            {title && <p className="font-semibold">{title}</p>}
            {description && <p className="text-sm opacity-90">{description}</p>}
          </div>
        </div>
        <button
          onClick={handleClose}
          className="absolute right-2 top-2 rounded-md p-1 opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    );
  }
);
Toast.displayName = "Toast";

interface ToastContextValue {
  toasts: Omit<ToastProps, "onClose">[];
  addToast: (toast: Omit<ToastProps, "onClose">) => void;
  removeToast: (index: number) => void;
}

const ToastContext = React.createContext<ToastContextValue | undefined>(undefined);

interface ToasterProps {
  position?: "top-right" | "top-left" | "bottom-right" | "bottom-left" | "top-center" | "bottom-center";
  className?: string;
}

const positionClasses = {
  "top-right": "top-4 right-4",
  "top-left": "top-4 left-4",
  "bottom-right": "bottom-4 right-4",
  "bottom-left": "bottom-4 left-4",
  "top-center": "top-4 left-1/2 -translate-x-1/2",
  "bottom-center": "bottom-4 left-1/2 -translate-x-1/2",
};

function Toaster({ position = "top-right", className }: ToasterProps) {
  const context = React.useContext(ToastContext);
  if (!context) return null;

  const { toasts } = context;

  return (
    <div
      className={cn(
        "fixed z-50 flex flex-col gap-2 max-w-sm w-full",
        positionClasses[position],
        className
      )}
    >
      {toasts.map((toast, index) => (
        <Toast
          key={index}
          {...toast}
          onClose={() => context.removeToast(index)}
        />
      ))}
    </div>
  );
}

interface ToastProviderProps {
  children: React.ReactNode;
}

function ToastProvider({ children }: ToastProviderProps) {
  const [toasts, setToasts] = React.useState<Omit<ToastProps, "onClose">[]>([]);

  const addToast = React.useCallback((toast: Omit<ToastProps, "onClose">) => {
    setToasts((prev) => [...prev, toast]);
  }, []);

  const removeToast = React.useCallback((index: number) => {
    setToasts((prev) => prev.filter((_, i) => i !== index));
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
      {children}
      <Toaster />
    </ToastContext.Provider>
  );
}

// Hook for using toasts
export function useToast() {
  const context = React.useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within a ToastProvider");
  }

  return {
    toast: context.addToast,
    success: (props: Omit<ToastProps, "variant" | "onClose">) => {
      context.addToast({ ...props, variant: "success" });
    },
    error: (props: Omit<ToastProps, "variant" | "onClose">) => {
      context.addToast({ ...props, variant: "error" });
    },
    warning: (props: Omit<ToastProps, "variant" | "onClose">) => {
      context.addToast({ ...props, variant: "warning" });
    },
    info: (props: Omit<ToastProps, "variant" | "onClose">) => {
      context.addToast({ ...props, variant: "info" });
    },
  };
}

export { Toast, Toaster, ToastProvider };
