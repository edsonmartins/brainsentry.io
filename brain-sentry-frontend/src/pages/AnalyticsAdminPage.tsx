import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Activity, Zap, TrendingUp, Loader2, BarChart3, Clock, Database,
  Brain, Play, Timer, CheckCircle,
} from "lucide-react";
import { api, getErrorMessage } from "@/lib/api";
import { useToast } from "@/components/ui/toast";
import ReactEChartsCore from "echarts-for-react/lib/core";
import * as echarts from "echarts/core";
import { BarChart, LineChart, PieChart } from "echarts/charts";
import {
  GridComponent, TooltipComponent, TitleComponent, LegendComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";

echarts.use([BarChart, LineChart, PieChart, GridComponent, TooltipComponent, TitleComponent, LegendComponent, CanvasRenderer]);

interface BenchmarkResult {
  queryCount: number;
  k: number;
  avgLatencyMs: number;
  p50LatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
  avgRecall: number;
  avgPrecision: number;
  throughputQps: number;
}

export default function AnalyticsAdminPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [stats, setStats] = useState({
    totalMemories: 0,
    memoriesByCategory: {} as Record<string, number>,
    memoriesByImportance: {} as Record<string, number>,
    requestsToday: 0,
    injectionRate: 0,
    avgLatencyMs: 0,
    helpfulnessRate: 0,
    totalInjections: 0,
    activeMemories24h: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Benchmark state
  const [benchmarkRunning, setBenchmarkRunning] = useState(false);
  const [benchmarkResult, setBenchmarkResult] = useState<BenchmarkResult | null>(null);
  const [benchmarkQueryCount, setBenchmarkQueryCount] = useState(10);
  const [benchmarkK, setBenchmarkK] = useState(10);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await api.getStats();
        setStats(data);
      } catch (err) {
        setError(getErrorMessage(err));
      } finally {
        setLoading(false);
      }
    };
    fetchStats();
  }, []);

  const handleRunBenchmark = async () => {
    setBenchmarkRunning(true);
    setBenchmarkResult(null);
    try {
      const data = await api.runBenchmark(benchmarkQueryCount, benchmarkK);
      setBenchmarkResult(data);
      toast({ title: t("analytics.benchmarkDone"), variant: "success" });
    } catch (err) {
      toast({ title: t("analytics.benchmarkError"), description: getErrorMessage(err), variant: "error" });
    } finally {
      setBenchmarkRunning(false);
    }
  };

  const categoryColors: Record<string, string> = {
    INSIGHT: "#3B82F6", WARNING: "#EF4444", KNOWLEDGE: "#22C55E",
    ACTION: "#EAB308", CONTEXT: "#A855F7", REFERENCE: "#06B6D4",
    GENERAL: "#6B7280", PATTERN: "#3B82F6", DECISION: "#A855F7",
    BUG: "#EAB308", REFACTOR: "#22C55E",
  };

  const categoryBarColors: Record<string, string> = {
    INSIGHT: "bg-blue-500", WARNING: "bg-red-500", KNOWLEDGE: "bg-green-500",
    ACTION: "bg-yellow-500", CONTEXT: "bg-purple-500", REFERENCE: "bg-cyan-500",
    GENERAL: "bg-gray-500", PATTERN: "bg-blue-500", DECISION: "bg-purple-500",
    BUG: "bg-yellow-500", REFACTOR: "bg-green-500",
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="ml-2 text-muted-foreground">{t("analytics.loading")}</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 text-destructive p-4 rounded-md">
        <p className="font-medium">{t("analytics.loadError")}</p>
        <p className="text-sm">{error}</p>
      </div>
    );
  }

  // ECharts: Category Pie
  const categoryPieOption = {
    tooltip: { trigger: "item" },
    series: [{
      type: "pie",
      radius: ["40%", "70%"],
      data: Object.entries(stats.memoriesByCategory).map(([name, value]) => ({
        name, value, itemStyle: { color: categoryColors[name] || "#6B7280" },
      })),
      label: { fontSize: 11 },
    }],
  };

  // ECharts: Importance Bar
  const importanceBarOption = {
    tooltip: { trigger: "axis" },
    xAxis: {
      type: "category" as const,
      data: Object.keys(stats.memoriesByImportance),
    },
    yAxis: { type: "value" as const },
    series: [{
      type: "bar",
      data: Object.entries(stats.memoriesByImportance).map(([key, val]) => ({
        value: val,
        itemStyle: {
          color: key === "CRITICAL" ? "#EF4444" : key === "IMPORTANT" ? "#F97316" : "#6B7280",
        },
      })),
      barWidth: "50%",
    }],
  };

  // ECharts: Operations gauge-like metrics
  const operationsOption = {
    tooltip: { trigger: "axis" },
    xAxis: {
      type: "category" as const,
      data: [t("analytics.opsLabels.requestsToday"), t("analytics.opsLabels.injections"), t("analytics.opsLabels.active24h"), t("analytics.opsLabels.latency")],
    },
    yAxis: { type: "value" as const },
    series: [{
      type: "bar",
      data: [
        { value: stats.requestsToday, itemStyle: { color: "#3B82F6" } },
        { value: stats.totalInjections, itemStyle: { color: "#22C55E" } },
        { value: stats.activeMemories24h, itemStyle: { color: "#A855F7" } },
        { value: stats.avgLatencyMs, itemStyle: { color: "#F97316" } },
      ],
      barWidth: "50%",
    }],
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center gap-3">
            <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
              <BarChart3 className="h-5 w-5 text-white" />
            </div>
            <div>
              <h1 className="text-base font-bold leading-tight">{t("analytics.title")}</h1>
              <p className="text-xs text-white/80">{t("analytics.subtitle")}</p>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Stats Cards - All fields */}
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-6">
          {[
            { title: t("analytics.stats.totalMemories"), value: stats.totalMemories, icon: Database, color: "text-blue-600" },
            { title: t("analytics.stats.injectionRate"), value: `${(stats.injectionRate * 100).toFixed(1)}%`, icon: Zap, color: "text-green-600" },
            { title: t("analytics.stats.helpfulness"), value: `${(stats.helpfulnessRate * 100).toFixed(0)}%`, icon: TrendingUp, color: "text-purple-600" },
            { title: t("analytics.stats.requestsToday"), value: stats.requestsToday, icon: Activity, color: "text-blue-500" },
            { title: t("analytics.stats.avgLatency"), value: `${stats.avgLatencyMs.toFixed(0)}ms`, icon: Clock, color: "text-orange-500" },
            { title: t("analytics.stats.totalInjections"), value: stats.totalInjections, icon: Zap, color: "text-green-500" },
            { title: t("analytics.stats.active24h"), value: stats.activeMemories24h, icon: Timer, color: "text-purple-500" },
            { title: t("analytics.stats.categories"), value: Object.keys(stats.memoriesByCategory).length, icon: BarChart3, color: "text-cyan-500" },
          ].map((stat) => (
            <Card key={stat.title} className="shadow-sm">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs font-medium text-muted-foreground">{stat.title}</p>
                    <p className="text-lg font-bold leading-tight mt-1">{stat.value}</p>
                  </div>
                  <div className="p-2 bg-gradient-to-br from-brain-primary to-brain-accent rounded-lg text-white">
                    <stat.icon className="h-4 w-4" />
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Charts Row */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
          {/* Category Distribution - Pie Chart */}
          <Card>
            <CardHeader><CardTitle>{t("analytics.categoryDist")}</CardTitle></CardHeader>
            <CardContent>
              {Object.keys(stats.memoriesByCategory).length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-8">{t("analytics.noMemories")}</p>
              ) : (
                <ReactEChartsCore echarts={echarts} option={categoryPieOption} style={{ height: 280 }} />
              )}
            </CardContent>
          </Card>

          {/* Importance Distribution - Bar Chart */}
          <Card>
            <CardHeader><CardTitle>{t("analytics.importanceDist")}</CardTitle></CardHeader>
            <CardContent>
              {Object.keys(stats.memoriesByImportance).length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-8">{t("analytics.noMemories")}</p>
              ) : (
                <ReactEChartsCore echarts={echarts} option={importanceBarOption} style={{ height: 280 }} />
              )}
            </CardContent>
          </Card>
        </div>

        {/* Operations Chart */}
        <Card className="mb-6">
          <CardHeader><CardTitle>{t("analytics.opsMetrics")}</CardTitle></CardHeader>
          <CardContent>
            <ReactEChartsCore echarts={echarts} option={operationsOption} style={{ height: 280 }} />
          </CardContent>
        </Card>

        {/* Benchmark Section */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2"><Brain className="h-5 w-5" /> {t("analytics.benchmarkTitle")}</CardTitle>
                <p className="text-sm text-muted-foreground mt-1">{t("analytics.benchmarkDesc")}</p>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex items-end gap-4 mb-6">
              <div>
                <label className="text-xs font-medium text-muted-foreground">{t("analytics.queries")}</label>
                <input type="number" value={benchmarkQueryCount} min={1} max={100}
                  onChange={(e) => setBenchmarkQueryCount(Number(e.target.value))}
                  className="w-20 h-9 rounded-md border border-input bg-background px-2 text-sm" />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground">{t("analytics.topK")}</label>
                <input type="number" value={benchmarkK} min={1} max={100}
                  onChange={(e) => setBenchmarkK(Number(e.target.value))}
                  className="w-20 h-9 rounded-md border border-input bg-background px-2 text-sm" />
              </div>
              <Button onClick={handleRunBenchmark} disabled={benchmarkRunning}>
                {benchmarkRunning ? (
                  <><Loader2 className="h-4 w-4 animate-spin mr-2" /> {t("analytics.running")}</>
                ) : (
                  <><Play className="h-4 w-4 mr-2" /> {t("analytics.runBenchmark")}</>
                )}
              </Button>
            </div>

            {benchmarkResult && (
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {[
                  { label: t("analytics.metrics.avgLatency"), value: `${benchmarkResult.avgLatencyMs?.toFixed(1)}ms` },
                  { label: t("analytics.metrics.p50"), value: `${benchmarkResult.p50LatencyMs?.toFixed(1)}ms` },
                  { label: t("analytics.metrics.p95"), value: `${benchmarkResult.p95LatencyMs?.toFixed(1)}ms` },
                  { label: t("analytics.metrics.p99"), value: `${benchmarkResult.p99LatencyMs?.toFixed(1)}ms` },
                  { label: t("analytics.metrics.recall"), value: `${((benchmarkResult.avgRecall || 0) * 100).toFixed(1)}%` },
                  { label: t("analytics.metrics.precision"), value: `${((benchmarkResult.avgPrecision || 0) * 100).toFixed(1)}%` },
                  { label: t("analytics.metrics.throughput"), value: `${benchmarkResult.throughputQps?.toFixed(1)} QPS` },
                  { label: t("analytics.metrics.queries"), value: `${benchmarkResult.queryCount}` },
                ].map((m) => (
                  <div key={m.label} className="p-3 bg-accent rounded-lg">
                    <p className="text-xs text-muted-foreground">{m.label}</p>
                    <p className="text-lg font-bold">{m.value}</p>
                  </div>
                ))}
              </div>
            )}

            {!benchmarkResult && !benchmarkRunning && (
              <div className="text-center py-8 text-muted-foreground">
                <CheckCircle className="h-12 w-12 mx-auto mb-3 opacity-30" />
                <p className="text-sm">{t("analytics.benchmarkTip")}</p>
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
