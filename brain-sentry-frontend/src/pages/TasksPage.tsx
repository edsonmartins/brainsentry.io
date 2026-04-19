import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { ListTodo, RefreshCw, Activity, Clock, CheckCircle, XCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useToast } from "@/components/ui/toast";
import { api } from "@/lib/api";

interface TaskMetrics {
  processed: number;
  failed: number;
  recovered: number;
}

interface PendingResult {
  pending: number;
}

export default function TasksPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [metrics, setMetrics] = useState<TaskMetrics | null>(null);
  const [pending, setPending] = useState<number>(0);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [metricsResp, pendingResp] = await Promise.all([
        api.axiosInstance.get<TaskMetrics>("/v1/tasks/metrics"),
        api.axiosInstance.get<PendingResult>("/v1/tasks/pending"),
      ]);
      setMetrics(metricsResp.data);
      setPending(pendingResp.data.pending);
    } catch (err: any) {
      toast({ title: t("tasks.error"), description: err.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <ListTodo className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("tasks.title")}</h1>
                <p className="text-xs text-white/80">
                  {t("tasks.subtitle")}
                </p>
              </div>
            </div>
            <Button variant="outline" size="sm" className="bg-white/20 text-white border-white/30 hover:bg-white/30" onClick={fetchData}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {loading && !metrics ? (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-muted-foreground">{t("tasks.pending")}</p>
                      <p className="text-2xl font-bold">{pending}</p>
                    </div>
                    <div className="p-3 bg-yellow-100 dark:bg-yellow-900/30 rounded-lg">
                      <Clock className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-muted-foreground">{t("tasks.processed")}</p>
                      <p className="text-2xl font-bold">{metrics?.processed || 0}</p>
                    </div>
                    <div className="p-3 bg-green-100 dark:bg-green-900/30 rounded-lg">
                      <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400" />
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-muted-foreground">{t("tasks.failed")}</p>
                      <p className="text-2xl font-bold">{metrics?.failed || 0}</p>
                    </div>
                    <div className="p-3 bg-red-100 dark:bg-red-900/30 rounded-lg">
                      <XCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-muted-foreground">{t("tasks.recovered")}</p>
                      <p className="text-2xl font-bold">{metrics?.recovered || 0}</p>
                    </div>
                    <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                      <Activity className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {metrics && (metrics.processed > 0 || metrics.failed > 0) && (
              <Card>
                <CardHeader>
                  <CardTitle>{t("tasks.successRate")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div>
                      <div className="flex justify-between mb-1 text-sm">
                        <span>{t("tasks.success")}</span>
                        <span>{metrics.processed > 0 ? ((metrics.processed / (metrics.processed + metrics.failed)) * 100).toFixed(1) : 0}%</span>
                      </div>
                      <div className="w-full bg-muted rounded-full h-2">
                        <div
                          className="bg-green-500 h-2 rounded-full"
                          style={{
                            width: `${metrics.processed > 0 ? (metrics.processed / (metrics.processed + metrics.failed)) * 100 : 0}%`,
                          }}
                        />
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </>
        )}
      </main>
    </div>
  );
}
