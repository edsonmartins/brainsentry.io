import { useState, useEffect, useCallback, useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
  CheckSquare, Plus, RefreshCw, Lock, Unlock, Check, X,
  PauseCircle, PlayCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { TypeChips } from "@/components/ui/TypeChips";
import { useToast } from "@/components/ui/toast";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { api, type Action, type ActionStatus } from "@/lib/api/client";

const STATUS_COLORS: Record<ActionStatus, string> = {
  pending: "#6b7280",
  in_progress: "#3b82f6",
  blocked: "#f59e0b",
  completed: "#10b981",
  cancelled: "#ef4444",
};

export default function ActionsPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [actions, setActions] = useState<Action[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [leaseAgent, setLeaseAgent] = useState("agent-web-ui");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.listActions(statusFilter || undefined);
      setActions(data || []);
    } catch (err: any) {
      toast({ title: t("actions.loadFailed"), description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  }, [statusFilter, toast, t]);

  useEffect(() => { load(); }, [load]);

  const countsByStatus = useMemo(() => {
    const c: Record<string, number> = {};
    for (const a of actions) c[a.status] = (c[a.status] || 0) + 1;
    return c;
  }, [actions]);

  const chips = useMemo(
    () =>
      (Object.keys(STATUS_COLORS) as ActionStatus[]).map((s) => ({
        label: s,
        count: countsByStatus[s] || 0,
        color: STATUS_COLORS[s],
      })),
    [countsByStatus]
  );

  const handleStatusChange = async (id: string, status: string) => {
    try {
      await api.updateActionStatus(id, status);
      await load();
    } catch (err: any) {
      toast({ title: t("actions.updateFailed"), description: err?.message, variant: "error" });
    }
  };

  const handleAcquireLease = async (id: string) => {
    try {
      const lease = await api.acquireLease(id, leaseAgent, 10);
      toast({
        title: t("actions.leaseAcquired"),
        description: t("actions.leaseAcquiredDesc", { time: new Date(lease.expiresAt).toLocaleTimeString() }),
        variant: "success",
      });
      await load();
    } catch (err: any) {
      toast({ title: t("actions.leaseFailed"), description: err?.message, variant: "error" });
    }
  };

  const handleReleaseLease = async (id: string, completed: boolean) => {
    try {
      await api.releaseLease(id, leaseAgent, completed);
      toast({ title: t("actions.leaseReleased"), variant: "success" });
      await load();
    } catch (err: any) {
      toast({ title: t("actions.releaseFailed"), description: err?.message, variant: "error" });
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <CheckSquare className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("actions.title")}</h1>
                <p className="text-xs text-white/80">{t("actions.subtitle")}</p>
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
                onClick={() => setCreateOpen(true)}
              >
                <Plus className="h-4 w-4 mr-1" /> {t("actions.new")}
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 max-w-5xl">
        <Card className="mb-4">
          <CardContent className="p-3 flex items-center justify-between flex-wrap gap-3">
            <TypeChips items={chips} selected={statusFilter} onSelect={setStatusFilter} />
            <div className="flex items-center gap-2 text-xs">
              <span className="text-muted-foreground">{t("actions.agent")}</span>
              <input
                type="text"
                value={leaseAgent}
                onChange={(e) => setLeaseAgent(e.target.value)}
                className="text-xs bg-transparent border rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-brain-primary font-mono"
                placeholder={t("actions.agentPlaceholder")}
              />
            </div>
          </CardContent>
        </Card>

        {loading ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : actions.length === 0 ? (
          <EmptyState
            icon={CheckSquare}
            title={t("actions.empty.title")}
            description={t("actions.empty.desc")}
            action={{ label: t("actions.createNew"), onClick: () => setCreateOpen(true) }}
          />
        ) : (
          <div className="space-y-2">
            {actions.map((a) => (
              <ActionRow
                key={a.id}
                action={a}
                leaseAgent={leaseAgent}
                onStatusChange={handleStatusChange}
                onAcquireLease={handleAcquireLease}
                onReleaseLease={handleReleaseLease}
              />
            ))}
          </div>
        )}
      </main>

      {createOpen && (
        <CreateActionDialog
          onClose={() => setCreateOpen(false)}
          onCreated={() => { setCreateOpen(false); load(); }}
          defaultCreator={leaseAgent}
        />
      )}
    </div>
  );
}

function ActionRow({
  action, leaseAgent, onStatusChange, onAcquireLease, onReleaseLease,
}: {
  action: Action;
  leaseAgent: string;
  onStatusChange: (id: string, status: string) => void;
  onAcquireLease: (id: string) => void;
  onReleaseLease: (id: string, completed: boolean) => void;
}) {
  const { t } = useTranslation();
  const color = STATUS_COLORS[action.status];
  const held = !!action.assignedTo;
  const isHolder = action.assignedTo === leaseAgent;

  return (
    <Card className="border-l-[3px]" style={{ borderLeftColor: color }}>
      <CardContent className="p-3">
        <div className="flex items-start gap-3">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap mb-1">
              <span className="text-sm font-medium truncate">{action.title}</span>
              <span
                className="text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded border"
                style={{ color, borderColor: color + "60" }}
              >
                {t(`actions.statusLabels.${action.status}`, action.status)}
              </span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted">
                P{action.priority}
              </span>
              {action.assignedTo && (
                <span
                  className={`text-[10px] px-1.5 py-0.5 rounded border flex items-center gap-1 ${
                    isHolder ? "border-blue-500/60 text-blue-500" : "border-orange-500/60 text-orange-500"
                  }`}
                >
                  <Lock className="h-2.5 w-2.5" />
                  {action.assignedTo}
                </span>
              )}
              {action.dependsOn && action.dependsOn.length > 0 && (
                <span className="text-[10px] text-muted-foreground">
                  {t("actions.dependsOn", { count: action.dependsOn.length })}
                </span>
              )}
            </div>
            {action.description && (
              <p className="text-xs text-muted-foreground">{action.description}</p>
            )}
            {action.tags && action.tags.length > 0 && (
              <div className="flex gap-1 mt-1">
                {action.tags.map((tag) => (
                  <span key={tag} className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                    {tag}
                  </span>
                ))}
              </div>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            {!held ? (
              <Button size="sm" variant="outline" onClick={() => onAcquireLease(action.id)}>
                <Lock className="h-3 w-3 mr-1" />
                {t("actions.claim")}
              </Button>
            ) : isHolder ? (
              <div className="flex gap-1">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => onReleaseLease(action.id, true)}
                  className="text-green-600"
                >
                  <Check className="h-3 w-3" />
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => onReleaseLease(action.id, false)}
                >
                  <Unlock className="h-3 w-3" />
                </Button>
              </div>
            ) : null}

            {action.status === "pending" && (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => onStatusChange(action.id, "blocked")}
                className="text-xs"
              >
                <PauseCircle className="h-3 w-3 mr-1" />
                {t("actions.block")}
              </Button>
            )}
            {action.status === "blocked" && (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => onStatusChange(action.id, "pending")}
                className="text-xs"
              >
                <PlayCircle className="h-3 w-3 mr-1" />
                {t("actions.resume")}
              </Button>
            )}
            {action.status !== "cancelled" && action.status !== "completed" && (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => onStatusChange(action.id, "cancelled")}
                className="text-xs text-destructive"
              >
                <X className="h-3 w-3 mr-1" />
                {t("actions.cancel")}
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function CreateActionDialog({
  onClose, onCreated, defaultCreator,
}: {
  onClose: () => void;
  onCreated: () => void;
  defaultCreator: string;
}) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState(5);
  const [tags, setTags] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    if (!title.trim()) return;
    setBusy(true);
    try {
      await api.createAction({
        title,
        description,
        createdBy: defaultCreator,
        priority,
        tags: tags.split(",").map((x) => x.trim()).filter(Boolean),
      });
      toast({ title: t("actions.actionCreated"), variant: "success" });
      onCreated();
    } catch (err: any) {
      toast({ title: t("actions.createFailed"), description: err?.message, variant: "error" });
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog open onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md" onClose={onClose}>
        <DialogHeader>
          <DialogTitle>{t("actions.createNew")}</DialogTitle>
        </DialogHeader>
        <div className="p-6 pt-0 space-y-3">
          <div>
            <label className="text-xs uppercase tracking-wider text-muted-foreground">{t("actions.title_field")}</label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full text-sm bg-transparent border rounded px-2 py-1 mt-1"
              autoFocus
            />
          </div>
          <div>
            <label className="text-xs uppercase tracking-wider text-muted-foreground">{t("actions.description")}</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full text-sm bg-transparent border rounded px-2 py-1 mt-1 min-h-[80px]"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs uppercase tracking-wider text-muted-foreground">{t("actions.priority")}</label>
              <input
                type="number"
                min={1}
                max={10}
                value={priority}
                onChange={(e) => setPriority(Number(e.target.value))}
                className="w-full text-sm bg-transparent border rounded px-2 py-1 mt-1"
              />
            </div>
            <div>
              <label className="text-xs uppercase tracking-wider text-muted-foreground">{t("actions.tags")}</label>
              <input
                type="text"
                value={tags}
                onChange={(e) => setTags(e.target.value)}
                placeholder={t("actions.tagsPlaceholder")}
                className="w-full text-sm bg-transparent border rounded px-2 py-1 mt-1"
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button size="sm" variant="outline" onClick={onClose} disabled={busy}>
              {t("common.cancel")}
            </Button>
            <Button size="sm" onClick={submit} disabled={busy || !title.trim()}>
              {busy ? <Spinner size="sm" /> : t("common.create")}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
