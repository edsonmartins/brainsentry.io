import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  Database, RefreshCw, Trash2, Sparkles, MessageSquare, Clock,
  Archive, ChevronRight,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import {
  api,
  type SessionInteraction,
} from "@/lib/api/client";

export default function SessionCachePage() {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const [sessions, setSessions] = useState<string[]>([]);
  const [selected, setSelected] = useState<string | null>(null);
  const [interactions, setInteractions] = useState<SessionInteraction[]>([]);
  const [loadingList, setLoadingList] = useState(true);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [busy, setBusy] = useState(false);

  const loadSessions = useCallback(async () => {
    setLoadingList(true);
    try {
      const data = await api.listSessionCaches();
      setSessions(data.sessions || []);
      if (!selected && data.sessions && data.sessions.length > 0) {
        setSelected(data.sessions[0]);
      }
    } catch (err: any) {
      toast({ title: t("sessionCache.loadFailed"), description: err?.message, variant: "error" });
    } finally {
      setLoadingList(false);
    }
  }, [selected, toast, t]);

  const loadDetail = useCallback(async (sessionId: string) => {
    setLoadingDetail(true);
    try {
      const data = await api.getSessionCache(sessionId, 50);
      setInteractions(data.interactions || []);
    } catch (err: any) {
      toast({ title: t("sessionCache.loadFailed"), description: err?.message, variant: "error" });
      setInteractions([]);
    } finally {
      setLoadingDetail(false);
    }
  }, [toast, t]);

  useEffect(() => {
    loadSessions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (selected) loadDetail(selected);
  }, [selected, loadDetail]);

  const clear = async () => {
    if (!selected) return;
    if (!confirm(t("sessionCache.clearConfirm", { id: selected }))) return;
    setBusy(true);
    try {
      await api.clearSessionCache(selected);
      toast({ title: t("sessionCache.cleared"), variant: "success" });
      setInteractions([]);
      await loadSessions();
    } catch (err: any) {
      toast({ title: t("sessionCache.clearFailed"), description: err?.message, variant: "error" });
    } finally {
      setBusy(false);
    }
  };

  const cognify = async (clearAfter: boolean) => {
    if (!selected) return;
    setBusy(true);
    try {
      const r = await api.cognifySessionCache(selected, clearAfter);
      toast({
        title: t("sessionCache.cognified"),
        description: t("sessionCache.cognifiedDesc", {
          created: r.memoriesCreated.length,
          total: r.interactions,
        }),
        variant: "success",
      });
      if (clearAfter) {
        setInteractions([]);
        await loadSessions();
      }
    } catch (err: any) {
      toast({ title: t("sessionCache.cognifyFailed"), description: err?.message, variant: "error" });
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Database className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("sessionCache.title")}</h1>
                <p className="text-xs text-white/80">{t("sessionCache.subtitle")}</p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="bg-white/20 border-white/30 text-white hover:bg-white/30"
              onClick={loadSessions}
              disabled={loadingList}
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 max-w-6xl">
        {loadingList ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : sessions.length === 0 ? (
          <EmptyState
            icon={Database}
            title={t("sessionCache.empty.title")}
            description={t("sessionCache.empty.desc")}
          />
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
            <Card className="lg:col-span-1">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm flex items-center gap-2">
                  <MessageSquare className="h-4 w-4" />
                  {t("sessionCache.sessions")}
                  <span className="text-[10px] text-muted-foreground">({sessions.length})</span>
                </CardTitle>
              </CardHeader>
              <CardContent className="p-2 space-y-1 max-h-[70vh] overflow-y-auto">
                {sessions.map((sid) => (
                  <button
                    key={sid}
                    onClick={() => setSelected(sid)}
                    className={`w-full text-left px-2 py-1.5 rounded text-xs font-mono truncate flex items-center gap-2 transition-colors ${
                      selected === sid
                        ? "bg-brain-primary/15 text-brain-primary"
                        : "hover:bg-muted"
                    }`}
                  >
                    <ChevronRight
                      className={`h-3 w-3 flex-shrink-0 ${selected === sid ? "opacity-100" : "opacity-30"}`}
                    />
                    {sid}
                  </button>
                ))}
              </CardContent>
            </Card>

            <div className="lg:col-span-2 space-y-3">
              {selected && (
                <>
                  <Card>
                    <CardContent className="p-3 flex items-center justify-between flex-wrap gap-2">
                      <div>
                        <p className="text-[10px] uppercase tracking-wider text-muted-foreground">{t("sessionCache.selected")}</p>
                        <p className="text-sm font-mono">{selected}</p>
                        <p className="text-[10px] text-muted-foreground">
                          {t("sessionCache.interactions", { count: interactions.length })}
                        </p>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => cognify(false)}
                          disabled={busy || interactions.length === 0}
                        >
                          <Sparkles className="h-3.5 w-3.5 mr-2" />
                          {t("sessionCache.cognify")}
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => cognify(true)}
                          disabled={busy || interactions.length === 0}
                        >
                          <Archive className="h-3.5 w-3.5 mr-2" />
                          {t("sessionCache.cognifyClear")}
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={clear}
                          disabled={busy}
                          className="text-destructive"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </CardContent>
                  </Card>

                  {loadingDetail ? (
                    <div className="flex justify-center py-8"><Spinner /></div>
                  ) : interactions.length === 0 ? (
                    <EmptyState
                      icon={MessageSquare}
                      title={t("sessionCache.noInteractions.title")}
                      description={t("sessionCache.noInteractions.desc")}
                    />
                  ) : (
                    <div className="space-y-2">
                      {interactions.map((it) => (
                        <InteractionCard key={it.id} it={it} locale={i18n.language} />
                      ))}
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

function InteractionCard({ it, locale }: { it: SessionInteraction; locale: string }) {
  const { t } = useTranslation();
  return (
    <Card>
      <CardContent className="p-3">
        <div className="flex items-center gap-2 text-[10px] text-muted-foreground mb-2">
          <Clock className="h-3 w-3" />
          {new Date(it.createdAt).toLocaleString(locale)}
          {it.memoryIds && it.memoryIds.length > 0 && (
            <span className="ml-2">{it.memoryIds.length} {t("sessionCache.memoryContext")}</span>
          )}
        </div>

        <div className="mb-2">
          <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-0.5">{t("sessionCache.query")}</p>
          <p className="text-sm">{it.query}</p>
        </div>

        <div>
          <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-0.5">{t("sessionCache.response")}</p>
          <p className="text-sm whitespace-pre-wrap break-words">{it.response}</p>
        </div>

        {it.metadata && Object.keys(it.metadata).length > 0 && (
          <div className="mt-2 pt-2 border-t flex flex-wrap gap-1">
            {Object.entries(it.metadata).map(([k, v]) => (
              <span key={k} className="text-[10px] px-1.5 py-0.5 rounded bg-muted">
                {k}: {v}
              </span>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
