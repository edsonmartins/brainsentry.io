import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams, useNavigate } from "react-router-dom";
// d3-transition extends d3-selection with .interrupt() — required by react-force-graph-2d.
import "d3-transition";
import ForceGraph2D from "react-force-graph-2d";
import {
  GitBranch, RefreshCw, X, Search, ArrowLeft, Target, Network as NetworkIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import { api, type GraphNode } from "@/lib/api/client";

const HOP_COLORS = ["#ef4444", "#f59e0b", "#eab308", "#10b981", "#6b7280"];

export default function GraphEgoPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const graphRef = useRef<any>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const [seedId, setSeedId] = useState(searchParams.get("id") ?? "");
  const [inputId, setInputId] = useState(searchParams.get("id") ?? "");
  const [hops, setHops] = useState(Number(searchParams.get("hops") ?? 2));
  const [limit, setLimit] = useState(Number(searchParams.get("limit") ?? 30));
  const [data, setData] = useState<{ nodes: GraphNode[]; links: any[] } | null>(null);
  const [loading, setLoading] = useState(false);
  const [selected, setSelected] = useState<GraphNode | null>(null);
  const [history, setHistory] = useState<string[]>([]);
  const [dims, setDims] = useState<{ w: number; h: number }>({ w: 800, h: 600 });

  const load = useCallback(async (id: string) => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await api.getGraphEgo(id, hops, limit);
      setData({
        nodes: (res.nodes || []).map((n) => ({ ...n })),
        links: (res.edges || []).map((e) => ({ ...e })),
      });
      setSeedId(id);
      setSearchParams({ id, hops: String(hops), limit: String(limit) });
    } catch (err: any) {
      toast({ title: t("graphEgo.loadError"), description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  }, [hops, limit, toast, t, setSearchParams]);

  useEffect(() => {
    if (seedId) load(seedId);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setDims({
          w: Math.max(400, entry.contentRect.width),
          h: Math.max(400, entry.contentRect.height),
        });
      }
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  const stats = useMemo(() => {
    if (!data) return { nodes: 0, edges: 0 };
    return { nodes: data.nodes.length, edges: data.links.length };
  }, [data]);

  const handleSearch = () => {
    if (!inputId.trim()) return;
    setHistory((h) => (seedId ? [seedId, ...h].slice(0, 10) : h));
    load(inputId.trim());
  };

  const handleNodeClick = useCallback((n: any) => {
    setSelected(n as GraphNode);
  }, []);

  const recenterOn = (id: string) => {
    setHistory((h) => [seedId, ...h].slice(0, 10));
    setInputId(id);
    setSelected(null);
    load(id);
  };

  const goBack = () => {
    if (history.length === 0) return;
    const [prev, ...rest] = history;
    setHistory(rest);
    setInputId(prev);
    setSelected(null);
    load(prev);
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <GitBranch className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("graphEgo.title")}</h1>
                <p className="text-xs text-white/80">{t("graphEgo.subtitle")}</p>
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              {history.length > 0 && (
                <Button
                  variant="outline"
                  size="sm"
                  className="bg-white/20 border-white/30 text-white hover:bg-white/30"
                  onClick={goBack}
                >
                  <ArrowLeft className="h-4 w-4" />
                </Button>
              )}
              <Button
                variant="outline"
                size="sm"
                className="bg-white/20 border-white/30 text-white hover:bg-white/30"
                onClick={() => seedId && load(seedId)}
                disabled={loading || !seedId}
              >
                <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="flex-1 flex flex-col min-h-0">
        <div className="border-b bg-muted/20 px-4 py-3 space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <input
              value={inputId}
              onChange={(e) => setInputId(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              placeholder={t("graphEgo.idPlaceholder")}
              className="flex-1 min-w-[200px] bg-background border rounded px-2 py-1 text-xs font-mono focus:outline-none focus:ring-1 focus:ring-brain-primary"
            />
            <label className="flex items-center gap-1 text-[11px]">
              <span className="text-muted-foreground">{t("graphEgo.hops")}:</span>
              <select
                value={hops}
                onChange={(e) => setHops(Number(e.target.value))}
                className="bg-background border rounded px-1 py-0.5 text-xs"
              >
                {[1, 2, 3, 4].map((n) => <option key={n} value={n}>{n}</option>)}
              </select>
            </label>
            <label className="flex items-center gap-1 text-[11px]">
              <span className="text-muted-foreground">{t("graphEgo.limit")}:</span>
              <input
                type="number"
                min={5}
                max={100}
                value={limit}
                onChange={(e) => setLimit(Number(e.target.value))}
                className="w-16 bg-background border rounded px-1 py-0.5 text-xs"
              />
            </label>
            <Button size="sm" onClick={handleSearch} disabled={!inputId.trim() || loading}>
              <Search className="h-3.5 w-3.5 mr-1" /> {t("graphEgo.explore")}
            </Button>
          </div>
          <div className="flex flex-wrap gap-4 text-[11px] text-muted-foreground">
            <span><strong className="font-mono text-foreground">{stats.nodes}</strong> {t("graphGlobal.nodes")}</span>
            <span><strong className="font-mono text-foreground">{stats.edges}</strong> {t("graphGlobal.edges")}</span>
            <span className="ml-auto flex items-center gap-2">
              {HOP_COLORS.slice(0, Math.min(hops + 1, HOP_COLORS.length)).map((c, i) => (
                <span key={i} className="flex items-center gap-1">
                  <span className="inline-block w-2 h-2 rounded-full" style={{ background: c }} />
                  {t("graphEgo.hopLabel", { n: i })}
                </span>
              ))}
            </span>
          </div>
        </div>

        <div className="flex-1 flex min-h-0">
          <div
            ref={containerRef}
            data-testid="graph-ego-canvas"
            className="flex-1 relative bg-background min-h-[500px]"
          >
            {loading ? (
              <div className="absolute inset-0 flex items-center justify-center">
                <Spinner size="lg" />
              </div>
            ) : !data || data.nodes.length === 0 ? (
              <div className="absolute inset-0 flex items-center justify-center">
                <EmptyState
                  icon={GitBranch}
                  title={t("graphEgo.empty.title")}
                  description={t("graphEgo.empty.desc")}
                  action={{
                    label: t("graphEgo.goMemories"),
                    onClick: () => navigate("/app/memories"),
                  }}
                />
              </div>
            ) : (
              <ForceGraph2D
                ref={graphRef}
                graphData={data as any}
                width={dims.w}
                height={dims.h}
                nodeId="id"
                nodeLabel={(n: any) => `${n.label}\n[hop ${n.hopDistance ?? 0}]`}
                nodeColor={(n: any) => HOP_COLORS[Math.min(n.hopDistance ?? 0, HOP_COLORS.length - 1)]}
                nodeVal={(n: any) => (n.hopDistance === 0 ? 12 : 4 + Math.max(0, 6 - (n.hopDistance ?? 0) * 1.5))}
                nodeRelSize={4}
                nodeCanvasObject={(node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
                  const color = HOP_COLORS[Math.min(node.hopDistance ?? 0, HOP_COLORS.length - 1)];
                  const isSeed = node.hopDistance === 0;
                  const r = isSeed ? 9 : 4 + Math.max(0, 6 - (node.hopDistance ?? 0) * 1.5);
                  ctx.beginPath();
                  ctx.arc(node.x, node.y, r, 0, 2 * Math.PI);
                  ctx.fillStyle = color;
                  ctx.fill();
                  if (isSeed) {
                    ctx.lineWidth = 2;
                    ctx.strokeStyle = "#fff";
                    ctx.stroke();
                  }
                  const label = node.label as string;
                  const fontSize = Math.max(10 / globalScale, 2);
                  ctx.font = `${fontSize}px sans-serif`;
                  ctx.textAlign = "center";
                  ctx.textBaseline = "middle";
                  ctx.fillStyle = "rgba(100,116,139,0.9)";
                  if (globalScale > 1.2 && label) {
                    ctx.fillText(label.slice(0, 28), node.x, node.y + r + fontSize);
                  }
                }}
                linkColor={() => "rgba(148,163,184,0.35)"}
                linkWidth={1}
                linkDirectionalArrowLength={3}
                linkDirectionalArrowRelPos={1}
                cooldownTicks={50}
                onNodeClick={handleNodeClick}
                onNodeDragEnd={(node: any) => {
                  node.fx = node.x;
                  node.fy = node.y;
                }}
                backgroundColor="transparent"
              />
            )}
          </div>

          {selected && (
            <aside className="w-80 border-l bg-muted/10 p-4 overflow-y-auto">
              <div className="flex items-start justify-between mb-3">
                <h3 className="text-sm font-semibold flex items-center gap-2">
                  <Target className="h-4 w-4" />
                  {t("graphEgo.detail")}
                </h3>
                <button onClick={() => setSelected(null)} className="p-1 rounded hover:bg-muted/50">
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>

              <div className="space-y-3 text-xs">
                <div>
                  <p className="text-[10px] uppercase tracking-wider text-muted-foreground">ID</p>
                  <p className="font-mono break-all text-[11px]">{selected.id}</p>
                </div>
                <div>
                  <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
                    {t("graphEgo.label")}
                  </p>
                  <p className="text-sm">{selected.label}</p>
                </div>
                <div className="flex gap-2 flex-wrap">
                  <span className="px-2 py-0.5 rounded bg-muted text-[10px]">{selected.category || "—"}</span>
                  <span className="px-2 py-0.5 rounded bg-muted text-[10px]">{selected.importance || "—"}</span>
                  <span
                    className="px-2 py-0.5 rounded text-[10px] text-white"
                    style={{ background: HOP_COLORS[Math.min(selected.hopDistance ?? 0, HOP_COLORS.length - 1)] }}
                  >
                    hop {selected.hopDistance ?? 0}
                  </span>
                  {selected.score != null && (
                    <span className="px-2 py-0.5 rounded bg-muted text-[10px] font-mono">
                      score {selected.score.toFixed(2)}
                    </span>
                  )}
                </div>

                {selected.id !== seedId && (
                  <Button size="sm" className="w-full" onClick={() => recenterOn(selected.id)}>
                    <NetworkIcon className="h-3.5 w-3.5 mr-1.5" />
                    {t("graphEgo.recenter")}
                  </Button>
                )}
              </div>
            </aside>
          )}
        </div>
      </main>
    </div>
  );
}
