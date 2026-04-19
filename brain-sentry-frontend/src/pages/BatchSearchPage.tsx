import { useState, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Layers3, Zap, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import { api, type BatchSearchResponse, type BatchScore } from "@/lib/api/client";

const DEFAULT_QUERIES = `How does the memory pipeline work?
What are the cognitive features?
Which database backends are supported?`;

export default function BatchSearchPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [queriesText, setQueriesText] = useState(DEFAULT_QUERIES);
  const [limit, setLimit] = useState(10);
  const [result, setResult] = useState<BatchSearchResponse | null>(null);
  const [loading, setLoading] = useState(false);

  const runSearch = useCallback(async () => {
    const queries = queriesText.split("\n").map((q) => q.trim()).filter(Boolean);
    if (queries.length === 0) {
      toast({ title: t("batchSearch.warnEmpty"), variant: "warning" });
      return;
    }
    if (queries.length > 20) {
      toast({ title: t("batchSearch.warnTooMany"), variant: "warning" });
      return;
    }

    setLoading(true);
    try {
      const resp = await api.batchSearch({ queries, limit });
      setResult(resp);
    } catch (err: any) {
      toast({ title: t("batchSearch.failed"), description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  }, [queriesText, limit, toast, t]);

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center gap-3">
            <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
              <Layers3 className="h-5 w-5 text-white" />
            </div>
            <div>
              <h1 className="text-base font-bold leading-tight">{t("batchSearch.title")}</h1>
              <p className="text-xs text-white/80">{t("batchSearch.subtitle")}</p>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 max-w-6xl">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-4">
          <Card className="lg:col-span-2">
            <CardHeader>
              <CardTitle className="text-sm">{t("batchSearch.queries")}</CardTitle>
            </CardHeader>
            <CardContent>
              <textarea
                value={queriesText}
                onChange={(e) => setQueriesText(e.target.value)}
                className="w-full bg-transparent resize-y outline-none text-sm min-h-[150px] border rounded p-2 focus:ring-1 focus:ring-brain-primary font-mono"
                placeholder={DEFAULT_QUERIES}
              />
              <p className="text-[10px] text-muted-foreground mt-1">
                {t("batchSearch.queriesHint")}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("batchSearch.options")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div>
                <label className="text-[10px] uppercase tracking-wider text-muted-foreground">{t("batchSearch.perQueryLimit")}</label>
                <input
                  type="number"
                  min={1}
                  max={50}
                  value={limit}
                  onChange={(e) => setLimit(Number(e.target.value))}
                  className="w-full text-sm bg-transparent border rounded px-2 py-1 mt-1"
                />
              </div>
              <Button
                onClick={runSearch}
                disabled={loading}
                className="w-full bg-gradient-to-r from-brain-primary to-brain-accent text-white"
              >
                {loading ? (
                  <><Loader2 className="h-3.5 w-3.5 mr-2 animate-spin" />{t("batchSearch.running")}</>
                ) : (
                  <><Zap className="h-3.5 w-3.5 mr-2" />{t("batchSearch.runSearch")}</>
                )}
              </Button>
            </CardContent>
          </Card>
        </div>

        {result ? (
          <MatrixView result={result} />
        ) : (
          <EmptyState
            icon={Layers3}
            title={t("batchSearch.emptyTitle")}
            description={t("batchSearch.emptyDesc")}
          />
        )}
      </main>
    </div>
  );
}

function MatrixView({ result }: { result: BatchSearchResponse }) {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="p-3 flex flex-wrap items-center gap-4 text-xs">
          <span>
            <span className="text-muted-foreground">{t("batchSearch.matched")}</span>{" "}
            <span className="font-mono font-semibold">{result.results.length}</span>
          </span>
          <span>
            <span className="text-muted-foreground">{t("batchSearch.queriesCount")}</span>{" "}
            <span className="font-mono font-semibold">{result.queries.length}</span>
          </span>
          <span>
            <span className="text-muted-foreground">{t("batchSearch.time")}</span>{" "}
            <span className="font-mono font-semibold">{result.searchTimeMs}ms</span>
          </span>
        </CardContent>
      </Card>

      {result.results.length === 0 ? (
        <EmptyState
          icon={Zap}
          title={t("batchSearch.noMatches.title")}
          description={t("batchSearch.noMatches.desc")}
        />
      ) : (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t("batchSearch.matrixTitle")}</CardTitle>
          </CardHeader>
          <CardContent className="overflow-x-auto p-0">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-2 sticky left-0 bg-background">{t("batchSearch.memoryCol")}</th>
                  {result.queries.map((q, i) => (
                    <th key={i} className="p-2 text-center min-w-[120px] max-w-[180px]">
                      <div className="font-normal text-muted-foreground truncate" title={q}>
                        Q{i + 1}
                      </div>
                      <div className="truncate text-[10px] text-muted-foreground italic max-w-[170px] mx-auto" title={q}>
                        {q}
                      </div>
                    </th>
                  ))}
                  <th className="p-2 text-center">{t("batchSearch.maxCol")}</th>
                </tr>
              </thead>
              <tbody>
                {result.results.map((r) => <MatrixRow key={r.memoryId} row={r} />)}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function MatrixRow({ row }: { row: BatchScore }) {
  const { t } = useTranslation();
  return (
    <tr className="border-b hover:bg-muted/30">
      <td className="p-2 sticky left-0 bg-background">
        <div className="max-w-[260px]">
          <p className="font-medium truncate text-sm">
            {row.summary || row.memoryId.slice(0, 8)}
          </p>
          <div className="flex items-center gap-2 text-[10px] text-muted-foreground mt-0.5">
            {row.category && <span>{row.category}</span>}
            <span className="font-mono">{row.memoryId.slice(0, 8)}</span>
            <span>{t("batchSearch.matchedOf", { matched: row.matchedQueries.length, total: row.perQuery.length })}</span>
          </div>
        </div>
      </td>
      {row.perQuery.map((score, i) => (
        <td key={i} className="p-1 text-center">
          <HeatCell value={score} matched={row.matchedQueries.includes(i)} />
        </td>
      ))}
      <td className="p-2 text-center font-mono">
        {row.max.toFixed(2)}
      </td>
    </tr>
  );
}

function HeatCell({ value, matched }: { value: number; matched: boolean }) {
  const ratio = Math.min(1, Math.max(0, value));
  const opacity = matched ? 0.2 + 0.8 * ratio : 0.05;
  return (
    <div
      className="inline-flex items-center justify-center rounded px-2 py-1 min-w-[50px]"
      style={{
        backgroundColor: matched ? `rgba(230, 126, 80, ${opacity})` : "rgba(107, 114, 128, 0.1)",
        color: matched && ratio > 0.5 ? "white" : undefined,
      }}
    >
      <span className="font-mono text-[11px]">{value.toFixed(2)}</span>
    </div>
  );
}
