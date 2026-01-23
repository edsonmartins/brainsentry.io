import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, Zap, TrendingUp, Loader2, BarChart3 } from "lucide-react";
import { api, getErrorMessage } from "@/lib/api";

export default function AnalyticsAdminPage() {
  const [stats, setStats] = useState({
    totalMemories: 0,
    memoriesByCategory: {} as Record<string, number>,
    memoriesByImportance: {} as Record<string, number>,
    requestsToday: 0,
    injectionRate: 0,
    avgLatencyMs: 0,
    helpfulnessRate: 0,
    totalInjections: 0,
    activeMemories24h: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await api.getStats();
        setStats(data);
      } catch (err) {
        setError(getErrorMessage(err));
        console.error("Error fetching stats:", err);
      } finally {
        setLoading(false);
      }
    };
    fetchStats();
  }, []);

  const statsData = [
    {
      title: "Total Memórias",
      value: stats.totalMemories,
      change: "+12%",
      icon: Activity,
      color: "text-blue-600",
    },
    {
      title: "Taxa de Injeção",
      value: `${(stats.injectionRate * 100).toFixed(1)}%`,
      change: "+2.3%",
      icon: Zap,
      color: "text-green-600",
    },
    {
      title: "Satisfação",
      value: `${(stats.helpfulnessRate * 100).toFixed(0)}%`,
      change: "+5%",
      icon: TrendingUp,
      color: "text-purple-600",
    },
  ];

  const categoryColors: Record<string, string> = {
    PATTERN: "bg-blue-500",
    DECISION: "bg-purple-500",
    BUG: "bg-yellow-500",
    REFACTOR: "bg-green-500",
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="ml-2 text-muted-foreground">Carregando estatísticas...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 text-destructive p-4 rounded-md">
        <p className="font-medium">Erro ao carregar estatísticas</p>
        <p className="text-sm">{error}</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white -mx-0">
        <div className="px-6 py-3">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-white/20 rounded-lg backdrop-blur-sm">
              <BarChart3 className="h-6 w-6 text-white" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">Analytics</h1>
              <p className="text-sm text-white/80">
                Métricas e estatísticas do sistema
              </p>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          {statsData.map((stat) => (
            <Card key={stat.title} className="shadow-sm">
              <CardContent className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">{stat.title}</p>
                    <p className="text-2xl font-bold">{stat.value}</p>
                  </div>
                  <div className="p-3 bg-gradient-to-br from-brain-primary to-brain-accent rounded-lg text-white">
                    <stat.icon className="h-5 w-5" />
                  </div>
                </div>
                <div className="text-xs text-green-600 dark:text-green-400">{stat.change}</div>
              </CardContent>
            </Card>
          ))}
        </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Distribuição por Categoria</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center">
              <div className="w-full space-y-3">
                {Object.entries(stats.memoriesByCategory).map(([category, count]) => (
                  <div key={category}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm text-muted-foreground">{category}</span>
                      <span className="text-sm font-medium">{count}</span>
                    </div>
                    <div
                      className={`h-2 rounded-md ${categoryColors[category] || "bg-gray-500"}`}
                      style={{ width: `${Math.min((count / stats.totalMemories) * 100, 100)}%` }}
                    ></div>
                  </div>
                ))}
                {Object.keys(stats.memoriesByCategory).length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-8">
                    Nenhuma memória cadastrada
                  </p>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Distribuição por Importância</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center">
              <div className="w-full space-y-3">
                {Object.entries(stats.memoriesByImportance).map(([importance, count]) => (
                  <div key={importance}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm text-muted-foreground">{importance}</span>
                      <span className="text-sm font-medium">{count}</span>
                    </div>
                    <div
                      className={`h-2 rounded-md ${
                        importance === "CRITICAL" ? "bg-red-500" :
                        importance === "IMPORTANT" ? "bg-orange-500" :
                        "bg-gray-500"
                      }`}
                      style={{ width: `${Math.min((count / stats.totalMemories) * 100, 100)}%` }}
                    ></div>
                  </div>
                ))}
                {Object.keys(stats.memoriesByImportance).length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-8">
                    Nenhuma memória cadastrada
                  </p>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
      </main>
    </div>
  );
}
