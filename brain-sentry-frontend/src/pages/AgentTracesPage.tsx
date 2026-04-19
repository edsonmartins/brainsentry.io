import { useState, useEffect, useMemo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  Activity, Filter, RefreshCw, CheckCircle, XCircle, Brain, Clock,
  ChevronDown, ChevronUp,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { TypeChips } from "@/components/ui/TypeChips";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import {
  api,
  type AgentTrace,
  type AgentTraceStats,
  type AgentTraceFilter,
} from "@/lib/api/client";

export default function AgentTracesPage() {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const [traces, setTraces] = useState<AgentTrace[]>([]);
  const [stats, setStats] = useState<AgentTraceStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  // Filters
  const [statusFilter, setStatusFilter] = useState<"success" | "error" | null>(null);
  const [sessionFilter, setSessionFilter] = useState("");
  const [agentFilter, setAgentFilter] = useState("");

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const filter: AgentTraceFilter = { limit: 100 };
      if (statusFilter) filter.status = statusFilter;
      if (sessionFilter) filter.sessionId = sessionFilter;
      if (agentFilter) filter.agentId = agentFilter;

      const [list, s] = await Promise.all([
        api.listAgentTraces(filter),
        api.getAgentTraceStats(),
      ]);
      setTraces(list.traces || []);
      setStats(s);
    } catch (err: any) {
      toast({
        title: t("traces.loadFailed"),
        description: err?.message || t("common.error"),
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  }, [statusFilter, sessionFilter, agentFilter, toast, t]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const toggleExpand = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  // Derive chip options for status
  const statusChips = useMemo(() => {
    if (!stats) return [];
    return [
      { label: "success", count: stats.success, color: "#10b981" },
      { label: "error", count: stats.errors, color: "#ef4444" },
    ];
  }, [stats]);

  // Distinct agents/sessions for filter hints
  const distinctSessions = useMemo(
    () => Array.from(new Set(traces.map((t) => t.sessionId).filter(Boolean))) as string[],
    [traces]
  );
  const distinctAgents = useMemo(
    () => Array.from(new Set(traces.map((t) => t.agentId).filter(Boolean))) as string[],
    [traces]
  );

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Activity className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("traces.title")}</h1>
                <p className="text-xs text-white/80">{t("traces.subtitle")}</p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="bg-white/20 border-white/30 text-white hover:bg-white/30"
              onClick={fetchData}
              disabled={loading}
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6">
        {/* Stats */}
        <div className="grid grid-cols-2 lg:grid-cols-5 gap-3 mb-6">
          <StatCard
            label={t("traces.stats.total")}
            value={stats?.total ?? 0}
            icon={<Activity className="h-4 w-4" />}
            loading={loading}
          />
          <StatCard
            label={t("traces.stats.success")}
            value={stats?.success ?? 0}
            icon={<CheckCircle className="h-4 w-4" />}
            loading={loading}
            accent="#10b981"
          />
          <StatCard
            label={t("traces.stats.errors")}
            value={stats?.errors ?? 0}
            icon={<XCircle className="h-4 w-4" />}
            loading={loading}
            accent={stats && stats.errors > 0 ? "#ef4444" : undefined}
          />
          <StatCard
            label={t("traces.stats.withMemory")}
            value={stats?.withMemory ?? 0}
            icon={<Brain className="h-4 w-4" />}
            loading={loading}
          />
          <StatCard
            label={t("traces.stats.avgDuration")}
            value={`${Math.round(stats?.avgDurationMs ?? 0)}ms`}
            icon={<Clock className="h-4 w-4" />}
            loading={loading}
          />
        </div>

        {/* Filters */}
        <Card className="mb-6">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm flex items-center gap-2">
              <Filter className="h-4 w-4" /> {t("traces.filters")}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">{t("traces.status")}</p>
              <TypeChips
                items={statusChips}
                selected={statusFilter}
                onSelect={(v) => setStatusFilter(v as "success" | "error" | null)}
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">{t("traces.sessionId")}</p>
                <input
                  list="sessions"
                  value={sessionFilter}
                  onChange={(e) => setSessionFilter(e.target.value)}
                  placeholder={t("traces.sessionPlaceholder")}
                  className="w-full text-sm bg-transparent border rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-brain-primary"
                />
                <datalist id="sessions">
                  {distinctSessions.map((s) => <option key={s} value={s} />)}
                </datalist>
              </div>
              <div>
                <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">{t("traces.agentId")}</p>
                <input
                  list="agents"
                  value={agentFilter}
                  onChange={(e) => setAgentFilter(e.target.value)}
                  placeholder={t("traces.agentPlaceholder")}
                  className="w-full text-sm bg-transparent border rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-brain-primary"
                />
                <datalist id="agents">
                  {distinctAgents.map((a) => <option key={a} value={a} />)}
                </datalist>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Traces list */}
        {loading ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : traces.length === 0 ? (
          <EmptyState
            icon={Activity}
            title={t("traces.empty.title")}
            description={t("traces.empty.desc")}
          />
        ) : (
          <div className="space-y-2">
            {traces.map((tr) => (
              <TraceRow
                key={tr.id}
                trace={tr}
                expanded={expanded.has(tr.id)}
                onToggle={() => toggleExpand(tr.id)}
                locale={i18n.language}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}

function StatCard({
  label,
  value,
  icon,
  loading,
  accent,
}: {
  label: string;
  value: number | string;
  icon: React.ReactNode;
  loading: boolean;
  accent?: string;
}) {
  return (
    <Card className={accent ? "border-l-[3px]" : ""} style={accent ? { borderLeftColor: accent } : undefined}>
      <CardContent className="p-3">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</p>
            {loading ? (
              <Spinner size="sm" />
            ) : (
              <p className="text-xl font-bold" style={accent ? { color: accent } : undefined}>
                {value}
              </p>
            )}
          </div>
          <div className="p-2 rounded-md bg-muted text-muted-foreground">{icon}</div>
        </div>
      </CardContent>
    </Card>
  );
}

function TraceRow({
  trace,
  expanded,
  onToggle,
  locale,
}: {
  trace: AgentTrace;
  expanded: boolean;
  onToggle: () => void;
  locale: string;
}) {
  const { t } = useTranslation();
  const isError = trace.status === "error";
  const color = isError ? "#ef4444" : "#10b981";

  return (
    <Card className="border-l-[3px]" style={{ borderLeftColor: color }}>
      <CardContent className="p-3">
        <div className="flex items-start gap-3">
          <div className="flex-shrink-0 mt-0.5">
            {isError ? (
              <XCircle className="h-4 w-4 text-destructive" />
            ) : (
              <CheckCircle className="h-4 w-4 text-green-500" />
            )}
          </div>

          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-sm font-mono font-medium truncate">{trace.originFunction}</span>
              {trace.withMemory && (
                <span className="text-[10px] px-1.5 py-0.5 rounded border text-blue-500 border-blue-500/40">
                  <Brain className="h-2.5 w-2.5 inline mr-1" />
                  {t("traces.memory")}
                </span>
              )}
              {trace.sessionId && (
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                  {t("traces.session")} {trace.sessionId.slice(0, 8)}
                </span>
              )}
              {trace.agentId && (
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                  {trace.agentId}
                </span>
              )}
            </div>

            {trace.memoryQuery && (
              <p className="text-xs text-muted-foreground mt-1 truncate">
                <span className="text-foreground">{t("traces.query")}</span> {trace.memoryQuery}
              </p>
            )}

            {isError && trace.errorMessage && (
              <p className="text-xs text-destructive mt-1 line-clamp-2">{trace.errorMessage}</p>
            )}

            <div className="flex items-center gap-3 mt-2 text-[10px] text-muted-foreground">
              <span>
                <Clock className="h-2.5 w-2.5 inline mr-1" />
                {trace.durationMs}ms
              </span>
              <span>{new Date(trace.createdAt).toLocaleString(locale)}</span>
              {trace.memoryIds && trace.memoryIds.length > 0 && (
                <span>{trace.memoryIds.length} {t("traces.memories")}</span>
              )}
            </div>

            {expanded && (
              <div className="mt-3 pt-3 border-t space-y-2 text-xs">
                {trace.methodParams && Object.keys(trace.methodParams).length > 0 && (
                  <div>
                    <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">{t("traces.params")}</p>
                    <pre className="bg-muted p-2 rounded overflow-auto text-[11px]">
                      {JSON.stringify(trace.methodParams, null, 2)}
                    </pre>
                  </div>
                )}
                {trace.methodReturn != null && (
                  <div>
                    <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">{t("traces.returnValue")}</p>
                    <pre className="bg-muted p-2 rounded overflow-auto text-[11px] max-h-40">
                      {JSON.stringify(trace.methodReturn, null, 2)}
                    </pre>
                  </div>
                )}
                {trace.memoryContext && (
                  <div>
                    <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">{t("traces.memoryContext")}</p>
                    <pre className="bg-muted p-2 rounded overflow-auto text-[11px] max-h-40 whitespace-pre-wrap">
                      {trace.memoryContext}
                    </pre>
                  </div>
                )}
              </div>
            )}
          </div>

          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onToggle}>
            {expanded ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
