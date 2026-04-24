import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import ForceGraph2D from "react-force-graph-2d";
import {
  Network, RefreshCw, X, GitBranch, ExternalLink,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import { api, type GraphNode, type GraphEdge } from "@/lib/api/client";

interface Graph {
  nodes: (GraphNode & { x?: number; y?: number })[];
  links: (GraphEdge & { source: string | GraphNode; target: string | GraphNode })[];
  communities: number;
  modularity: number;
}

const CATEGORIES = [
  "INSIGHT", "DECISION", "WARNING", "KNOWLEDGE",
  "ACTION", "CONTEXT", "REFERENCE",
];
const IMPORTANCES = ["CRITICAL", "IMPORTANT", "MINOR"];

function colorForCommunity(id: number): string {
  if (id < 0) return "#94a3b8"; // slate-400 for unassigned
  // Golden angle distribution for visually distinct hues
  const hue = (id * 137.508) % 360;
  return `hsl(${hue.toFixed(0)}, 68%, 55%)`;
}

function nodeSize(n: GraphNode): number {
  const base = 3;
  const boost = Math.log2(1 + (n.accessCount ?? 0) + (n.helpfulCount ?? 0));
  return base + boost * 1.2;
}

function feedbackOpacity(n: GraphNode): number {
  const h = n.helpfulCount ?? 0;
  const nh = n.notHelpfulCount ?? 0;
  const total = h + nh;
  if (total === 0) return 0.85;
  return Math.max(0.35, h / total);
}

export default function GraphGlobalPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const graphRef = useRef<any>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const [data, setData] = useState<Graph | null>(null);
  const [loading, setLoading] = useState(true);
  const [category, setCategory] = useState<string | null>(null);
  const [importance, setImportance] = useState<string | null>(null);
  const [showFeedbackOverlay, setShowFeedbackOverlay] = useState(false);
  const [selected, setSelected] = useState<GraphNode | null>(null);
  const [dims, setDims] = useState<{ w: number; h: number }>({ w: 800, h: 600 });

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getGraphGlobal({
        limit: 500,
        category: category ?? undefined,
        importance: importance ?? undefined,
        communities: true,
      });
      const nodes = (res.nodes || []).map((n) => ({ ...n }));
      const links = (res.edges || []).map((e) => ({ ...e }));
      setData({
        nodes,
        links,
        communities: res.communities?.length ?? 0,
        modularity: res.modularity ?? 0,
      });
    } catch (err: any) {
      toast({
        title: t("graphGlobal.loadError"),
        description: err?.message,
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  }, [category, importance, toast, t]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        setDims({ w: Math.max(400, width), h: Math.max(400, height) });
      }
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  const stats = useMemo(() => {
    if (!data) return { nodes: 0, edges: 0, communities: 0, modularity: 0 };
    return {
      nodes: data.nodes.length,
      edges: data.links.length,
      communities: data.communities,
      modularity: data.modularity,
    };
  }, [data]);

  const handleNodeClick = useCallback((n: any) => {
    setSelected(n as GraphNode);
  }, []);

  const openEgoView = () => {
    if (!selected) return;
    navigate(`/app/graph/ego?id=${encodeURIComponent(selected.id)}`);
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Network className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("graphGlobal.title")}</h1>
                <p className="text-xs text-white/80">{t("graphGlobal.subtitle")}</p>
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
        {/* Filters + stats */}
        <div className="border-b bg-muted/20 px-4 py-3 space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-[10px] uppercase tracking-wider text-muted-foreground mr-1">
              {t("graphGlobal.category")}
            </span>
            <button
              onClick={() => setCategory(null)}
              className={`px-2 py-0.5 text-[10px] rounded-full border transition-colors ${
                !category ? "bg-foreground text-background" : "text-muted-foreground border-border hover:border-foreground/50"
              }`}
            >
              {t("graphGlobal.all")}
            </button>
            {CATEGORIES.map((c) => (
              <button
                key={c}
                onClick={() => setCategory(c === category ? null : c)}
                className={`px-2 py-0.5 text-[10px] rounded-full border transition-colors ${
                  category === c ? "bg-foreground text-background" : "text-muted-foreground border-border hover:border-foreground/50"
                }`}
              >
                {c}
              </button>
            ))}
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-[10px] uppercase tracking-wider text-muted-foreground mr-1">
              {t("graphGlobal.importance")}
            </span>
            {IMPORTANCES.map((i) => (
              <button
                key={i}
                onClick={() => setImportance(i === importance ? null : i)}
                className={`px-2 py-0.5 text-[10px] rounded-full border transition-colors ${
                  importance === i ? "bg-foreground text-background" : "text-muted-foreground border-border hover:border-foreground/50"
                }`}
              >
                {i}
              </button>
            ))}
            <label className="ml-auto flex items-center gap-1.5 text-[11px] text-muted-foreground cursor-pointer">
              <input
                type="checkbox"
                checked={showFeedbackOverlay}
                onChange={(e) => setShowFeedbackOverlay(e.target.checked)}
                className="accent-brain-primary"
              />
              {t("graphGlobal.overlayFeedback")}
            </label>
          </div>
          <div className="flex flex-wrap gap-4 text-[11px] text-muted-foreground pt-1">
            <span><strong className="font-mono text-foreground">{stats.nodes}</strong> {t("graphGlobal.nodes")}</span>
            <span><strong className="font-mono text-foreground">{stats.edges}</strong> {t("graphGlobal.edges")}</span>
            <span><strong className="font-mono text-foreground">{stats.communities}</strong> {t("graphGlobal.communities")}</span>
            <span><strong className="font-mono text-foreground">{stats.modularity.toFixed(3)}</strong> {t("graphGlobal.modularity")}</span>
          </div>
        </div>

        {/* Graph canvas */}
        <div className="flex-1 flex min-h-0">
          <div ref={containerRef} className="flex-1 relative bg-background min-h-[500px]">
            {loading ? (
              <div className="absolute inset-0 flex items-center justify-center">
                <Spinner size="lg" />
              </div>
            ) : !data || data.nodes.length === 0 ? (
              <div className="absolute inset-0 flex items-center justify-center">
                <EmptyState
                  icon={Network}
                  title={t("graphGlobal.empty.title")}
                  description={t("graphGlobal.empty.desc")}
                />
              </div>
            ) : (
              <ForceGraph2D
                ref={graphRef}
                graphData={data as any}
                width={dims.w}
                height={dims.h}
                nodeId="id"
                nodeLabel={(n: any) => `${n.label}\n[${n.category ?? "?"} · ${n.importance ?? "?"}]`}
                nodeColor={(n: any) => colorForCommunity(n.communityId)}
                nodeVal={nodeSize as any}
                nodeRelSize={4}
                nodeCanvasObjectMode={() => (showFeedbackOverlay ? "replace" : undefined)}
                nodeCanvasObject={
                  showFeedbackOverlay
                    ? (node: any, ctx: CanvasRenderingContext2D) => {
                        const size = nodeSize(node);
                        const op = feedbackOpacity(node);
                        ctx.globalAlpha = op;
                        ctx.beginPath();
                        ctx.arc(node.x, node.y, size, 0, 2 * Math.PI);
                        ctx.fillStyle = colorForCommunity(node.communityId);
                        ctx.fill();
                        ctx.globalAlpha = 1;
                      }
                    : undefined
                }
                linkSource="source"
                linkTarget="target"
                linkColor={(l: any) => (l.type === "SUPERSEDES" ? "rgba(239,68,68,0.5)" : "rgba(148,163,184,0.25)")}
                linkWidth={(l: any) => (l.type === "SUPERSEDES" ? 1.5 : 0.7)}
                linkDirectionalArrowLength={(l: any) => (l.type === "SUPERSEDES" ? 4 : 0)}
                linkDirectionalArrowRelPos={1}
                cooldownTicks={60}
                onNodeClick={handleNodeClick}
                onNodeDragEnd={(node: any) => {
                  node.fx = node.x;
                  node.fy = node.y;
                }}
                backgroundColor="transparent"
              />
            )}
          </div>

          {/* Side panel */}
          {selected && (
            <aside className="w-80 border-l bg-muted/10 p-4 overflow-y-auto">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2">
                  <span
                    className="inline-block h-3 w-3 rounded-full"
                    style={{ background: colorForCommunity(selected.communityId) }}
                  />
                  <h3 className="text-sm font-semibold">{t("graphGlobal.detail")}</h3>
                </div>
                <button
                  onClick={() => setSelected(null)}
                  className="p-1 rounded hover:bg-muted/50"
                >
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
                    {t("graphGlobal.label")}
                  </p>
                  <p className="text-sm">{selected.label}</p>
                </div>
                <div className="flex gap-2">
                  <span className="px-2 py-0.5 rounded bg-muted text-[10px]">
                    {selected.category || "—"}
                  </span>
                  <span className="px-2 py-0.5 rounded bg-muted text-[10px]">
                    {selected.importance || "—"}
                  </span>
                  {selected.communityId >= 0 && (
                    <span className="px-2 py-0.5 rounded text-[10px] text-white" style={{ background: colorForCommunity(selected.communityId) }}>
                      C{selected.communityId}
                    </span>
                  )}
                </div>
                <div className="grid grid-cols-2 gap-2 text-[11px]">
                  <Stat label={t("graphGlobal.stat.access")} value={selected.accessCount ?? 0} />
                  <Stat label={t("graphGlobal.stat.helpful")} value={selected.helpfulCount ?? 0} />
                  <Stat label={t("graphGlobal.stat.notHelpful")} value={selected.notHelpfulCount ?? 0} />
                  <Stat label={t("graphGlobal.stat.emotional")} value={(selected.emotionalWeight ?? 0).toFixed(2)} />
                </div>
                <div>
                  <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
                    {t("graphGlobal.stat.recorded")}
                  </p>
                  <p>{new Date(selected.recordedAt).toLocaleString()}</p>
                </div>
                {selected.supersededBy && (
                  <div className="p-2 rounded bg-red-500/10 border border-red-500/20">
                    <p className="text-[10px] text-red-500 font-medium uppercase tracking-wider">
                      {t("graphGlobal.superseded")}
                    </p>
                    <p className="font-mono text-[10px] break-all mt-0.5">{selected.supersededBy}</p>
                  </div>
                )}

                <div className="pt-2 space-y-2">
                  <Button size="sm" variant="outline" className="w-full" onClick={openEgoView}>
                    <GitBranch className="h-3.5 w-3.5 mr-1.5" />
                    {t("graphGlobal.openEgo")}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="w-full"
                    onClick={() => navigate(`/app/memories?id=${encodeURIComponent(selected.id)}`)}
                  >
                    <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                    {t("graphGlobal.openMemory")}
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

function Stat({ label, value }: { label: string; value: number | string }) {
  return (
    <Card className="p-2">
      <CardContent className="p-0 text-center">
        <p className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</p>
        <p className="font-mono text-sm">{value}</p>
      </CardContent>
    </Card>
  );
}

// Silence unused import warnings for Card types used via Stat.
void CardHeader;
void CardTitle;
