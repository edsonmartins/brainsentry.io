import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Plus, Search, Loader2, FileText, Download, Upload,
  ChevronLeft, ChevronRight, Flag, History, RotateCcw,
  AlertTriangle, CheckCircle, XCircle, Clock,
} from "lucide-react";
import { MemoryCard, MemoryDialog } from "@/components/memory";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { api, type Memory, getErrorMessage } from "@/lib/api";
import { useToast } from "@/components/ui/toast";

interface MemoryVersion {
  version: number;
  content: string;
  summary: string;
  category: string;
  importance: string;
  tags: string[];
  changedAt: string;
  changedBy?: string;
}

interface ConflictResult {
  memoryId: string;
  conflicts: Array<{
    conflictingMemoryId: string;
    conflictingSummary: string;
    reason: string;
    severity: string;
  }>;
}

export default function MemoryAdminPage() {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const dateLocale = i18n.language === "en" ? "en-US" : "pt-BR";
  const [memories, setMemories] = useState<Memory[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedMemory, setSelectedMemory] = useState<Memory | undefined>(undefined);

  // Pagination
  const [page, setPage] = useState(0);
  const [pageSize] = useState(20);
  const [totalPages, setTotalPages] = useState(0);
  const [totalItems, setTotalItems] = useState(0);

  // Versions dialog
  const [versionsDialogOpen, setVersionsDialogOpen] = useState(false);
  const [versionsMemoryId, setVersionsMemoryId] = useState<string | null>(null);
  const [versions, setVersions] = useState<MemoryVersion[]>([]);
  const [versionsLoading, setVersionsLoading] = useState(false);

  // Flag dialog
  const [flagDialogOpen, setFlagDialogOpen] = useState(false);
  const [flagMemoryId, setFlagMemoryId] = useState<string | null>(null);
  const [flagReason, setFlagReason] = useState("");
  const [flagLoading, setFlagLoading] = useState(false);

  // Conflicts dialog
  const [conflictsDialogOpen, setConflictsDialogOpen] = useState(false);
  const [conflictsMemoryId, setConflictsMemoryId] = useState<string | null>(null);
  const [conflicts, setConflicts] = useState<ConflictResult | null>(null);
  const [conflictsLoading, setConflictsLoading] = useState(false);

  // Buscar memórias do backend (server-side pagination)
  const fetchMemories = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      let response;
      if (searchQuery.length >= 2) {
        const searchResults = await api.searchMemories(searchQuery, pageSize);
        response = { memories: searchResults, total: searchResults.length, totalPages: 1 };
      } else {
        response = await api.getMemories(page, pageSize);
      }
      setMemories(response.memories || []);
      setTotalItems(response.total || 0);
      setTotalPages(response.totalPages || Math.ceil((response.total || 0) / pageSize));
    } catch (err) {
      setError(getErrorMessage(err));
      console.error("Error fetching memories:", err);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, searchQuery]);

  useEffect(() => {
    fetchMemories();
  }, [fetchMemories]);

  useEffect(() => {
    setPage(0);
  }, [searchQuery]);

  const handleExport = async () => {
    try {
      const resp = await api.axiosInstance.get("/v1/batch/export", { responseType: "blob" });
      const url = window.URL.createObjectURL(new Blob([resp.data]));
      const a = document.createElement("a");
      a.href = url;
      a.download = `memories-export-${new Date().toISOString()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      toast({ title: t("memory.exportOk"), variant: "success" });
    } catch (err) {
      toast({ title: t("memory.exportError"), description: getErrorMessage(err), variant: "error" });
    }
  };

  const handleImport = () => {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = ".json";
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;
      try {
        const text = await file.text();
        const data = JSON.parse(text);
        await api.axiosInstance.post("/v1/batch/import", data);
        toast({ title: t("memory.importOk"), variant: "success" });
        fetchMemories();
      } catch (err) {
        toast({ title: t("memory.importError"), description: getErrorMessage(err), variant: "error" });
      }
    };
    input.click();
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("memory.deleteConfirmLong"))) return;
    try {
      await api.deleteMemory(id);
      setMemories(memories.filter(m => m.id !== id));
      toast({ title: t("memory.deleted"), variant: "success" });
    } catch (err) {
      toast({ title: t("memory.deleteError"), description: getErrorMessage(err), variant: "error" });
    }
  };

  const handleView = (id: string) => {
    const memory = memories.find(m => m.id === id);
    if (memory) {
      setSelectedMemory(memory);
      setDialogOpen(true);
    }
  };

  const handleEdit = (id: string) => {
    const memory = memories.find(m => m.id === id);
    if (memory) {
      setSelectedMemory(memory);
      setDialogOpen(true);
    }
  };

  const handleCreate = () => {
    setSelectedMemory(undefined);
    setDialogOpen(true);
  };

  const handleDialogClose = () => {
    setDialogOpen(false);
    setSelectedMemory(undefined);
  };

  // --- Versions ---
  const handleViewVersions = async (id: string) => {
    setVersionsMemoryId(id);
    setVersionsDialogOpen(true);
    setVersionsLoading(true);
    try {
      const data = await api.getMemoryVersions(id);
      setVersions(Array.isArray(data) ? data : data?.versions || []);
    } catch (err) {
      toast({ title: t("memory.versionsLoadError"), description: getErrorMessage(err), variant: "error" });
      setVersions([]);
    } finally {
      setVersionsLoading(false);
    }
  };

  const handleRollback = async (version: number) => {
    if (!versionsMemoryId) return;
    if (!confirm(t("memory.rollbackConfirm", { n: version }))) return;
    try {
      await api.rollbackMemory(versionsMemoryId, version);
      toast({ title: t("memory.rollbackOk"), description: t("memory.rollbackDesc", { n: version }), variant: "success" });
      setVersionsDialogOpen(false);
      fetchMemories();
    } catch (err) {
      toast({ title: t("memory.rollbackError"), description: getErrorMessage(err), variant: "error" });
    }
  };

  // --- Flag ---
  const handleOpenFlag = (id: string) => {
    setFlagMemoryId(id);
    setFlagReason("");
    setFlagDialogOpen(true);
  };

  const handleSubmitFlag = async () => {
    if (!flagMemoryId || !flagReason.trim()) return;
    setFlagLoading(true);
    try {
      await api.flagMemory(flagMemoryId, flagReason);
      toast({ title: t("memory.flagged"), description: t("memory.flaggedDesc"), variant: "success" });
      setFlagDialogOpen(false);
    } catch (err) {
      toast({ title: t("memory.flagError"), description: getErrorMessage(err), variant: "error" });
    } finally {
      setFlagLoading(false);
    }
  };

  const handleReviewMemory = async (id: string, action: "approve" | "reject") => {
    try {
      await api.reviewCorrection(id, action);
      toast({ title: action === "approve" ? t("memory.reviewApproved") : t("memory.reviewRejected"), variant: "success" });
      fetchMemories();
    } catch (err) {
      toast({ title: t("memory.reviewError"), description: getErrorMessage(err), variant: "error" });
    }
  };

  // --- Conflicts ---
  const handleDetectConflicts = async (id: string) => {
    setConflictsMemoryId(id);
    setConflictsDialogOpen(true);
    setConflictsLoading(true);
    try {
      const data = await api.detectConflicts(id);
      setConflicts(data);
    } catch (err) {
      toast({ title: t("memory.conflictsError"), description: getErrorMessage(err), variant: "error" });
      setConflicts(null);
    } finally {
      setConflictsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <FileText className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("memory.title")}</h1>
                <p className="text-xs text-white/80">
                  {t("memory.subtitle")}
                </p>
              </div>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Actions Bar */}
        <div className="flex items-center justify-between gap-4 mb-6">
          <div className="flex-1 max-w-sm">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("memory.searchPlaceholder")}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 text-sm bg-background border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-brain-primary/50 focus:border-brain-primary transition-colors"
              />
            </div>
          </div>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={handleImport}>
              <Upload className="h-4 w-4 mr-2" />
              {t("common.import")}
            </Button>
            <Button size="sm" variant="outline" onClick={handleExport}>
              <Download className="h-4 w-4 mr-2" />
              {t("common.export")}
            </Button>
            <Button size="sm" className="bg-white text-brain-primary hover:bg-white/90" onClick={handleCreate}>
              <Plus className="h-4 w-4 mr-2" />
              {t("memory.new")}
            </Button>
          </div>
        </div>

      {/* Loading State */}
      {loading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          <span className="ml-2 text-muted-foreground">{t("common.loading")}</span>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-destructive/10 text-destructive p-4 rounded-md">
          <p className="font-medium">{t("memory.loadError")}</p>
          <p className="text-sm">{error}</p>
          <Button
            size="sm"
            variant="outline"
            className="mt-2"
            onClick={fetchMemories}
          >
            {t("common.try_again")}
          </Button>
        </div>
      )}

      {/* Memories Grid */}
      {!loading && !error && (
        <>
          <div className="text-sm text-muted-foreground mb-3">
            {t("memory.count", { count: totalItems })}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {memories.map((memory) => (
              <div key={memory.id} className="relative group">
                <MemoryCard
                  memory={memory}
                  onView={handleView}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                />
                {/* Extra action buttons */}
                <div className="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={() => handleViewVersions(memory.id)}
                    className="p-1.5 bg-background/90 rounded-md border shadow-sm hover:bg-accent"
                    title={t("memory.history")}
                  >
                    <History className="h-3.5 w-3.5 text-blue-600" />
                  </button>
                  <button
                    onClick={() => handleOpenFlag(memory.id)}
                    className="p-1.5 bg-background/90 rounded-md border shadow-sm hover:bg-accent"
                    title={t("memory.flag")}
                  >
                    <Flag className="h-3.5 w-3.5 text-orange-500" />
                  </button>
                  <button
                    onClick={() => handleDetectConflicts(memory.id)}
                    className="p-1.5 bg-background/90 rounded-md border shadow-sm hover:bg-accent"
                    title={t("memory.detectConflicts")}
                  >
                    <AlertTriangle className="h-3.5 w-3.5 text-yellow-600" />
                  </button>
                </div>
              </div>
            ))}
          </div>

          {memories.length === 0 && !searchQuery && (
            <div className="text-center py-12">
              <p className="text-muted-foreground">{t("memory.noneRegistered")}</p>
              <p className="text-sm text-muted-foreground mt-1">
                {t("memory.createFirst")}
              </p>
            </div>
          )}

          {memories.length === 0 && searchQuery && (
            <div className="text-center py-12">
              <p className="text-muted-foreground">{t("memory.noResultsFor", { query: searchQuery })}</p>
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-4 mt-6">
              <Button
                size="sm"
                variant="outline"
                disabled={page === 0}
                onClick={() => setPage((p) => Math.max(0, p - 1))}
              >
                <ChevronLeft className="h-4 w-4" />
                {t("common.previous")}
              </Button>
              <span className="text-sm text-muted-foreground">
                {t("memory.page", { current: page + 1, total: totalPages })}
              </span>
              <Button
                size="sm"
                variant="outline"
                disabled={page >= totalPages - 1}
                onClick={() => setPage((p) => p + 1)}
              >
                {t("common.next")}
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          )}
        </>
      )}

      {/* Memory Dialog */}
      <MemoryDialog
        open={dialogOpen}
        onOpenChange={handleDialogClose}
        memory={selectedMemory}
        onSuccess={fetchMemories}
      />

      {/* Versions Dialog */}
      <Dialog open={versionsDialogOpen} onOpenChange={setVersionsDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <History className="h-5 w-5 text-blue-600" />
              {t("memory.versionsTitle")}
            </DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto py-4">
            {versionsLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                <span className="ml-2 text-muted-foreground">{t("memory.versionsLoading")}</span>
              </div>
            ) : versions.length === 0 ? (
              <p className="text-center text-muted-foreground py-8">{t("memory.versionsNone")}</p>
            ) : (
              <div className="space-y-4">
                {versions.map((v, idx) => (
                  <Card key={v.version || idx} className="relative">
                    <CardHeader className="pb-2">
                      <div className="flex items-center justify-between">
                        <CardTitle className="text-sm flex items-center gap-2">
                          <Clock className="h-4 w-4 text-muted-foreground" />
                          {t("memory.versionLabel", { n: v.version })}
                          {idx === 0 && (
                            <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">{t("memory.current")}</span>
                          )}
                        </CardTitle>
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-muted-foreground">
                            {v.changedAt ? new Date(v.changedAt).toLocaleString(dateLocale) : "-"}
                          </span>
                          {idx > 0 && (
                            <Button size="sm" variant="outline" onClick={() => handleRollback(v.version)}>
                              <RotateCcw className="h-3 w-3 mr-1" />
                              {t("memory.rollback")}
                            </Button>
                          )}
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="pt-0">
                      <p className="text-sm font-medium">{v.summary}</p>
                      <p className="text-xs text-muted-foreground mt-1 line-clamp-3">{v.content}</p>
                      <div className="flex gap-2 mt-2">
                        <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded">{v.category}</span>
                        <span className="text-xs bg-orange-100 text-orange-700 px-2 py-0.5 rounded">{v.importance}</span>
                        {v.tags?.map(t => (
                          <span key={t} className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded">{t}</span>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* Flag Dialog */}
      <Dialog open={flagDialogOpen} onOpenChange={setFlagDialogOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Flag className="h-5 w-5 text-orange-500" />
              {t("memory.flagTitle")}
            </DialogTitle>
          </DialogHeader>
          <div className="py-4 space-y-4">
            <p className="text-sm text-muted-foreground">
              {t("memory.flagDesc")}
            </p>
            <textarea
              value={flagReason}
              onChange={(e) => setFlagReason(e.target.value)}
              placeholder={t("memory.flagPlaceholder")}
              className="w-full h-24 rounded-md border border-input bg-background px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-brain-primary/50"
            />
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <AlertTriangle className="h-3.5 w-3.5" />
              {t("memory.flagHint")}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setFlagDialogOpen(false)}>{t("common.cancel")}</Button>
            <Button
              onClick={handleSubmitFlag}
              disabled={!flagReason.trim() || flagLoading}
              className="bg-orange-500 hover:bg-orange-600 text-white"
            >
              {flagLoading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Flag className="h-4 w-4 mr-2" />}
              {t("memory.flag")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Conflicts Dialog */}
      <Dialog open={conflictsDialogOpen} onOpenChange={setConflictsDialogOpen}>
        <DialogContent className="max-w-lg max-h-[70vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-yellow-600" />
              {t("memory.conflictsTitle")}
            </DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto py-4">
            {conflictsLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                <span className="ml-2 text-muted-foreground">{t("memory.conflictsAnalyzing")}</span>
              </div>
            ) : !conflicts || !conflicts.conflicts || conflicts.conflicts.length === 0 ? (
              <div className="text-center py-8">
                <CheckCircle className="h-12 w-12 mx-auto text-green-500 mb-3" />
                <p className="font-medium text-green-700">{t("memory.noConflicts")}</p>
                <p className="text-sm text-muted-foreground mt-1">{t("memory.noConflictsDesc")}</p>
              </div>
            ) : (
              <div className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  {t("memory.conflictsCount", { count: conflicts.conflicts.length })}
                </p>
                {conflicts.conflicts.map((c, idx) => (
                  <Card key={idx} className="border-yellow-200 bg-yellow-50/50">
                    <CardContent className="p-4">
                      <div className="flex items-start gap-3">
                        <XCircle className="h-5 w-5 text-yellow-600 shrink-0 mt-0.5" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">{c.conflictingSummary || c.conflictingMemoryId}</p>
                          <p className="text-xs text-muted-foreground mt-1">{c.reason}</p>
                          <div className="flex items-center gap-2 mt-2">
                            <span className={`text-xs px-2 py-0.5 rounded ${
                              c.severity === "HIGH" ? "bg-red-100 text-red-700" :
                              c.severity === "MEDIUM" ? "bg-yellow-100 text-yellow-700" :
                              "bg-gray-100 text-gray-600"
                            }`}>
                              {c.severity || "MEDIUM"}
                            </span>
                            <Button
                              size="sm"
                              variant="outline"
                              className="h-6 text-xs"
                              onClick={() => {
                                handleReviewMemory(conflictsMemoryId!, "approve");
                                setConflictsDialogOpen(false);
                              }}
                            >
                              <CheckCircle className="h-3 w-3 mr-1" />
                              {t("memory.approve")}
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              className="h-6 text-xs"
                              onClick={() => {
                                handleReviewMemory(conflictsMemoryId!, "reject");
                                setConflictsDialogOpen(false);
                              }}
                            >
                              <XCircle className="h-3 w-3 mr-1" />
                              {t("memory.reject")}
                            </Button>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
      </main>
    </div>
  );
}
