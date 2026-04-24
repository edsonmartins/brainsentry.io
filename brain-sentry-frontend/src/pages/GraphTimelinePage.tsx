import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import {
  CalendarClock, RefreshCw, X, ArrowRight, GitBranch, ExternalLink,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import { api, type GraphNode, type GraphEdge } from "@/lib/api/client";

const CATEGORY_COLORS: Record<string, string> = {
  INSIGHT: "#3b82f6", DECISION: "#8b5cf6", WARNING: "#ef4444",
  KNOWLEDGE: "#10b981", ACTION: "#f59e0b", CONTEXT: "#06b6d4",
  REFERENCE: "#64748b", PATTERN: "#10b981", ANTIPATTERN: "#ef4444",
  BUG: "#dc2626", OPTIMIZATION: "#8b5cf6", INTEGRATION: "#06b6d4",
  DOMAIN: "#f59e0b",
};

const RANGE_PRESETS: Array<{ key: string; labelKey: string; days?: number }> = [
  { key: "24h", labelKey: "graphTimeline.range24h", days: 1 },
  { key: "7d", labelKey: "graphTimeline.range7d", days: 7 },
  { key: "30d", labelKey: "graphTimeline.range30d", days: 30 },
  { key: "all", labelKey: "graphTimeline.rangeAll" },
];

const LANE_HEIGHT = 56;
const MARGIN = { top: 24, right: 32, bottom: 40, left: 140 };

export default function GraphTimelinePage() {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const containerRef = useRef<HTMLDivElement | null>(null);

  const [nodes, setNodes] = useState<GraphNode[]>([]);
  const [edges, setEdges] = useState<GraphEdge[]>([]);
  const [loading, setLoading] = useState(true);
  const [range, setRange] = useState<string>("30d");
  const [selected, setSelected] = useState<GraphNode | null>(null);
  const [width, setWidth] = useState<number>(1000);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const preset = RANGE_PRESETS.find((r) => r.key === range);
      const params: { from?: string; to?: string; limit?: number } = { limit: 500 };
      if (preset?.days != null) {
        const now = new Date();
        const from = new Date(now.getTime() - preset.days * 24 * 60 * 60 * 1000);
        params.from = from.toISOString();
        params.to = now.toISOString();
      }
      const res = await api.getGraphTimeline(params);
      setNodes(res.nodes || []);
      setEdges(res.edges || []);
    } catch (err: any) {
      toast({ title: t("graphTimeline.loadError"), description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  }, [range, toast, t]);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setWidth(Math.max(800, entry.contentRect.width));
      }
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  const { lanes, tExtent } = useMemo(() => {
    const lanesSet = new Set<string>();
    let tMin = Number.POSITIVE_INFINITY;
    let tMax = Number.NEGATIVE_INFINITY;
    for (const n of nodes) {
      lanesSet.add(n.category || "OTHER");
      const ts = new Date(n.recordedAt).getTime();
      if (ts < tMin) tMin = ts;
      if (ts > tMax) tMax = ts;
    }
    if (!isFinite(tMin)) tMin = Date.now() - 24 * 3600 * 1000;
    if (!isFinite(tMax)) tMax = Date.now();
    if (tMin === tMax) tMax = tMin + 60_000;
    const ordered = Array.from(lanesSet).sort();
    return { lanes: ordered, tExtent: [tMin, tMax] as [number, number] };
  }, [nodes]);

  const height = MARGIN.top + MARGIN.bottom + lanes.length * LANE_HEIGHT;
  const innerWidth = Math.max(400, width - MARGIN.left - MARGIN.right);

  const xFor = useCallback(
    (iso: string) => {
      const ts = new Date(iso).getTime();
      const [a, b] = tExtent;
      if (b === a) return MARGIN.left;
      return MARGIN.left + ((ts - a) / (b - a)) * innerWidth;
    },
    [tExtent, innerWidth],
  );

  const yFor = useCallback(
    (category: string) => {
      const idx = lanes.indexOf(category || "OTHER");
      return MARGIN.top + Math.max(0, idx) * LANE_HEIGHT + LANE_HEIGHT / 2;
    },
    [lanes],
  );

  const ticks = useMemo(() => {
    const [a, b] = tExtent;
    const n = Math.min(8, Math.max(3, Math.floor(innerWidth / 130)));
    const out: number[] = [];
    for (let i = 0; i <= n; i++) out.push(a + ((b - a) * i) / n);
    return out;
  }, [tExtent, innerWidth]);

  const nodesById = useMemo(() => {
    const m = new Map<string, GraphNode>();
    for (const n of nodes) m.set(n.id, n);
    return m;
  }, [nodes]);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <CalendarClock className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("graphTimeline.title")}</h1>
                <p className="text-xs text-white/80">{t("graphTimeline.subtitle")}</p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="bg-white/20 border-white/30 text-white hover:bg-white/30"
              onClick={load}
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            </Button>
          </div>
        </div>
      </header>

      <main className="flex-1 flex flex-col min-h-0">
        <div className="border-b bg-muted/20 px-4 py-3 flex flex-wrap items-center gap-2">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground mr-1">
            {t("graphTimeline.range")}
          </span>
          {RANGE_PRESETS.map((r) => (
            <button
              key={r.key}
              onClick={() => setRange(r.key)}
              className={`px-2 py-0.5 text-[10px] rounded-full border transition-colors uppercase ${
                range === r.key
                  ? "bg-foreground text-background"
                  : "text-muted-foreground border-border hover:border-foreground/50"
              }`}
            >
              {t(r.labelKey)}
            </button>
          ))}
          <div className="ml-auto flex gap-4 text-[11px] text-muted-foreground">
            <span><strong className="font-mono text-foreground">{nodes.length}</strong> {t("graphTimeline.memories")}</span>
            <span><strong className="font-mono text-foreground">{edges.length}</strong> {t("graphTimeline.supersedes")}</span>
            <span className="flex items-center gap-1">
              <span className="w-3 h-0.5 bg-red-500" />
              {t("graphTimeline.legendSupersedes")}
            </span>
          </div>
        </div>

        <div className="flex-1 flex min-h-0">
          <div ref={containerRef} className="flex-1 overflow-auto bg-background">
            {loading ? (
              <div className="flex justify-center py-16"><Spinner size="lg" /></div>
            ) : nodes.length === 0 ? (
              <div className="flex items-center justify-center h-full">
                <EmptyState
                  icon={CalendarClock}
                  title={t("graphTimeline.empty.title")}
                  description={t("graphTimeline.empty.desc")}
                />
              </div>
            ) : (
              <svg
                width={width}
                height={height}
                className="block"
                data-testid="timeline-svg"
              >
                <defs>
                  <marker
                    id="sup-arrow"
                    viewBox="0 -5 10 10"
                    refX={10}
                    refY={0}
                    markerWidth={6}
                    markerHeight={6}
                    orient="auto"
                  >
                    <path d="M0,-5L10,0L0,5" fill="#ef4444" />
                  </marker>
                </defs>

                {/* Lanes */}
                {lanes.map((lane, i) => {
                  const y = MARGIN.top + i * LANE_HEIGHT;
                  const color = CATEGORY_COLORS[lane] || "#64748b";
                  return (
                    <g key={lane}>
                      <rect
                        x={0}
                        y={y}
                        width={width}
                        height={LANE_HEIGHT}
                        fill={i % 2 === 0 ? "rgba(148,163,184,0.04)" : "transparent"}
                      />
                      <text
                        x={MARGIN.left - 10}
                        y={y + LANE_HEIGHT / 2}
                        textAnchor="end"
                        dominantBaseline="middle"
                        className="text-[10px] fill-muted-foreground uppercase tracking-wider"
                      >
                        <tspan fill={color}>●</tspan> {lane}
                      </text>
                    </g>
                  );
                })}

                {/* X axis */}
                <line
                  x1={MARGIN.left}
                  y1={height - MARGIN.bottom + 4}
                  x2={width - MARGIN.right}
                  y2={height - MARGIN.bottom + 4}
                  className="stroke-border"
                />
                {ticks.map((t) => {
                  const x = MARGIN.left + ((t - tExtent[0]) / (tExtent[1] - tExtent[0] || 1)) * innerWidth;
                  return (
                    <g key={t}>
                      <line
                        x1={x}
                        y1={MARGIN.top}
                        x2={x}
                        y2={height - MARGIN.bottom + 4}
                        className="stroke-border/40"
                        strokeDasharray="3 3"
                      />
                      <text
                        x={x}
                        y={height - MARGIN.bottom + 20}
                        textAnchor="middle"
                        className="text-[10px] fill-muted-foreground font-mono"
                      >
                        {new Date(t).toLocaleDateString(i18n.language, { month: "short", day: "2-digit", hour: "2-digit" })}
                      </text>
                    </g>
                  );
                })}

                {/* SUPERSEDES arrows */}
                {edges.map((e, i) => {
                  const src = nodesById.get(e.source);
                  const tgt = nodesById.get(e.target);
                  if (!src || !tgt) return null;
                  const x1 = xFor(src.recordedAt);
                  const y1 = yFor(src.category || "OTHER");
                  const x2 = xFor(tgt.recordedAt);
                  const y2 = yFor(tgt.category || "OTHER");
                  const midY = (y1 + y2) / 2 - Math.abs(x2 - x1) * 0.1;
                  return (
                    <path
                      key={`e${i}`}
                      d={`M${x1},${y1} Q${(x1 + x2) / 2},${midY} ${x2},${y2}`}
                      fill="none"
                      stroke="#ef4444"
                      strokeWidth={1.3}
                      markerEnd="url(#sup-arrow)"
                      opacity={0.7}
                    />
                  );
                })}

                {/* Node markers */}
                {nodes.map((n) => {
                  const x = xFor(n.recordedAt);
                  const y = yFor(n.category || "OTHER");
                  const color = CATEGORY_COLORS[n.category || "OTHER"] || "#64748b";
                  const isSelected = selected?.id === n.id;
                  const isSuperseded = !!n.supersededBy;
                  return (
                    <g key={n.id} onClick={() => setSelected(n)} className="cursor-pointer">
                      <circle
                        cx={x}
                        cy={y}
                        r={isSelected ? 8 : 5}
                        fill={color}
                        fillOpacity={isSuperseded ? 0.35 : 0.9}
                        stroke={isSelected ? "#fff" : "none"}
                        strokeWidth={2}
                      >
                        <title>{`${n.label}\n${new Date(n.recordedAt).toLocaleString(i18n.language)}`}</title>
                      </circle>
                    </g>
                  );
                })}
              </svg>
            )}
          </div>

          {selected && (
            <aside className="w-80 border-l bg-muted/10 p-4 overflow-y-auto">
              <div className="flex items-start justify-between mb-3">
                <h3 className="text-sm font-semibold flex items-center gap-2">
                  <CalendarClock className="h-4 w-4" />
                  {t("graphTimeline.detail")}
                </h3>
                <button onClick={() => setSelected(null)} className="p-1 rounded hover:bg-muted/50">
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>

              <div className="space-y-3 text-xs">
                <div>
                  <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
                    {t("graphTimeline.stat.recorded")}
                  </p>
                  <p className="font-mono text-[11px]">
                    {new Date(selected.recordedAt).toLocaleString(i18n.language)}
                  </p>
                </div>
                {selected.validFrom && (
                  <div>
                    <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
                      {t("graphTimeline.stat.validFrom")}
                    </p>
                    <p className="font-mono text-[11px]">
                      {new Date(selected.validFrom).toLocaleString(i18n.language)}
                    </p>
                  </div>
                )}
                {selected.validTo && (
                  <div>
                    <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
                      {t("graphTimeline.stat.validTo")}
                    </p>
                    <p className="font-mono text-[11px]">
                      {new Date(selected.validTo).toLocaleString(i18n.language)}
                    </p>
                  </div>
                )}
                <div>
                  <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
                    {t("graphTimeline.stat.label")}
                  </p>
                  <p>{selected.label}</p>
                </div>
                <div className="flex gap-2 flex-wrap">
                  <span className="px-2 py-0.5 rounded bg-muted text-[10px]">{selected.category || "—"}</span>
                  <span className="px-2 py-0.5 rounded bg-muted text-[10px]">{selected.importance || "—"}</span>
                </div>
                {selected.supersededBy && (
                  <div className="p-2 rounded bg-red-500/10 border border-red-500/20 flex items-center gap-2">
                    <ArrowRight className="h-3 w-3 text-red-500 flex-shrink-0" />
                    <span className="font-mono text-[10px] break-all">{selected.supersededBy}</span>
                  </div>
                )}

                <div className="pt-2 space-y-2">
                  <Button
                    size="sm"
                    variant="outline"
                    className="w-full"
                    onClick={() => navigate(`/app/graph/ego?id=${encodeURIComponent(selected.id)}`)}
                  >
                    <GitBranch className="h-3.5 w-3.5 mr-1.5" />
                    {t("graphTimeline.openEgo")}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="w-full"
                    onClick={() => navigate(`/app/memories?id=${encodeURIComponent(selected.id)}`)}
                  >
                    <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                    {t("graphTimeline.openMemory")}
                  </Button>
                </div>
              </div>
            </aside>
          )}
        </div>
      </main>
    </div>
  );
}
