import { useTranslation } from "react-i18next";
import type { WebSocketStatus } from "@/hooks/useWebSocket";

interface LiveIndicatorProps {
  status: WebSocketStatus;
  className?: string;
}

const statusConfig: Record<WebSocketStatus, { color: string; animate: boolean; labelKey: string }> = {
  connected: { color: "bg-green-500", animate: true, labelKey: "indicator.connected" },
  connecting: { color: "bg-yellow-500", animate: true, labelKey: "indicator.connecting" },
  disconnected: { color: "bg-gray-500", animate: false, labelKey: "indicator.disconnected" },
  error: { color: "bg-red-500", animate: false, labelKey: "indicator.errorStatus" },
};

export function LiveIndicator({ status, className = "" }: LiveIndicatorProps) {
  const { t } = useTranslation();
  const config = statusConfig[status];

  return (
    <div className={`flex items-center gap-1.5 ${className}`}>
      <span className="relative flex h-2 w-2">
        {config.animate && (
          <span className={`animate-ping absolute inline-flex h-full w-full rounded-full ${config.color} opacity-75`} />
        )}
        <span className={`relative inline-flex rounded-full h-2 w-2 ${config.color}`} />
      </span>
      <span className="text-[10px] text-muted-foreground uppercase tracking-wider">{t(config.labelKey)}</span>
    </div>
  );
}
