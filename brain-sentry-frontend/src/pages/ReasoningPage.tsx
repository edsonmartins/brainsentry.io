import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Brain, Play, ChevronDown, ChevronUp, Wand2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import {
  api,
  type AbductionResult,
  type AbductionHypothesis,
} from "@/lib/api/client";

export default function ReasoningPage() {
  const { t } = useTranslation();
  const { toast } = useToast();

  const [decisionId, setDecisionId] = useState("");
  const [question, setQuestion] = useState("");
  const [maxHypotheses, setMaxHypotheses] = useState(5);
  const [result, setResult] = useState<AbductionResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [expanded, setExpanded] = useState<Set<number>>(new Set());

  const fetchLatest = async () => {
    setFetching(true);
    try {
      const res = await api.listDecisions({ limit: 1 });
      if (res.decisions && res.decisions.length > 0) {
        setDecisionId(res.decisions[0].id);
        toast({ title: "Decisão mais recente carregada", variant: "success" });
      } else {
        toast({ title: "Nenhuma decisão encontrada", variant: "warning" });
      }
    } catch (err: any) {
      toast({ title: "Erro ao buscar decisão", description: err?.message, variant: "error" });
    } finally {
      setFetching(false);
    }
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!decisionId.trim()) {
      toast({ title: "decisionId é obrigatório", variant: "warning" });
      return;
    }
    setLoading(true);
    setResult(null);
    setExpanded(new Set());
    try {
      const res = await api.abduceReasoning({
        decisionId: decisionId.trim(),
        question: question.trim() || undefined,
        maxHypotheses,
      });
      setResult(res);
    } catch (err: any) {
      toast({ title: "Erro no raciocínio abdutivo", description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  };

  const toggle = (i: number) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(i)) next.delete(i);
      else next.add(i);
      return next;
    });
  };

  const sortedHypotheses: AbductionHypothesis[] = result?.hypotheses
    ? [...result.hypotheses].sort((a, b) => (b.confidence ?? 0) - (a.confidence ?? 0))
    : [];

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center gap-3">
            <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
              <Brain className="h-5 w-5 text-white" />
            </div>
            <div>
              <h1 className="text-base font-bold leading-tight">{t("nav.reasoning")}</h1>
              <p className="text-xs text-white/80">Raciocínio abdutivo sobre decisões</p>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 max-w-5xl space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm flex items-center gap-2">
              <Wand2 className="h-4 w-4" /> Gerar hipóteses
            </CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={submit} className="space-y-3 text-sm">
              <div>
                <label className="text-[10px] uppercase tracking-wider text-muted-foreground">DecisionID</label>
                <div className="flex gap-2 mt-0.5">
                  <input
                    value={decisionId}
                    onChange={(e) => setDecisionId(e.target.value)}
                    placeholder="UUID"
                    className="flex-1 bg-transparent border rounded px-2 py-1 font-mono text-xs"
                  />
                  <Button type="button" size="sm" variant="outline" onClick={fetchLatest} disabled={fetching}>
                    {fetching ? <Spinner size="sm" /> : "Pegar mais recente"}
                  </Button>
                </div>
              </div>
              <div>
                <label className="text-[10px] uppercase tracking-wider text-muted-foreground">Pergunta (opcional)</label>
                <textarea
                  value={question}
                  onChange={(e) => setQuestion(e.target.value)}
                  rows={2}
                  className="w-full bg-transparent border rounded px-2 py-1 mt-0.5"
                  placeholder="Por que essa decisão foi tomada?"
                />
              </div>
              <div>
                <label className="text-[10px] uppercase tracking-wider text-muted-foreground">
                  Max hipóteses ({maxHypotheses})
                </label>
                <input
                  type="range"
                  min={1}
                  max={10}
                  step={1}
                  value={maxHypotheses}
                  onChange={(e) => setMaxHypotheses(Number(e.target.value))}
                  className="w-full accent-brain-primary mt-0.5"
                />
              </div>
              <Button
                type="submit"
                size="sm"
                disabled={loading}
                className="bg-gradient-to-r from-brain-primary to-brain-accent text-white"
              >
                {loading ? <Spinner size="sm" /> : <Play className="h-4 w-4 mr-1" />}
                Abduzir
              </Button>
            </form>
          </CardContent>
        </Card>

        {loading && (
          <div className="flex justify-center py-10"><Spinner size="lg" /></div>
        )}

        {!loading && !result && (
          <EmptyState
            icon={Brain}
            title="Sem resultados ainda"
            description="Informe um decisionId e clique em Abduzir para gerar hipóteses."
          />
        )}

        {result && (
          <>
            <Card className="border-l-4 border-brain-primary">
              <CardHeader>
                <CardTitle className="text-sm">Decisão alvo</CardTitle>
              </CardHeader>
              <CardContent className="text-xs space-y-2">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                  <Row label="ID" value={result.decision?.id ?? ""} mono />
                  <Row label="Categoria" value={result.decision?.category ?? ""} />
                  <Row label="Resultado" value={result.decision?.outcome ?? ""} />
                  <Row
                    label="Confiança"
                    value={`${Math.round((result.decision?.confidence ?? 0) * 100)}%`}
                  />
                </div>
                <div>
                  <p className="text-[10px] uppercase tracking-wider text-muted-foreground">Cenário</p>
                  <p>{result.decision?.scenario}</p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm">
                  Hipóteses ({sortedHypotheses.length})
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {sortedHypotheses.length === 0 ? (
                  <p className="text-xs text-muted-foreground">Nenhuma hipótese gerada.</p>
                ) : (
                  sortedHypotheses.map((h, i) => {
                    const isOpen = expanded.has(i);
                    const pct = Math.round((h.confidence ?? 0) * 100);
                    return (
                      <div key={i} className="border rounded p-3 text-xs">
                        <div className="flex items-center justify-between">
                          <div className="flex-1 min-w-0">
                            <p className="font-medium">{h.cause}</p>
                            <div className="mt-1.5 h-1.5 bg-muted rounded overflow-hidden">
                              <div
                                className="h-full bg-gradient-to-r from-brain-primary to-brain-accent"
                                style={{ width: `${pct}%` }}
                              />
                            </div>
                            <p className="text-[10px] text-muted-foreground mt-0.5">
                              Confiança: {pct}%
                            </p>
                          </div>
                          <Button size="sm" variant="ghost" className="h-7 w-7 p-0 ml-2" onClick={() => toggle(i)}>
                            {isOpen ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
                          </Button>
                        </div>

                        {isOpen && (
                          <div className="mt-3 pt-3 border-t space-y-2">
                            {h.evidence && h.evidence.length > 0 && (
                              <div>
                                <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                                  Evidências
                                </p>
                                <ul className="list-disc ml-4 space-y-0.5">
                                  {h.evidence.map((ev, k) => (
                                    <li key={k}>{ev}</li>
                                  ))}
                                </ul>
                              </div>
                            )}
                            {h.memoryIds && h.memoryIds.length > 0 && (
                              <div>
                                <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                                  Memórias
                                </p>
                                <div className="flex flex-wrap gap-1">
                                  {h.memoryIds.map((mid) => (
                                    <span
                                      key={mid}
                                      className="px-1.5 py-0.5 rounded-full bg-muted font-mono text-[10px]"
                                    >
                                      {mid.slice(0, 8)}
                                    </span>
                                  ))}
                                </div>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    );
                  })
                )}
              </CardContent>
            </Card>
          </>
        )}
      </main>
    </div>
  );
}

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <p className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</p>
      <p className={mono ? "font-mono break-all" : ""}>{value}</p>
    </div>
  );
}
