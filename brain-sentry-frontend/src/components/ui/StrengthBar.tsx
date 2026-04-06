interface StrengthBarProps {
  value: number; // 0-1 or 0-100
  max?: number;
  label?: string;
  showValue?: boolean;
  size?: "sm" | "md";
  className?: string;
}

function getBarColor(ratio: number): string {
  if (ratio >= 0.7) return "#10b981"; // green
  if (ratio >= 0.4) return "#f59e0b"; // yellow
  return "#ef4444"; // red
}

export function StrengthBar({ value, max = 1, label, showValue = true, size = "sm", className = "" }: StrengthBarProps) {
  const ratio = Math.min(1, Math.max(0, value / max));
  const color = getBarColor(ratio);
  const height = size === "sm" ? "h-1.5" : "h-2.5";

  return (
    <div className={`flex items-center gap-2 ${className}`}>
      {label && <span className="text-xs text-muted-foreground w-20 truncate">{label}</span>}
      <div className={`flex-1 bg-muted rounded-full overflow-hidden ${height}`}>
        <div
          className={`${height} rounded-full transition-all duration-300`}
          style={{ width: `${ratio * 100}%`, backgroundColor: color }}
        />
      </div>
      {showValue && (
        <span className="text-xs font-mono w-10 text-right" style={{ color }}>
          {max <= 1 ? `${Math.round(ratio * 100)}%` : `${Math.round(value)}`}
        </span>
      )}
    </div>
  );
}

interface ImportanceBarProps {
  importance: string;
  className?: string;
}

export function ImportanceBar({ importance, className = "" }: ImportanceBarProps) {
  const config: Record<string, { value: number; color: string }> = {
    CRITICAL: { value: 1.0, color: "#ef4444" },
    IMPORTANT: { value: 0.7, color: "#f59e0b" },
    MINOR: { value: 0.3, color: "#10b981" },
  };
  const { value, color } = config[importance] || { value: 0.5, color: "#6b7280" };

  return (
    <div className={`flex items-center gap-1.5 ${className}`}>
      <div className="w-16 bg-muted rounded-full overflow-hidden h-1.5">
        <div className="h-1.5 rounded-full" style={{ width: `${value * 100}%`, backgroundColor: color }} />
      </div>
      <span className="text-[10px] uppercase tracking-wider" style={{ color }}>{importance}</span>
    </div>
  );
}
