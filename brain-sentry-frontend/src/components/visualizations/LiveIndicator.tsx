import type { WebSocketStatus } from "@/hooks/useWebSocket";

interface LiveIndicatorProps {
  status: WebSocketStatus;
  className?: string;
}

const statusConfig: Record<WebSocketStatus, { color: string; label: string; animate: boolean }> = {
  connected: { color: "bg-green-500", label: "Live", animate: true },
  connecting: { color: "bg-yellow-500", label: "Connecting", animate: true },
  disconnected: { color: "bg-gray-500", label: "Offline", animate: false },
  error: { color: "bg-red-500", label: "Error", animate: false },
};

export function LiveIndicator({ status, className = "" }: LiveIndicatorProps) {
  const config = statusConfig[status];

  return (
    <div className={`flex items-center gap-1.5 ${className}`}>
      <span className="relative flex h-2 w-2">
        {config.animate && (
          <span className={`animate-ping absolute inline-flex h-full w-full rounded-full ${config.color} opacity-75`} />
        )}
        <span className={`relative inline-flex rounded-full h-2 w-2 ${config.color}`} />
      </span>
      <span className="text-[10px] text-muted-foreground uppercase tracking-wider">{config.label}</span>
    </div>
  );
}
