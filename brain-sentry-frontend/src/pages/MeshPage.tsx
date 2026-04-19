import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  Share2, Plus, RefreshCw, CheckCircle, XCircle, Globe,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useToast } from "@/components/ui/toast";
import { api, type MeshPeer } from "@/lib/api/client";

const SCOPES = ["memories", "actions", "semantic", "procedural", "relations"];

export default function MeshPage() {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const [peers, setPeers] = useState<MeshPeer[]>([]);
  const [loading, setLoading] = useState(true);
  const [registerOpen, setRegisterOpen] = useState(false);
  const [syncing, setSyncing] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.listMeshPeers();
      setPeers(data || []);
    } catch (err: any) {
      toast({ title: t("mesh.loadFailed"), description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  }, [toast, t]);

  useEffect(() => { load(); }, [load]);

  const syncScope = async (scope: string) => {
    setSyncing(true);
    try {
      const results = await api.meshSync(scope, []);
      const ok = results.filter((r) => !r.error).length;
      toast({
        title: t("mesh.syncComplete", { scope }),
        description: t("mesh.syncDesc", { ok, total: results.length }),
        variant: "success",
      });
      await load();
    } catch (err: any) {
      toast({ title: t("mesh.syncFailed"), description: err?.message, variant: "error" });
    } finally {
      setSyncing(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Share2 className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("mesh.title")}</h1>
                <p className="text-xs text-white/80">{t("mesh.subtitle")}</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                className="bg-white/20 border-white/30 text-white hover:bg-white/30"
                onClick={load}
              >
                <RefreshCw className="h-4 w-4" />
              </Button>
              <Button
                size="sm"
                className="bg-white text-brain-primary hover:bg-white/90"
                onClick={() => setRegisterOpen(true)}
              >
                <Plus className="h-4 w-4 mr-1" /> {t("mesh.registerPeer")}
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 max-w-5xl">
        {loading ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : peers.length === 0 ? (
          <EmptyState
            icon={Globe}
            title={t("mesh.empty.title")}
            description={t("mesh.empty.desc")}
            action={{ label: t("mesh.registerPeer"), onClick: () => setRegisterOpen(true) }}
          />
        ) : (
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">{t("mesh.syncScopes")}</CardTitle>
              </CardHeader>
              <CardContent className="flex flex-wrap gap-2">
                {SCOPES.map((s) => (
                  <Button
                    key={s}
                    variant="outline"
                    size="sm"
                    disabled={syncing}
                    onClick={() => syncScope(s)}
                    className="text-xs"
                  >
                    {syncing ? <Spinner size="sm" /> : null}
                    {t("mesh.push", { scope: s })}
                  </Button>
                ))}
              </CardContent>
            </Card>

            <div className="space-y-2">
              {peers.map((p) => <PeerRow key={p.id} peer={p} locale={i18n.language} />)}
            </div>
          </div>
        )}
      </main>

      {registerOpen && (
        <RegisterPeerDialog
          onClose={() => setRegisterOpen(false)}
          onRegistered={() => { setRegisterOpen(false); load(); }}
        />
      )}
    </div>
  );
}

function PeerRow({ peer, locale }: { peer: MeshPeer; locale: string }) {
  const { t } = useTranslation();
  const isActive = peer.status === "active";
  const color = isActive ? "#10b981" : "#ef4444";

  return (
    <Card className="border-l-[3px]" style={{ borderLeftColor: color }}>
      <CardContent className="p-3">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <div className="flex items-center gap-2 mb-1">
              {isActive ? (
                <CheckCircle className="h-3.5 w-3.5 text-green-500" />
              ) : (
                <XCircle className="h-3.5 w-3.5 text-destructive" />
              )}
              <span className="text-sm font-mono">{peer.id}</span>
              <span
                className="text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded border"
                style={{ color, borderColor: color + "60" }}
              >
                {peer.status || "unknown"}
              </span>
            </div>
            <p className="text-xs text-muted-foreground truncate">{peer.url}</p>
            {peer.sharedScopes && peer.sharedScopes.length > 0 && (
              <div className="flex gap-1 mt-1">
                {peer.sharedScopes.map((s) => (
                  <span key={s} className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                    {s}
                  </span>
                ))}
              </div>
            )}
          </div>
          {peer.lastSyncAt && (
            <div className="text-[10px] text-muted-foreground">
              {t("mesh.last")} {new Date(peer.lastSyncAt).toLocaleString(locale)}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function RegisterPeerDialog({
  onClose, onRegistered,
}: {
  onClose: () => void;
  onRegistered: () => void;
}) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [id, setId] = useState("");
  const [url, setUrl] = useState("https://");
  const [scopes, setScopes] = useState<string[]>(["memories"]);
  const [busy, setBusy] = useState(false);

  const toggleScope = (s: string) => {
    setScopes((prev) => prev.includes(s) ? prev.filter(x => x !== s) : [...prev, s]);
  };

  const submit = async () => {
    if (!id.trim() || !url.trim()) return;
    setBusy(true);
    try {
      await api.registerMeshPeer({ id, url, sharedScopes: scopes });
      toast({ title: t("mesh.registered"), variant: "success" });
      onRegistered();
    } catch (err: any) {
      toast({ title: t("mesh.registerFailed"), description: err?.message, variant: "error" });
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog open onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md" onClose={onClose}>
        <DialogHeader>
          <DialogTitle>{t("mesh.registerPeer")}</DialogTitle>
        </DialogHeader>
        <div className="p-6 pt-0 space-y-3">
          <div>
            <label className="text-xs uppercase tracking-wider text-muted-foreground">{t("mesh.peerId")}</label>
            <input
              type="text"
              value={id}
              onChange={(e) => setId(e.target.value)}
              placeholder="eu-west-1"
              className="w-full text-sm bg-transparent border rounded px-2 py-1 mt-1 font-mono"
              autoFocus
            />
          </div>
          <div>
            <label className="text-xs uppercase tracking-wider text-muted-foreground">{t("mesh.url")}</label>
            <input
              type="text"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://peer.example.com"
              className="w-full text-sm bg-transparent border rounded px-2 py-1 mt-1 font-mono"
            />
            <p className="text-[10px] text-muted-foreground mt-1">
              {t("mesh.urlHint")}
            </p>
          </div>
          <div>
            <label className="text-xs uppercase tracking-wider text-muted-foreground">{t("mesh.sharedScopes")}</label>
            <div className="flex flex-wrap gap-1.5 mt-1">
              {SCOPES.map((s) => (
                <button
                  key={s}
                  type="button"
                  onClick={() => toggleScope(s)}
                  className={`text-xs px-2 py-0.5 rounded-full border transition-colors ${
                    scopes.includes(s)
                      ? "bg-foreground text-background border-foreground"
                      : "border-border text-muted-foreground hover:border-foreground/50"
                  }`}
                >
                  {s}
                </button>
              ))}
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button size="sm" variant="outline" onClick={onClose} disabled={busy}>{t("common.cancel")}</Button>
            <Button size="sm" onClick={submit} disabled={busy || !id.trim() || !url.trim()}>
              {busy ? <Spinner size="sm" /> : t("mesh.register")}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
