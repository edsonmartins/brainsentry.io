import { useState, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  Sparkles, Layers, FlaskConical, Network, ArrowRight, Loader2,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StrengthBar } from "@/components/ui/StrengthBar";
import { TripletsViewer } from "@/components/memory/TripletsViewer";
import { useToast } from "@/components/ui/toast";
import { api, type CascadeExtractResponse } from "@/lib/api/client";

const SAMPLE_TEXT = `PostgreSQL is a relational database developed by the PostgreSQL Global Development Group.
It supports JSON via the jsonb type and is widely used with Go web services.
The pgvector extension adds vector similarity search for AI applications.`;

export default function ExtractionLabPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [input, setInput] = useState(SAMPLE_TEXT);
  const [tab, setTab] = useState<"triplets" | "cascade">("triplets");

  // Cascade-specific state
  const [cascadeResult, setCascadeResult] = useState<CascadeExtractResponse | null>(null);
  const [cascadeLoading, setCascadeLoading] = useState(false);

  // Triplet trigger — we pass `content` prop + a key that changes to force re-extract.
  const [tripletKey, setTripletKey] = useState(0);

  const runCascade = useCallback(async () => {
    if (!input.trim()) return;
    setCascadeLoading(true);
    try {
      const resp = await api.cascadeExtract(input);
      setCascadeResult(resp);
    } catch (err: any) {
      toast({
        title: t("extraction.extractionFailed"),
        description: err?.message || t("extraction.extractionFailedDesc"),
        variant: "error",
      });
    } finally {
      setCascadeLoading(false);
    }
  }, [input, toast, t]);

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center gap-3">
            <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
              <FlaskConical className="h-5 w-5 text-white" />
            </div>
            <div>
              <h1 className="text-base font-bold leading-tight">{t("extraction.title")}</h1>
              <p className="text-xs text-white/80">{t("extraction.subtitle")}</p>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 max-w-6xl">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Input column */}
          <div>
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">{t("extraction.content")}</CardTitle>
              </CardHeader>
              <CardContent>
                <textarea
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  placeholder={t("extraction.contentPlaceholder")}
                  className="w-full bg-transparent resize-y outline-none text-sm min-h-[240px] focus:ring-1 focus:ring-brain-primary rounded border p-3"
                />
                <p className="text-[10px] text-muted-foreground mt-2">
                  {t("extraction.charsHint", { count: input.length })}
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Output column */}
          <div>
            {/* Tab switch */}
            <div className="inline-flex rounded-lg bg-muted p-0.5 mb-3">
              {([
                { k: "triplets" as const, icon: Sparkles },
                { k: "cascade" as const, icon: Layers },
              ]).map((it) => (
                <button
                  key={it.k}
                  onClick={() => setTab(it.k)}
                  className={`px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${
                    tab === it.k
                      ? "bg-background text-foreground shadow-sm"
                      : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  <it.icon className="h-3.5 w-3.5 inline mr-1" />
                  {t(`extraction.tabs.${it.k}`)}
                </button>
              ))}
            </div>

            {tab === "triplets" ? (
              <div>
                <Button
                  size="sm"
                  onClick={() => setTripletKey((k) => k + 1)}
                  className="mb-3 bg-gradient-to-r from-brain-primary to-brain-accent text-white"
                  disabled={!input.trim()}
                >
                  <Sparkles className="h-3.5 w-3.5 mr-2" />
                  {t("extraction.tripletsButton")}
                </Button>
                {tripletKey > 0 && (
                  <TripletsViewer key={tripletKey} content={input} />
                )}
                {tripletKey === 0 && (
                  <EmptyState
                    icon={Sparkles}
                    title={t("extraction.emptyTriplets.title")}
                    description={t("extraction.emptyTriplets.desc")}
                  />
                )}
              </div>
            ) : (
              <div>
                <Button
                  size="sm"
                  onClick={runCascade}
                  className="mb-3 bg-gradient-to-r from-brain-primary to-brain-accent text-white"
                  disabled={!input.trim() || cascadeLoading}
                >
                  {cascadeLoading ? (
                    <><Loader2 className="h-3.5 w-3.5 mr-2 animate-spin" />{t("extraction.cascadeRunning")}</>
                  ) : (
                    <><Layers className="h-3.5 w-3.5 mr-2" />{t("extraction.cascadeButton")}</>
                  )}
                </Button>
                {cascadeResult ? (
                  <CascadeResult result={cascadeResult} />
                ) : (
                  <EmptyState
                    icon={Layers}
                    title={t("extraction.emptyCascade.title")}
                    description={t("extraction.emptyCascade.desc")}
                  />
                )}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

function CascadeResult({ result }: { result: CascadeExtractResponse }) {
  const { t } = useTranslation();
  const entities = result.entities || [];
  const relationships = result.relationships || [];

  const typeColors: Record<string, string> = {
    TECHNOLOGY: "#3b82f6", PERSON: "#10b981", PROJECT: "#f59e0b",
    CONCEPT: "#8b5cf6", LIBRARY: "#06b6d4", LANGUAGE: "#ec4899",
    TOOL: "#f97316", SERVICE: "#14b8a6", FILE: "#22c55e",
    FUNCTION: "#6366f1", ORGANIZATION: "#a855f7", LOCATION: "#eab308",
  };

  return (
    <div className="space-y-4">
      {/* Progress bar */}
      <Card>
        <CardContent className="p-3">
          <div className="flex items-center gap-2 text-xs">
            <span className="text-muted-foreground">{t("extraction.completed")}</span>
            <span className="font-mono">{result.passCount}/3</span>
            <span className="text-muted-foreground">{t("extraction.passes")}</span>
            <StrengthBar value={result.passCount} max={3} size="sm" showValue={false} />
          </div>
        </CardContent>
      </Card>

      {/* Entities */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center gap-2">
            <Network className="h-4 w-4" />
            {t("extraction.entities")}
            <span className="text-[10px] text-muted-foreground">({entities.length})</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {entities.length === 0 ? (
            <p className="text-xs text-muted-foreground">{t("extraction.noEntities")}</p>
          ) : (
            <div className="flex flex-wrap gap-1.5">
              {entities.map((e, idx) => {
                const c = typeColors[e.type] || "#6b7280";
                return (
                  <span
                    key={idx}
                    className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full border text-xs"
                    style={{ borderColor: c + "60", color: c }}
                  >
                    <span className="h-2 w-2 rounded-full" style={{ backgroundColor: c }} />
                    {e.name}
                    <span className="text-[10px] opacity-60">{e.type}</span>
                  </span>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Relationships */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center gap-2">
            <ArrowRight className="h-4 w-4" />
            {t("extraction.relationships")}
            <span className="text-[10px] text-muted-foreground">({relationships.length})</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {relationships.length === 0 ? (
            <p className="text-xs text-muted-foreground">{t("extraction.noRelationships")}</p>
          ) : (
            <div className="space-y-1.5">
              {relationships.map((r, idx) => (
                <div key={idx} className="flex items-center gap-2 text-sm p-1.5 rounded border">
                  <span className="font-medium text-blue-500">{r.source}</span>
                  <ArrowRight className="h-3 w-3 text-muted-foreground" />
                  <span className="text-xs px-1.5 py-0.5 rounded bg-brain-accent/15 text-brain-accent font-mono">
                    {r.type}
                  </span>
                  <ArrowRight className="h-3 w-3 text-muted-foreground" />
                  <span className="font-medium text-green-500">{r.target}</span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
