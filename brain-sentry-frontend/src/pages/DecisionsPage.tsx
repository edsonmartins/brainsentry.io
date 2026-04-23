import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  Scale, RefreshCw, Save, ChevronRight, Network, History, AlertCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import {
  api,
  type Decision,
  type DecisionOutcome,
  type DecisionPrecedent,
  type CausalNode,
  type RecordDecisionRequest,
} from "@/lib/api/client";

const OUTCOME_COLORS: Record<DecisionOutcome, string> = {
  approved: "#10b981",
  rejected: "#ef4444",
  deferred: "#f59e0b",
  pending: "#6b7280",
};

export default function DecisionsPage() {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();

  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [selected, setSelected] = useState<Decision | null>(null);
  const [precedents, setPrecedents] = useState<DecisionPrecedent[] | null>(null);
  const [causalChain, setCausalChain] = useState<CausalNode[] | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  const [form, setForm] = useState<RecordDecisionRequest>({
    category: "",
    scenario: "",
    reasoning: "",
    outcome: "pending",
    confidence: 0.8,
  });

  const loadDecisions = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.listDecisions({ limit: 50 });
      setDecisions(res.decisions || []);
    } catch (err: any) {
      toast({ title: "Falha ao carregar decisões", description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    loadDecisions();
  }, [loadDecisions]);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.category.trim() || !form.scenario.trim() || !form.reasoning.trim()) {
      toast({ title: "Campos obrigatórios ausentes", variant: "warning" });
      return;
    }
    setSaving(true);
    try {
      await api.recordDecision(form);
      toast({ title: "Decisão registrada", variant: "success" });
      setForm({ category: "", scenario: "", reasoning: "", outcome: "pending", confidence: 0.8 });
      await loadDecisions();
    } catch (err: any) {
      toast({ title: "Erro ao registrar decisão", description: err?.message, variant: "error" });
    } finally {
      setSaving(false);
    }
  };

  const openDetail = async (d: Decision) => {
    setSelected(d);
    setPrecedents(null);
    setCausalChain(null);
  };

  const loadPrecedents = async () => {
    if (!selected) return;
    setDetailLoading(true);
    try {
      const res = await api.findDecisionPrecedents(selected.id, 5);
      setPrecedents(res.precedents || []);
    } catch (err: any) {
      toast({ title: "Erro ao buscar precedentes", description: err?.message, variant: "error" });
    } finally {
      setDetailLoading(false);
    }
  };

  const loadCausalChain = async () => {
    if (!selected) return;
    setDetailLoading(true);
    try {
      const res = await api.getDecisionCausalChain(selected.id, 5);
      setCausalChain(res.chain || []);
    } catch (err: any) {
      toast({ title: "Erro ao buscar cadeia causal", description: err?.message, variant: "error" });
    } finally {
      setDetailLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Scale className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("nav.decisions")}</h1>
                <p className="text-xs text-white/80">Registro, precedentes e cadeia causal</p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="bg-white/20 border-white/30 text-white hover:bg-white/30"
              onClick={loadDecisions}
              disabled={loading}
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
          <div className="lg:col-span-2">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">Registrar decisão</CardTitle>
              </CardHeader>
              <CardContent>
                <form onSubmit={submit} className="space-y-3 text-sm">
                  <Field label="Categoria">
                    <input
                      value={form.category}
                      onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
                      className="w-full bg-transparent border rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-brain-primary"
                      placeholder="Ex: deploy"
                    />
                  </Field>
                  <Field label="Cenário">
                    <input
                      value={form.scenario}
                      onChange={(e) => setForm((f) => ({ ...f, scenario: e.target.value }))}
                      className="w-full bg-transparent border rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-brain-primary"
                      placeholder="Curto descritivo"
                    />
                  </Field>
                  <Field label="Raciocínio">
                    <textarea
                      value={form.reasoning}
                      onChange={(e) => setForm((f) => ({ ...f, reasoning: e.target.value }))}
                      rows={4}
                      className="w-full bg-transparent border rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-brain-primary"
                    />
                  </Field>
                  <div className="grid grid-cols-2 gap-3">
                    <Field label="Resultado">
                      <select
                        value={form.outcome}
                        onChange={(e) =>
                          setForm((f) => ({ ...f, outcome: e.target.value as DecisionOutcome }))
                        }
                        className="w-full bg-transparent border rounded px-2 py-1"
                      >
                        <option value="pending">pending</option>
                        <option value="approved">approved</option>
                        <option value="rejected">rejected</option>
                        <option value="deferred">deferred</option>
                      </select>
                    </Field>
                    <Field label={`Confiança (${Math.round((form.confidence ?? 0) * 100)}%)`}>
                      <input
                        type="range"
                        min={0}
                        max={1}
                        step={0.05}
                        value={form.confidence ?? 0}
                        onChange={(e) =>
                          setForm((f) => ({ ...f, confidence: Number(e.target.value) }))
                        }
                        className="w-full accent-brain-primary"
                      />
                    </Field>
                  </div>
                  <Field label="Decisão-pai (opcional)">
                    <input
                      value={form.parentDecisionId ?? ""}
                      onChange={(e) =>
                        setForm((f) => ({ ...f, parentDecisionId: e.target.value || undefined }))
                      }
                      className="w-full bg-transparent border rounded px-2 py-1 font-mono text-xs"
                      placeholder="UUID"
                    />
                  </Field>
                  <Button
                    type="submit"
                    disabled={saving}
                    size="sm"
                    className="bg-gradient-to-r from-brain-primary to-brain-accent text-white"
                  >
                    {saving ? <Spinner size="sm" /> : <Save className="h-4 w-4 mr-1" />}
                    Registrar
                  </Button>
                </form>
              </CardContent>
            </Card>
          </div>

          <div className="lg:col-span-3 space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">Decisões recentes</CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                {loading ? (
                  <div className="flex justify-center py-12"><Spinner size="lg" /></div>
                ) : decisions.length === 0 ? (
                  <EmptyState icon={Scale} title="Nenhuma decisão" description="Registre a primeira decisão no formulário." />
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full text-xs">
                      <thead className="border-b bg-muted/40">
                        <tr>
                          <th className="text-left p-2">Categoria</th>
                          <th className="text-left p-2">Cenário</th>
                          <th className="text-left p-2">Resultado</th>
                          <th className="text-left p-2">Confiança</th>
                          <th className="text-left p-2">Criada em</th>
                        </tr>
                      </thead>
                      <tbody>
                        {decisions.map((d) => (
                          <tr
                            key={d.id}
                            onClick={() => openDetail(d)}
                            className={`border-b cursor-pointer hover:bg-muted/30 ${
                              selected?.id === d.id ? "bg-muted/40" : ""
                            }`}
                          >
                            <td className="p-2 font-mono">{d.category}</td>
                            <td className="p-2 max-w-[220px] truncate" title={d.scenario}>{d.scenario}</td>
                            <td className="p-2">
                              <span
                                className="inline-block px-1.5 py-0.5 rounded text-[10px] font-medium text-white"
                                style={{ background: OUTCOME_COLORS[d.outcome] }}
                              >
                                {d.outcome}
                              </span>
                            </td>
                            <td className="p-2 font-mono">{Math.round((d.confidence ?? 0) * 100)}%</td>
                            <td className="p-2 text-muted-foreground">
                              {new Date(d.createdAt).toLocaleString(i18n.language)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </CardContent>
            </Card>

            {selected && (
              <Card className="border-brain-primary/30">
                <CardHeader>
                  <CardTitle className="text-sm flex items-center gap-2">
                    <ChevronRight className="h-4 w-4" /> Detalhe da decisão
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-xs">
                  <div>
                    <p className="text-[10px] uppercase tracking-wider text-muted-foreground">ID</p>
                    <p className="font-mono break-all">{selected.id}</p>
                  </div>
                  <div>
                    <p className="text-[10px] uppercase tracking-wider text-muted-foreground">Raciocínio</p>
                    <p className="whitespace-pre-wrap">{selected.reasoning}</p>
                  </div>
                  {selected.policyViolations && selected.policyViolations.length > 0 && (
                    <div>
                      <p className="text-[10px] uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                        <AlertCircle className="h-3 w-3 text-destructive" /> Violações de política
                      </p>
                      <div className="flex flex-wrap gap-1 mt-1">
                        {selected.policyViolations.map((v, i) => (
                          <span key={i} className="px-1.5 py-0.5 rounded bg-red-500/15 text-red-500 border border-red-500/30">
                            {v}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                  {selected.metadata && Object.keys(selected.metadata).length > 0 && (
                    <div>
                      <p className="text-[10px] uppercase tracking-wider text-muted-foreground">Metadata</p>
                      <pre className="bg-muted p-2 rounded text-[11px] overflow-auto max-h-40">
                        {JSON.stringify(selected.metadata, null, 2)}
                      </pre>
                    </div>
                  )}

                  <div className="flex gap-2 flex-wrap">
                    <Button size="sm" variant="outline" onClick={loadPrecedents} disabled={detailLoading}>
                      <History className="h-3.5 w-3.5 mr-1" /> Ver Precedentes
                    </Button>
                    <Button size="sm" variant="outline" onClick={loadCausalChain} disabled={detailLoading}>
                      <Network className="h-3.5 w-3.5 mr-1" /> Cadeia Causal
                    </Button>
                  </div>

                  {detailLoading && <div className="flex justify-center py-4"><Spinner size="sm" /></div>}

                  {precedents && (
                    <div>
                      <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                        Precedentes ({precedents.length})
                      </p>
                      {precedents.length === 0 ? (
                        <p className="text-muted-foreground">Nenhum precedente encontrado.</p>
                      ) : (
                        <ul className="space-y-1">
                          {precedents.map((p, i) => (
                            <li key={i} className="border rounded p-2">
                              <div className="flex items-center justify-between">
                                <span className="font-medium">{p.decision.scenario}</span>
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted">
                                  {Math.round(p.similarity * 100)}%
                                </span>
                              </div>
                              <p className="text-muted-foreground">
                                {p.decision.category} · {p.decision.outcome}
                              </p>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  )}

                  {causalChain && (
                    <div>
                      <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                        Cadeia causal ({causalChain.length})
                      </p>
                      {causalChain.length === 0 ? (
                        <p className="text-muted-foreground">Sem nós causais.</p>
                      ) : (
                        <ul className="space-y-1">
                          {causalChain.map((n, i) => (
                            <li
                              key={i}
                              className="border-l-2 pl-2 py-1"
                              style={{
                                marginLeft: `${n.depth * 12}px`,
                                borderColor: OUTCOME_COLORS[n.decision.outcome],
                              }}
                            >
                              <div className="flex items-center gap-2">
                                <span className="text-[10px] px-1 py-0.5 rounded bg-muted">
                                  {n.relation}
                                </span>
                                <span className="font-medium">{n.decision.scenario}</span>
                              </div>
                              <p className="text-muted-foreground">
                                {n.decision.category} · {n.decision.outcome}
                              </p>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</span>
      <div className="mt-0.5">{children}</div>
    </label>
  );
}
