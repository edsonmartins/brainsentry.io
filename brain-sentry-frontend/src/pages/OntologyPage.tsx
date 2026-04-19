import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  BookOpen, Save, RefreshCw, Search, AlertCircle,
  FileJson, CheckCircle, XCircle, Code,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import { api, type OntologyResolveResponse } from "@/lib/api/client";

interface OntologyEntityType {
  name: string;
  description?: string;
  parentType?: string;
  examples?: string[];
}

interface OntologyEntity {
  name: string;
  type: string;
  aliases?: string[];
}

interface OntologyRelationship {
  name: string;
  sourceType: string;
  targetType: string;
  symmetric?: boolean;
}

interface Ontology {
  name: string;
  version: string;
  entityTypes: OntologyEntityType[];
  entities?: OntologyEntity[];
  relationships: OntologyRelationship[];
}

const EMPTY_ONTOLOGY: Ontology = {
  name: "brainsentry",
  version: "1.0",
  entityTypes: [
    { name: "TECHNOLOGY", description: "Technologies and tools" },
    { name: "PERSON" },
    { name: "LANGUAGE" },
  ],
  entities: [
    { name: "PostgreSQL", type: "TECHNOLOGY", aliases: ["postgres", "psql"] },
    { name: "Go", type: "LANGUAGE", aliases: ["golang"] },
  ],
  relationships: [
    { name: "uses", sourceType: "*", targetType: "TECHNOLOGY" },
    { name: "implements", sourceType: "LANGUAGE", targetType: "TECHNOLOGY" },
  ],
};

