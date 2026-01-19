import { useState } from "react";
import * as React from "react";
import { cn } from "@/lib/utils";
import { AlertCircle, RefreshCcw, Home } from "lucide-react";
import { Button } from "./button";
import { Card, CardContent, CardHeader, CardTitle } from "./card";

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ComponentInfo | null;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.Component<{ error: Error; resetErrorBoundary: () => void }>;
  onError?: (error: Error, errorInfo: React.ComponentInfo) => void;
}

export class ErrorBoundary extends React.Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return {
      hasError: true,
      error,
      errorInfo: null,
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ComponentInfo) {
    this.setState({
      error,
      errorInfo,
    });

    // Log error to console (could be sent to logging service)
    console.error("ErrorBoundary caught an error:", error, errorInfo);

    // Call custom error handler
    this.props.onError?.(error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render() {
    if (this.state.hasError) {
      // Use custom fallback if provided
      if (this.props.fallback) {
        const FallbackComponent = this.props.fallback;
        return (
          <FallbackComponent
            error={this.state.error!}
            resetErrorBoundary={this.handleReset}
          />
        );
      }

      // Default error UI
      return (
        <div className="min-h-screen bg-background flex items-center justify-center p-4">
          <Card className="max-w-lg w-full">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-destructive/10 rounded-full">
                  <AlertCircle className="h-6 w-6 text-destructive" />
                </div>
                <CardTitle>Algo deu errado</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-muted-foreground">
                Ocorreu um erro inesperado. Por favor, tente recarregar a página.
              </p>

              {process.env.NODE_ENV === "development" && (
                <details className="p-4 bg-muted rounded-lg">
                  <summary className="cursor-pointer font-mono text-sm">
                    Detalhes do erro
                  </summary>
                  <pre className="mt-4 text-xs overflow-auto max-h-48">
                    {this.state.error?.toString()}
                    {"\n"}
                    {this.state.errorInfo?.componentStack}
                  </pre>
                </details>
              )}

              <div className="flex gap-3">
                <Button onClick={this.handleReset}>
                  <RefreshCcw className="h-4 w-4 mr-2" />
                  Tentar novamente
                </Button>
                <Button
                  variant="outline"
                  onClick={() => (window.location.href = "/")}
                >
                  <Home className="h-4 w-4 mr-2" />
                  Ir para o dashboard
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

// Hook for using error boundary
export function useErrorBoundary() {
  return React.useContext(ErrorBoundaryContext);
}

// Simple error display component for inline errors
interface InlineErrorProps {
  error?: Error | string | null;
  message?: string;
  className?: string;
  onRetry?: () => void;
}

export function InlineError({
  error,
  message = "Ocorreu um erro ao carregar.",
  className,
  onRetry,
}: InlineErrorProps) {
  const errorMessage = error instanceof Error ? error.message : error || message;

  return (
    <div className={cn("flex flex-col items-center justify-center p-8 text-center", className)}>
      <AlertCircle className="h-12 w-12 text-muted-foreground mb-4" />
      <p className="text-muted-foreground mb-4">{errorMessage}</p>
      {onRetry && (
        <Button variant="outline" size="sm" onClick={onRetry}>
          <RefreshCcw className="h-4 w-4 mr-2" />
          Tentar novamente
        </Button>
      )}
    </div>
  );
}

// Loading error state for data fetching
interface LoadingErrorProps {
  isLoading: boolean;
  error: Error | null;
  hasData: boolean;
  loadingComponent?: React.ReactNode;
  errorComponent?: React.ReactNode;
  emptyComponent?: React.ReactNode;
  children: React.ReactNode;
}

export function LoadingError({
  isLoading,
  error,
  hasData,
  loadingComponent,
  errorComponent,
  emptyComponent,
  children,
}: LoadingErrorProps) {
  if (isLoading) {
    return (
      loadingComponent || (
        <div className="flex justify-center p-8">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        </div>
      );
    }

  if (error) {
    return (
      errorComponent || <InlineError error={error} onRetry={() => window.location.reload()} />
    );
  }

  if (!hasData) {
    return (
      emptyComponent || (
        <div className="text-center p-8 text-muted-foreground">
          Nenhum dado disponível.
        </div>
      )
    );
  }

  return <>{children}</>;
}

// Async error boundary for async operations
interface AsyncErrorBoundaryState {
  error: Error | null;
}

export function AsyncErrorBoundary({
  children,
  fallback,
}: {
  children: React.ReactNode;
  fallback?: React.Component<{ error: Error; resetError: () => void }>;
}) {
  const [state, setState] = React.useState<AsyncErrorBoundaryState>({ error: null });

  const handleError = (error: Error) => {
    setState({ error });
  };

  const resetError = () => {
    setState({ error: null });
  };

  return (
    <ErrorBoundary fallback={fallback} onError={handleError}>
      {state.error ? (
        fallback ? (
          <fallback error={state.error} resetErrorBoundary={resetError} />
        ) : (
          <InlineError error={state.error} onRetry={resetError} />
        )
      ) : (
        <ErrorBoundaryProvider value={{ error: state.error, setError: handleError }}>
          {children}
        </ErrorBoundaryProvider>
      )}
    </ErrorBoundary>
  );
}

// Context for error boundary
interface ErrorBoundaryContextValue {
  error: Error | null;
  setError: (error: Error) => void;
}

const ErrorBoundaryContext = React.createContext<ErrorBoundaryContextValue | undefined>(
  undefined
);

function ErrorBoundaryProvider({
  value,
  children,
}: {
  value?: ErrorBoundaryContextValue;
  children: React.ReactNode;
}) {
  return (
    <ErrorBoundaryContext.Provider value={value ?? { error: null, setError: () => {} }}>
      {children}
    </ErrorBoundaryContext.Provider>
  );
}

export { ErrorBoundaryProvider };
