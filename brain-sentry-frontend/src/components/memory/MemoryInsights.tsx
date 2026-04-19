import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Tag, ThumbsUp, ThumbsDown, X, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { StrengthBar } from "@/components/ui/StrengthBar";
import { useToast } from "@/components/ui/toast";
import { api, type FeedbackWeightResponse } from "@/lib/api/client";

interface MemoryInsightsProps {
  memoryId: string;
}

/**
 * Side panel showing NodeSets (multi-set membership) and Feedback Weight
 * for a single memory. Read+write for sets; read-only for feedback.
 */
export function MemoryInsights({ memoryId }: MemoryInsightsProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [sets, setSets] = useState<string[]>([]);
  const [newSet, setNewSet] = useState("");
  const [feedback, setFeedback] = useState<FeedbackWeightResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [s, f] = await Promise.all([
        api.getMemorySets(memoryId).catch(() => ({ memoryId, sets: [] })),
        api.getFeedbackWeight(memoryId).catch(() => null),
      ]);
      setSets(s.sets || []);
      setFeedback(f);
    } finally {
      setLoading(false);
    }
  }, [memoryId]);

  useEffect(() => {
    load();
  }, [load]);

  const addSet = async () => {
    const name = newSet.trim();
    if (!name) return;
    setBusy(true);
    try {
      const resp = await api.addMemorySets(memoryId, [name]);
      setSets(resp.sets || []);
      setNewSet("");
    } catch (err: any) {
      toast({ title: t("common.error"), description: err?.message, variant: "error" });
    } finally {
      setBusy(false);
    }
  };

  const removeSet = async (name: string) => {
    setBusy(true);
    try {
      const resp = await api.removeMemorySets(memoryId, [name]);
      setSets(resp.sets || []);
    } catch (err: any) {
      toast({ title: t("common.error"), description: err?.message, variant: "error" });
    } finally {
      setBusy(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center py-4">
        <Spinner size="sm" />
      </div>
    );
  }

  return (
    <div className="space-y-4 text-sm">
      {/* Sets */}
      <section>
        <div className="flex items-center gap-2 mb-2">
          <Tag className="h-3.5 w-3.5 text-muted-foreground" />
          <p className="text-[10px] uppercase tracking-wider text-muted-foreground">{t("memory.nodeSets")}</p>
        </div>
        <div className="flex flex-wrap gap-1.5 mb-2">
          {sets.length === 0 ? (
            <p className="text-xs text-muted-foreground italic">{t("memory.noSets")}</p>
          ) : (
            sets.map((s) => (
              <span
                key={s}
                className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full border bg-muted/50"
              >
                {s}
                <button
                  onClick={() => removeSet(s)}
                  disabled={busy}
                  className="hover:text-destructive"
                  title={t("common.remove")}
                >
                  <X className="h-2.5 w-2.5" />
                </button>
              </span>
            ))
          )}
        </div>
        <div className="flex gap-2">
          <input
            type="text"
            value={newSet}
            onChange={(e) => setNewSet(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && addSet()}
            placeholder={t("memory.addSet")}
            className="flex-1 text-xs bg-transparent border rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-brain-primary"
            disabled={busy}
          />
          <Button size="sm" variant="outline" onClick={addSet} disabled={busy || !newSet.trim()}>
            <Plus className="h-3 w-3" />
          </Button>
        </div>
      </section>

      {/* Feedback */}
      {feedback && (
        <section>
          <div className="flex items-center gap-2 mb-2">
            <ThumbsUp className="h-3.5 w-3.5 text-muted-foreground" />
            <p className="text-[10px] uppercase tracking-wider text-muted-foreground">{t("memory.feedbackWeight")}</p>
          </div>

          <div className="space-y-2 p-2 rounded border bg-muted/20">
            <div className="flex items-center gap-3">
              <div className="flex-1 min-w-[80px]">
                <StrengthBar value={feedback.feedbackWeight} max={1.0} size="sm" />
              </div>
              <span className="text-xs font-mono">{Math.round(feedback.feedbackWeight * 100)}%</span>
            </div>
            <div className="flex items-center gap-4 text-[10px] text-muted-foreground">
              <span className="flex items-center gap-1">
                <ThumbsUp className="h-3 w-3 text-green-500" />
                {feedback.helpfulCount} {t("memory.helpful")}
              </span>
              <span className="flex items-center gap-1">
                <ThumbsDown className="h-3 w-3 text-red-500" />
                {feedback.notHelpfulCount} {t("memory.notHelpful")}
              </span>
              <span className="ml-auto">α={feedback.alpha}</span>
            </div>
          </div>
        </section>
      )}
    </div>
  );
}