export default function OntologyPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [ontology, setOntology] = useState<Ontology | null>(null);
  const [rawEditor, setRawEditor] = useState("");
  const [editMode, setEditMode] = useState<"visual" | "json">("visual");
  const [jsonError, setJsonError] = useState<string | null>(null);

  const [resolveInput, setResolveInput] = useState("");
  const [resolveResult, setResolveResult] = useState<OntologyResolveResponse | null>(null);
  const [resolving, setResolving] = useState(false);

  const loadOntology = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getOntology();
      setOntology(data);
      setRawEditor(JSON.stringify(data, null, 2));
    } catch {
      setOntology(null);
      setRawEditor(JSON.stringify(EMPTY_ONTOLOGY, null, 2));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadOntology();
  }, [loadOntology]);

  const save = async () => {
    let payload: Ontology;
    try {
      payload = JSON.parse(rawEditor) as Ontology;
    } catch (e: any) {
      setJsonError(e.message);
      toast({ title: t("ontology.invalidJson"), description: e.message, variant: "error" });
      return;
    }
    setJsonError(null);
    setSaving(true);
    try {
      await api.setOntology(payload as unknown as Record<string, any>);
      toast({ title: t("ontology.saved"), variant: "success" });
      setOntology(payload);
    } catch (err: any) {
      toast({ title: t("ontology.saveFailed"), description: err?.message, variant: "error" });
    } finally {
      setSaving(false);
    }
  };

  const resolve = async () => {
    if (!resolveInput.trim()) return;
    setResolving(true);
    try {
      const r = await api.resolveOntologyEntity(resolveInput);
      setResolveResult(r);
    } catch {
      setResolveResult({ input: resolveInput, matched: false, canonical: "", type: "" });
    } finally {
      setResolving(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <BookOpen className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("ontology.title")}</h1>
                <p className="text-xs text-white/80">{t("ontology.subtitle")}</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                className="bg-white/20 border-white/30 text-white hover:bg-white/30"
                onClick={loadOntology}
              >
                <RefreshCw className="h-4 w-4" />
              </Button>
              <Button
                size="sm"
                className="bg-white text-brain-primary hover:bg-white/90"
                onClick={save}
                disabled={saving}
              >
                {saving ? <Spinner size="sm" /> : <Save className="h-4 w-4 mr-2" />}
                {t("ontology.save")}
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 max-w-6xl">
        {loading ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 space-y-4">
              {!ontology && (
                <Card className="border-dashed">
                  <CardContent className="p-4 flex items-start gap-2">
                    <AlertCircle className="h-4 w-4 text-yellow-500 mt-0.5" />
                    <div>
                      <p className="text-sm font-medium">{t("ontology.noOntology")}</p>
                      <p className="text-xs text-muted-foreground">
                        {t("ontology.noOntologyDesc")}
                        <code className="mx-1 px-1 py-0.5 rounded bg-muted">ONTOLOGY_PATH</code>
                        {t("ontology.envVarSuffix")}
                      </p>
                    </div>
                  </CardContent>
                </Card>
              )}

              <div className="inline-flex rounded-lg bg-muted p-0.5">
                {([
                  { k: "visual", icon: BookOpen },
                  { k: "json", icon: FileJson },
                ] as const).map((it) => (
                  <button
                    key={it.k}
                    onClick={() => setEditMode(it.k)}
                    className={`px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${
                      editMode === it.k
                        ? "bg-background text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    <it.icon className="h-3.5 w-3.5 inline mr-1" />
                    {t(`ontology.tabs.${it.k}`)}
                  </button>
                ))}
              </div>

              {editMode === "json" ? (
                <Card>
                  <CardContent className="p-0">
                    <textarea
                      value={rawEditor}
                      onChange={(e) => {
                        setRawEditor(e.target.value);
                        setJsonError(null);
                      }}
                      className="w-full h-[480px] p-3 font-mono text-xs bg-transparent outline-none resize-y"
                    />
                    {jsonError && (
                      <div className="p-2 border-t text-xs text-destructive bg-destructive/5">
                        {jsonError}
                      </div>
                    )}
                  </CardContent>
                </Card>
              ) : (
                <VisualView ontology={parseOrNull(rawEditor)} />
              )}
            </div>

            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm flex items-center gap-2">
                    <Search className="h-4 w-4" /> {t("ontology.resolve")}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-[10px] text-muted-foreground mb-2">
                    {t("ontology.resolveDesc")}
                  </p>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={resolveInput}
                      onChange={(e) => setResolveInput(e.target.value)}
                      onKeyDown={(e) => e.key === "Enter" && resolve()}
                      placeholder={t("ontology.placeholder")}
                      className="flex-1 text-sm bg-transparent border rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-brain-primary"
                    />
                    <Button size="sm" onClick={resolve} disabled={resolving}>
                      {resolving ? <Spinner size="sm" /> : t("ontology.go")}
                    </Button>
                  </div>
                  {resolveResult && (
                    <div className="mt-3 text-xs">
                      {resolveResult.matched ? (
                        <div className="flex items-start gap-2">
                          <CheckCircle className="h-4 w-4 text-green-500 mt-0.5 flex-shrink-0" />
                          <div>
                            <p>
                              <span className="text-muted-foreground">{t("ontology.canonical")}</span>{" "}
                              <span className="font-medium">{resolveResult.canonical}</span>
                            </p>
                            <p>
                              <span className="text-muted-foreground">{t("ontology.type")}</span>{" "}
                              <span className="font-mono">{resolveResult.type}</span>
                            </p>
                          </div>
                        </div>
                      ) : (
                        <div className="flex items-start gap-2">
                          <XCircle className="h-4 w-4 text-muted-foreground mt-0.5 flex-shrink-0" />
                          <p className="text-muted-foreground">{t("ontology.noMatch")}</p>
                        </div>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Stats ontology={parseOrNull(rawEditor)} />
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

function parseOrNull(raw: string): Ontology | null {
  try {
    return JSON.parse(raw) as Ontology;
  } catch {
    return null;
  }
}

function Stats({ ontology }: { ontology: Ontology | null }) {
  const { t } = useTranslation();
  if (!ontology) return null;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm">{t("ontology.stats")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 text-xs">
        <Row label={t("ontology.nameLabel")} value={ontology.name} />
        <Row label={t("ontology.versionLabel")} value={ontology.version} />
        <Row label={t("ontology.entityTypes")} value={ontology.entityTypes?.length ?? 0} />
        <Row label={t("ontology.entities")} value={ontology.entities?.length ?? 0} />
        <Row label={t("ontology.relationships")} value={ontology.relationships?.length ?? 0} />
      </CardContent>
    </Card>
  );
}

function Row({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex justify-between">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-mono">{value}</span>
    </div>
  );
}

function VisualView({ ontology }: { ontology: Ontology | null }) {
  const { t } = useTranslation();

  if (!ontology) {
    return (
      <Card>
        <CardContent className="p-4">
          <p className="text-xs text-destructive flex items-center gap-2">
            <Code className="h-4 w-4" />
            {t("ontology.invalidJsonFix")}
          </p>
        </CardContent>
      </Card>
    );
  }

  if (ontology.entityTypes?.length === 0 && ontology.relationships?.length === 0) {
    return (
      <EmptyState
        icon={BookOpen}
        title={t("ontology.emptyTitle")}
        description={t("ontology.emptyDesc")}
      />
    );
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">
            {t("ontology.entityTypes")} <span className="text-[10px] text-muted-foreground">({ontology.entityTypes?.length ?? 0})</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          {(ontology.entityTypes || []).map((et) => (
            <div key={et.name} className="px-3 py-2 rounded-lg border">
              <p className="text-sm font-medium">{et.name}</p>
              {et.description && <p className="text-[10px] text-muted-foreground mt-0.5">{et.description}</p>}
            </div>
          ))}
        </CardContent>
      </Card>

      {ontology.entities && ontology.entities.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">
              {t("ontology.canonicalEntities")} <span className="text-[10px] text-muted-foreground">({ontology.entities.length})</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-1.5">
            {ontology.entities.map((e, idx) => (
              <div key={idx} className="flex items-center gap-2 p-1.5 rounded border">
                <span className="text-sm font-medium">{e.name}</span>
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">{e.type}</span>
                {e.aliases && e.aliases.length > 0 && (
                  <div className="flex items-center gap-1 text-[10px] text-muted-foreground">
                    <span>{t("ontology.aka")}</span>
                    {e.aliases.map((a) => (
                      <span key={a} className="px-1 rounded bg-muted/60">
                        {a}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">
            {t("ontology.allowedRelationships")} <span className="text-[10px] text-muted-foreground">({ontology.relationships?.length ?? 0})</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-1.5">
          {(ontology.relationships || []).map((r, idx) => (
            <div key={idx} className="flex items-center gap-2 p-1.5 rounded border text-sm">
              <span className="text-blue-500 font-mono text-xs">{r.sourceType}</span>
              <span className="text-muted-foreground">—</span>
              <span className="px-1.5 py-0.5 rounded bg-brain-accent/15 text-brain-accent font-mono text-xs">
                {r.name}
              </span>
              <span className="text-muted-foreground">→</span>
              <span className="text-green-500 font-mono text-xs">{r.targetType}</span>
              {r.symmetric && (
                <span className="text-[10px] px-1 py-0.5 rounded bg-muted ml-auto">{t("ontology.symmetric")}</span>
              )}
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
