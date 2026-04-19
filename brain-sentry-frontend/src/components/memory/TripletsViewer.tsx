import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Sparkles, RefreshCw, ArrowRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { StrengthBar } from "@/components/ui/StrengthBar";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import { api, type Triplet } from "@/lib/api/client";

interface TripletsViewerProps {
  /** Either a memoryId (server will look up content) or raw content. */
  memoryId?: string;
  content?: string;
  /** Hide the header card; useful for embedding. */
  compact?: boolean;
}

export function TripletsViewer({ memoryId, content, compact }: TripletsViewerProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [triplets, setTriplets] = useState<Triplet[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasRun, setHasRun] = useState(false);

  const extract = useCallback(async () => {
    if (!content && !memoryId) return;
    setLoading(true);
    try {
      const resp = await api.extractTriplets(content || "", memoryId);
      setTriplets(resp.triplets || []);
      setHasRun(true);
    } catch (err: any) {
      toast({
        title: t("triplets.extractionFailed"),
        description: err?.message || t("common.error"),
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  }, [content, memoryId, toast, t]);

  // Auto-extract on mount: for memoryId (cached server-side) or content (one-off).
  useEffect(() => {
    if ((memoryId || content) && !hasRun) extract();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [memoryId, content]);

  const body = loading ? (
    <div className="py-8 flex justify-center"><Spinner /></div>
  ) : !hasRun ? (
    <EmptyState
      icon={Sparkles}
      title={t("triplets.emptyTitle")}
      description={t("triplets.emptyDesc")}
      action={{ label: t("triplets.extractNow"), onClick: extract }}
    />
  ) : triplets.length === 0 ? (
    <EmptyState
      icon={Sparkles}
      title={t("triplets.noneTitle")}
      description={t("triplets.noneDesc")}
    />
  ) : (
    <div className="space-y-2">
      {triplets.map((tr) => (
        <TripletRow key={tr.id} triplet={tr} />
      ))}
    </div>
  );

  if (compact) return body;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-sm flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-brain-accent" />
          {t("triplets.title")}
          {hasRun && triplets.length > 0 && (
            <span className="text-[10px] text-muted-foreground">({triplets.length})</span>
          )}
        </CardTitle>
        {hasRun && (
          <Button variant="ghost" size="sm" onClick={extract} disabled={loading}>
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
          </Button>
        )}
      </CardHeader>
      <CardContent>{body}</CardContent>
    </Card>
  );
}

function TripletRow({ triplet }: { triplet: Triplet }) {
  const { t } = useTranslation();
  return (
    <div className="p-2 rounded-md border hover:bg-muted/40 transition-colors">
      <div className="flex items-center gap-2 flex-wrap">
        <span className="text-sm font-medium text-blue-500">{triplet.subject}</span>
        <ArrowRight className="h-3 w-3 text-muted-foreground" />
        <span className="text-xs px-1.5 py-0.5 rounded bg-brain-accent/15 text-brain-accent font-mono">
          {triplet.predicate}
        </span>
        <ArrowRight className="h-3 w-3 text-muted-foreground" />
        <span className="text-sm font-medium text-green-500">{triplet.object}</span>
      </div>
      <div className="mt-1.5 flex items-center gap-2">
        <span className="text-[10px] text-muted-foreground w-14">{t("triplets.confidence")}</span>
        <StrengthBar value={triplet.confidence} max={1.0} size="sm" />
      </div>
    </div>
  );
}
