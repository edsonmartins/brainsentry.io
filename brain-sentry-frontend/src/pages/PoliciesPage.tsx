import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  ShieldCheck, RefreshCw, Save, Trash2, Power, PlayCircle, AlertCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import {
  api,
  type Policy,
  type PolicySeverity,
  type PolicyRuleType,
  type PolicyViolation,
  type CreatePolicyRequest,
} from "@/lib/api/client";

const SEVERITY_COLORS: Record<PolicySeverity, string> = {
  info: "#3b82f6",
  warning: "#f59e0b",
  error: "#ef4444",
  critical: "#991b1b",
};

const RULE_TYPES: PolicyRuleType[] = [
  "min_confidence",
  "requires_memory",
  "requires_entity",
  "forbidden_outcome",
  "requires_reasoning",
  "category_blocked",
];

export default function PoliciesPage() {
  const { t } = useTranslation();
  const { toast } = useToast();

  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const [form, setForm] = useState<CreatePolicyRequest>({
    name: "",
    description: "",
    category: "*",
    severity: "warning",
    ruleType: "min_confidence",
    ruleConfig: { min: 0.7 },
    enabled: true,
  });
  const [ruleConfigRaw, setRuleConfigRaw] = useState(JSON.stringify({ min: 0.7 }, null, 2));
  const [ruleConfigError, setRuleConfigError] = useState<string | null>(null);

  const [enforceDecisionId, setEnforceDecisionId] = useState("");
  const [enforceResult, setEnforceResult] = useState<{ violations: PolicyViolation[]; compliant: boolean } | null>(null);
  const [enforcing, setEnforcing] = useState(false);

  const loadPolicies = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.listPolicies();
      setPolicies(res.policies || []);
    } catch (err: any) {
      toast({ title: "Falha ao carregar políticas", description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    loadPolicies();
  }, [loadPolicies]);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) {
      toast({ title: "Nome obrigatório", variant: "warning" });
      return;
    }
    let cfg: Record<string, unknown>;
    try {
      cfg = JSON.parse(ruleConfigRaw);
    } catch (err: any) {
      setRuleConfigError(err.message);
      return;
    }
    setRuleConfigError(null);
    setSaving(true);
    try {
      await api.createPolicy({ ...form, ruleConfig: cfg });
      toast({ title: "Política criada", variant: "success" });
      setForm({
        name: "",
        description: "",
        category: "*",
        severity: "warning",
        ruleType: "min_confidence",
        ruleConfig: { min: 0.7 },
        enabled: true,
      });
      setRuleConfigRaw(JSON.stringify({ min: 0.7 }, null, 2));
      await loadPolicies();
    } catch (err: any) {
      toast({ title: "Erro ao criar política", description: err?.message, variant: "error" });
    } finally {
      setSaving(false);
    }
  };

  const toggle = async (p: Policy) => {
    try {
      await api.updatePolicy(p.id, {
        name: p.name,
        description: p.description,
        category: p.category,
        severity: p.severity,
        ruleType: p.ruleType,
        ruleConfig: p.ruleConfig,
        enabled: !p.enabled,
      });
      toast({ title: p.enabled ? "Política desativada" : "Política ativada", variant: "success" });
      await loadPolicies();
    } catch (err: any) {
      toast({ title: "Erro ao alternar", description: err?.message, variant: "error" });
    }
  };

  const remove = async (p: Policy) => {
    if (!confirm(`Remover política "${p.name}"?`)) return;
    try {
      await api.deletePolicy(p.id);
      toast({ title: "Política removida", variant: "success" });
      await loadPolicies();
    } catch (err: any) {
      toast({ title: "Erro ao remover", description: err?.message, variant: "error" });
    }
  };

  const enforce = async () => {
    if (!enforceDecisionId.trim()) return;
    setEnforcing(true);
    try {
      const res = await api.enforcePolicy(enforceDecisionId.trim());
      setEnforceResult({ violations: res.violations || [], compliant: res.compliant });
    } catch (err: any) {
      toast({ title: "Erro ao verificar", description: err?.message, variant: "error" });
      setEnforceResult(null);
    } finally {
      setEnforcing(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <ShieldCheck className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("nav.policies")}</h1>
                <p className="text-xs text-white/80">Regras e compliance sobre decisões</p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="bg-white/20 border-white/30 text-white hover:bg-white/30"
              onClick={loadPolicies}
              disabled={loading}
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Criar política</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={submit} className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
              <Field label="Nome">
                <input
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  className="w-full bg-transparent border rounded px-2 py-1"
                />
              </Field>
              <Field label="Categoria">
                <input
                  value={form.category}
                  onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
                  className="w-full bg-transparent border rounded px-2 py-1"
                  placeholder="* ou específica"
                />
              </Field>
              <Field label="Descrição">
                <input
                  value={form.description ?? ""}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  className="w-full bg-transparent border rounded px-2 py-1"
                />
              </Field>
              <Field label="Severidade">
                <select
                  value={form.severity}
                  onChange={(e) => setForm((f) => ({ ...f, severity: e.target.value as PolicySeverity }))}
                  className="w-full bg-transparent border rounded px-2 py-1"
                >
                  <option value="info">info</option>
                  <option value="warning">warning</option>
                  <option value="error">error</option>
                  <option value="critical">critical</option>
                </select>
              </Field>
              <Field label="Tipo de regra">
                <select
                  value={form.ruleType}
                  onChange={(e) => setForm((f) => ({ ...f, ruleType: e.target.value as PolicyRuleType }))}
                  className="w-full bg-transparent border rounded px-2 py-1"
                >
                  {RULE_TYPES.map((rt) => (
                    <option key={rt} value={rt}>{rt}</option>
                  ))}
                </select>
              </Field>
              <Field label="Habilitada">
                <label className="flex items-center gap-2 mt-1">
                  <input
                    type="checkbox"
                    checked={!!form.enabled}
                    onChange={(e) => setForm((f) => ({ ...f, enabled: e.target.checked }))}
                  />
                  <span className="text-xs">enabled</span>
                </label>
              </Field>
              <div className="md:col-span-2">
                <Field label="Rule config (JSON)">
                  <textarea
                    value={ruleConfigRaw}
                    onChange={(e) => {
                      setRuleConfigRaw(e.target.value);
                      setRuleConfigError(null);
                    }}
                    rows={4}
                    className="w-full bg-transparent border rounded px-2 py-1 font-mono text-xs"
                  />
                  {ruleConfigError && (
                    <p className="text-xs text-destructive mt-1">{ruleConfigError}</p>
                  )}
                </Field>
              </div>
              <div className="md:col-span-2">
                <Button
                  type="submit"
                  size="sm"
                  disabled={saving}
                  className="bg-gradient-to-r from-brain-primary to-brain-accent text-white"
                >
                  {saving ? <Spinner size="sm" /> : <Save className="h-4 w-4 mr-1" />}
                  Criar política
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Políticas ({policies.length})</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {loading ? (
              <div className="flex justify-center py-10"><Spinner size="lg" /></div>
            ) : policies.length === 0 ? (
              <EmptyState icon={ShieldCheck} title="Nenhuma política" description="Crie uma política no formulário acima." />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead className="border-b bg-muted/40">
                    <tr>
                      <th className="text-left p-2">Nome</th>
                      <th className="text-left p-2">Categoria</th>
                      <th className="text-left p-2">Tipo</th>
                      <th className="text-left p-2">Severidade</th>
                      <th className="text-left p-2">Status</th>
                      <th className="text-left p-2">Ações</th>
                    </tr>
                  </thead>
                  <tbody>
                    {policies.map((p) => (
                      <tr key={p.id} className="border-b hover:bg-muted/30">
                        <td className="p-2">
                          <p className="font-medium">{p.name}</p>
                          <p className="text-[10px] text-muted-foreground">{p.description}</p>
                        </td>
                        <td className="p-2 font-mono">{p.category}</td>
                        <td className="p-2 font-mono">{p.ruleType}</td>
                        <td className="p-2">
                          <span
                            className="px-1.5 py-0.5 rounded text-[10px] font-medium text-white"
                            style={{ background: SEVERITY_COLORS[p.severity] }}
                          >
                            {p.severity}
                          </span>
                        </td>
                        <td className="p-2">
                          {p.enabled ? (
                            <span className="text-green-500">ativa</span>
                          ) : (
                            <span className="text-muted-foreground">inativa</span>
                          )}
                        </td>
                        <td className="p-2">
                          <div className="flex gap-1">
                            <Button size="sm" variant="ghost" className="h-7 w-7 p-0" onClick={() => toggle(p)}>
                              <Power className="h-3.5 w-3.5" />
                            </Button>
                            <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-destructive" onClick={() => remove(p)}>
                              <Trash2 className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm flex items-center gap-2">
              <PlayCircle className="h-4 w-4" /> Aplicar política a uma decisão
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="flex gap-2">
              <input
                value={enforceDecisionId}
                onChange={(e) => setEnforceDecisionId(e.target.value)}
                placeholder="decisionId (UUID)"
                className="flex-1 bg-transparent border rounded px-2 py-1 font-mono text-xs"
              />
              <Button size="sm" onClick={enforce} disabled={enforcing || !enforceDecisionId.trim()}>
                {enforcing ? <Spinner size="sm" /> : "Verificar"}
              </Button>
            </div>

            {enforceResult && (
              <div>
                <p className="text-xs mb-2">
                  Compliant:{" "}
                  <span className={enforceResult.compliant ? "text-green-500 font-medium" : "text-destructive font-medium"}>
                    {enforceResult.compliant ? "sim" : "não"}
                  </span>
                </p>
                {enforceResult.violations.length === 0 ? (
                  <p className="text-xs text-muted-foreground">Nenhuma violação.</p>
                ) : (
                  <ul className="space-y-1">
                    {enforceResult.violations.map((v, i) => (
                      <li key={i} className="flex items-start gap-2 border rounded p-2 text-xs">
                        <AlertCircle className="h-3.5 w-3.5 mt-0.5" style={{ color: SEVERITY_COLORS[v.severity] }} />
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{v.policyName}</span>
                            <span
                              className="px-1.5 py-0.5 rounded text-[10px] font-medium text-white"
                              style={{ background: SEVERITY_COLORS[v.severity] }}
                            >
                              {v.severity}
                            </span>
                          </div>
                          <p className="text-muted-foreground">{v.message}</p>
                        </div>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}
          </CardContent>
        </Card>
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
