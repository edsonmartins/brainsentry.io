import { useState, useRef, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  Brain, Send, Sparkles, MessageSquare, Zap, Trash2, RefreshCw,
  TrendingUp, Target,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { StrengthBar } from "@/components/ui/StrengthBar";
import { useToast } from "@/components/ui/toast";
import { api, type RecallResult, type RouterDecision } from "@/lib/api/client";

type Mode = "remember" | "recall";

interface ConsoleEntry {
  id: string;
  mode: Mode;
  timestamp: Date;
  input: string;
  output: {
    memoryId?: string;
    results?: RecallResult[];
    strategy?: string;
    error?: string;
  };
}

const STORAGE_KEY = "brainsentry.console.entries";

export default function ConsolePage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [mode, setMode] = useState<Mode>("recall");
  const [input, setInput] = useState("");
  const [entries, setEntries] = useState<ConsoleEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [router, setRouter] = useState<RouterDecision | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const feedRef = useRef<HTMLDivElement>(null);

  // Load last 20 entries from localStorage
  useEffect(() => {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw) as ConsoleEntry[];
        setEntries(
          parsed.map((e) => ({ ...e, timestamp: new Date(e.timestamp) }))
        );
      }
    } catch {
      // ignore corrupted state
    }
  }, []);

  // Persist on change
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(entries.slice(-20)));
    } catch {
      // quota ignored
    }
  }, [entries]);

  // Auto-scroll feed
  useEffect(() => {
    if (feedRef.current) {
      feedRef.current.scrollTop = feedRef.current.scrollHeight;
    }
  }, [entries]);

  // Preview classification as user types (for recall mode)
  const classifyPreview = useCallback(async (query: string) => {
    if (mode !== "recall" || query.trim().length < 3) {
      setRouter(null);
      return;
    }
    try {
      const r = await api.classifyQuery(query);
      setRouter(r);
    } catch {
      setRouter(null);
    }
  }, [mode]);

  useEffect(() => {
    const t = setTimeout(() => classifyPreview(input), 350);
    return () => clearTimeout(t);
  }, [input, classifyPreview]);

  const handleSubmit = async () => {
    const text = input.trim();
    if (!text || loading) return;

    setLoading(true);
    const newEntry: ConsoleEntry = {
      id: `${Date.now()}`,
      mode,
      timestamp: new Date(),
      input: text,
      output: {},
    };

    try {
      if (mode === "remember") {
        const resp = await api.remember({ text });
        newEntry.output = { memoryId: resp.memoryId };
        toast({
          title: t("console.savedToast"),
          description: t("console.savedDesc", { id: resp.memoryId.slice(0, 8) }),
          variant: "success",
        });
      } else {
        const resp = await api.recall({ query: text, limit: 8 });
        newEntry.output = { results: resp.results, strategy: resp.strategy };
      }
      setEntries((prev) => [...prev, newEntry]);
      setInput("");
      setRouter(null);
      textareaRef.current?.focus();
    } catch (err: any) {
      newEntry.output = { error: err?.message || t("console.requestFailed") };
      setEntries((prev) => [...prev, newEntry]);
      toast({
        title: t("common.error"),
        description: err?.message || t("console.requestFailed"),
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      handleSubmit();
    }
  };

  const handleImprove = async () => {
    setLoading(true);
    try {
      const resp = await api.improve({ dryRun: false });
      toast({
        title: t("console.improveOk"),
        description: resp.message,
        variant: "success",
      });
    } catch (err: any) {
      toast({
        title: t("console.improveFail"),
        description: err?.message || t("console.requestFailed"),
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const clearHistory = () => {
    setEntries([]);
    localStorage.removeItem(STORAGE_KEY);
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Header */}
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <MessageSquare className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("console.title")}</h1>
                <p className="text-xs text-white/80">{t("console.subtitle")}</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                className="bg-white/20 border-white/30 text-white hover:bg-white/30"
                onClick={handleImprove}
                disabled={loading}
              >
                <Sparkles className="h-4 w-4 mr-2" />
                {t("console.improve")}
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="bg-white/20 border-white/30 text-white hover:bg-white/30"
                onClick={clearHistory}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="flex-1 container mx-auto px-4 py-6 flex flex-col max-w-5xl">
        {/* Feed */}
        <div
          ref={feedRef}
          className="flex-1 overflow-y-auto space-y-3 mb-4 pb-4"
          style={{ maxHeight: "calc(100vh - 280px)" }}
        >
          {entries.length === 0 ? (
            <EmptyState
              icon={Brain}
              title={t("console.emptyTitle")}
              description={t("console.emptyDesc")}
            />
          ) : (
            entries.map((e) => <EntryCard key={e.id} entry={e} />)
          )}
        </div>

        {/* Mode switch + Router preview */}
        <div className="mb-3 flex items-center justify-between flex-wrap gap-3">
          <div className="inline-flex rounded-lg bg-muted p-0.5">
            {(["recall", "remember"] as Mode[]).map((m) => (
              <button
                key={m}
                onClick={() => setMode(m)}
                className={`px-4 py-1.5 text-xs font-medium rounded-md transition-colors capitalize ${
                  mode === m
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {m === "recall" ? <Brain className="h-3.5 w-3.5 inline mr-1" /> : <Zap className="h-3.5 w-3.5 inline mr-1" />}
                {t(`console.mode.${m}`)}
              </button>
            ))}
          </div>

          {router && mode === "recall" && (
            <div className="flex items-center gap-2 text-xs">
              <Target className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-muted-foreground">{t("console.router")}</span>
              <span
                className="font-mono px-2 py-0.5 rounded border"
                style={{
                  color: strategyColor(router.strategy),
                  borderColor: strategyColor(router.strategy) + "40",
                }}
              >
                {router.strategy}
              </span>
              <span className="text-muted-foreground">
                {Math.round(router.confidence * 100)}%
              </span>
              {router.fallback && (
                <span className="text-[10px] px-1 py-0.5 rounded bg-muted text-muted-foreground">{t("console.fallback")}</span>
              )}
            </div>
          )}
        </div>

        {/* Input */}
        <Card>
          <CardContent className="p-3">
            <textarea
              ref={textareaRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={
                mode === "remember"
                  ? t("console.placeholder.remember")
                  : t("console.placeholder.recall")
              }
              className="w-full bg-transparent resize-none outline-none text-sm min-h-[60px] max-h-[200px]"
              disabled={loading}
              autoFocus
            />
            <div className="flex items-center justify-between mt-2 pt-2 border-t">
              <p className="text-[10px] text-muted-foreground">
                {input.length} {t("console.footer.chars")} · {mode === "remember" ? t("console.footer.remember") : t("console.footer.recall")}
              </p>
              <Button
                size="sm"
                onClick={handleSubmit}
                disabled={!input.trim() || loading}
                className="bg-gradient-to-r from-brain-primary to-brain-accent text-white"
              >
                {loading ? (
                  <Spinner size="sm" />
                ) : (
                  <>
                    <Send className="h-3.5 w-3.5 mr-2" />
                    {mode === "remember" ? t("console.button.remember") : t("console.button.recall")}
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>
      </main>
    </div>
  );
}

function EntryCard({ entry }: { entry: ConsoleEntry }) {
  const { t, i18n } = useTranslation();
  const isRemember = entry.mode === "remember";
  const color = isRemember ? "#f59e0b" : "#3b82f6";
  const Icon = isRemember ? Zap : Brain;

  return (
    <div className="flex gap-3">
      <div className="flex-shrink-0 mt-1">
        <div
          className="h-7 w-7 rounded-full flex items-center justify-center"
          style={{ backgroundColor: color + "20", color }}
        >
          <Icon className="h-3.5 w-3.5" />
        </div>
      </div>
      <div className="flex-1 min-w-0">
        {/* Input row */}
        <div className="flex items-center gap-2 mb-1">
          <span
            className="text-[10px] uppercase tracking-wider font-medium"
            style={{ color }}
          >
            {t(`console.mode.${entry.mode}`)}
          </span>
          <span className="text-[10px] text-muted-foreground">
            {entry.timestamp.toLocaleTimeString(i18n.language, { hour: "2-digit", minute: "2-digit", second: "2-digit" })}
          </span>
        </div>
        <Card className="border-l-[3px]" style={{ borderLeftColor: color }}>
          <CardContent className="p-3">
            <p className="text-sm whitespace-pre-wrap break-words">{entry.input}</p>
          </CardContent>
        </Card>

        {/* Output */}
        {entry.output.error && (
          <p className="mt-2 text-xs text-destructive">× {entry.output.error}</p>
        )}
        {entry.output.memoryId && (
          <p className="mt-2 text-xs text-muted-foreground">
            {t("console.savedAs")} <span className="font-mono text-foreground">{entry.output.memoryId.slice(0, 8)}...</span>
          </p>
        )}
        {entry.output.results && (
          <div className="mt-2 space-y-2">
            <p className="text-xs text-muted-foreground">
              {t("console.resultsInfo", { count: entry.output.results.length })}{" "}
              <span className="font-mono" style={{ color: strategyColor(entry.output.strategy || "HYBRID") }}>
                {entry.output.strategy}
              </span>
            </p>
            {entry.output.results.slice(0, 5).map((r) => (
              <RecallResultCard key={r.memoryId} result={r} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function RecallResultCard({ result }: { result: RecallResult }) {
  const content =
    result.summary || (result.content.length > 200 ? result.content.slice(0, 197) + "..." : result.content);

  return (
    <Card>
      <CardContent className="p-3">
        <p className="text-sm mb-2">{content}</p>
        <div className="flex items-center gap-3 flex-wrap">
          {result.category && (
            <span className="text-[10px] px-1.5 py-0.5 rounded border text-muted-foreground">
              {result.category}
            </span>
          )}
          <div className="flex items-center gap-1.5 flex-1 min-w-[120px]">
            <TrendingUp className="h-3 w-3 text-muted-foreground flex-shrink-0" />
            <StrengthBar value={result.relevance} max={1.0} size="sm" />
          </div>
          <span className="text-[10px] text-muted-foreground font-mono">
            fb: {Math.round(result.feedbackWeight * 100)}%
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

function strategyColor(strategy: string): string {
  const map: Record<string, string> = {
    LEXICAL: "#10b981",
    SEMANTIC: "#3b82f6",
    GRAPH: "#8b5cf6",
    TEMPORAL: "#f59e0b",
    ENTITY: "#ec4899",
    CODING: "#f97316",
    CYPHER: "#ef4444",
    HYBRID: "#6b7280",
  };
  return map[strategy] || "#6b7280";
}
