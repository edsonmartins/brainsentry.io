import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Plug, RefreshCw, Play, CheckCircle, AlertCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useToast } from "@/components/ui/toast";
import { api } from "@/lib/api";

interface SyncResult {
  connector: string;
  documentsFound: number;
  chunksCreated: number;
  tasksSubmitted: number;
  error?: string;
}

export default function ConnectorsPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [connectors, setConnectors] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState<Record<string, boolean>>({});
  const [results, setResults] = useState<Record<string, SyncResult>>({});

  const fetchConnectors = async () => {
    setLoading(true);
    try {
      const resp = await api.axiosInstance.get<{ connectors: string[] }>("/v1/connectors");
      setConnectors(resp.data.connectors || []);
    } catch (err: any) {
      toast({ title: t("connectors.error"), description: err.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConnectors();
  }, []);

  const handleSync = async (name: string) => {
    setSyncing((prev) => ({ ...prev, [name]: true }));
    try {
      const resp = await api.axiosInstance.post<SyncResult>(`/v1/connectors/${name}/sync`);
      setResults((prev) => ({ ...prev, [name]: resp.data }));
      toast({
        title: t("connectors.syncOk"),
        description: t("connectors.syncOkDesc", { name, count: resp.data.documentsFound }),
        variant: "success",
      });
    } catch (err: any) {
      toast({ title: t("connectors.syncError"), description: err.message, variant: "error" });
    } finally {
      setSyncing((prev) => ({ ...prev, [name]: false }));
    }
  };

  const handleSyncAll = async () => {
    const allNames = connectors.reduce((acc, name) => ({ ...acc, [name]: true }), {} as Record<string, boolean>);
    setSyncing(allNames);
    try {
      const resp = await api.axiosInstance.post<Record<string, SyncResult>>("/v1/connectors/sync-all");
      setResults(resp.data);
      toast({ title: t("connectors.syncAllOk"), description: t("connectors.syncAllOkDesc"), variant: "success" });
    } catch (err: any) {
      toast({ title: t("connectors.error"), description: err.message, variant: "error" });
    } finally {
      setSyncing({});
    }
  };

  const connectorIcons: Record<string, string> = {
    github: "GitHub",
    notion: "Notion",
    drive: "Google Drive",
    webcrawler: "Web Crawler",
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Plug className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("connectors.title")}</h1>
                <p className="text-xs text-white/80">
                  {t("connectors.subtitle")}
                </p>
              </div>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" className="bg-white/20 text-white border-white/30 hover:bg-white/30" onClick={fetchConnectors}>
                <RefreshCw className="h-4 w-4" />
              </Button>
              {connectors.length > 0 && (
                <Button variant="outline" size="sm" className="bg-white text-brain-primary hover:bg-white/90 border-0" onClick={handleSyncAll}>
                  <Play className="h-4 w-4 mr-2" />
                  {t("connectors.syncAll")}
                </Button>
              )}
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {loading && (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        )}

        {!loading && connectors.length === 0 && (
          <Card>
            <CardContent className="p-12 text-center">
              <Plug className="h-16 w-16 mx-auto mb-4 text-muted-foreground opacity-50" />
              <h3 className="text-lg font-semibold mb-2">{t("connectors.empty")}</h3>
              <p className="text-muted-foreground">
                {t("connectors.emptyDesc")}
              </p>
            </CardContent>
          </Card>
        )}

        {!loading && connectors.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {connectors.map((name) => {
              const result = results[name];
              const isSyncing = syncing[name];

              return (
                <Card key={name}>
                  <CardHeader>
                    <CardTitle className="flex items-center justify-between">
                      <span>{connectorIcons[name] || name}</span>
                      {result && !result.error && <CheckCircle className="h-4 w-4 text-green-500" />}
                      {result?.error && <AlertCircle className="h-4 w-4 text-red-500" />}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {result && (
                      <div className="text-sm space-y-1">
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{t("connectors.documents")}</span>
                          <span className="font-medium">{result.documentsFound}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{t("connectors.chunks")}</span>
                          <span className="font-medium">{result.chunksCreated}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{t("connectors.tasks")}</span>
                          <span className="font-medium">{result.tasksSubmitted}</span>
                        </div>
                        {result.error && (
                          <p className="text-sm text-destructive mt-2">{result.error}</p>
                        )}
                      </div>
                    )}

                    <Button
                      className="w-full"
                      size="sm"
                      onClick={() => handleSync(name)}
                      disabled={isSyncing}
                    >
                      {isSyncing ? <Spinner size="sm" /> : <Play className="h-4 w-4 mr-2" />}
                      {t("connectors.sync")}
                    </Button>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </main>
    </div>
  );
}
