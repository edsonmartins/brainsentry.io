import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, Zap, TrendingUp, Loader2 } from "lucide-react";
import { api, getErrorMessage } from "@/lib/api";

export default function AnalyticsAdminPage() {
  const [stats, setStats] = useState({
    totalMemories: 0,
    byCategory: {} as Record<string, number>,
    byImportance: {} as Record<string, number>,
    avgInjectionRate: 0,
    avgHelpfulnessRate: 0,
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
      value: `${(stats.avgInjectionRate * 100).toFixed(1)}%`,
      change: "+2.3%",
      icon: Zap,
      color: "text-green-600",
    },
    {
      title: "Satisfação",
      value: `${(stats.avgHelpfulnessRate * 100).toFixed(0)}%`,
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
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Analytics</h1>
        <p className="text-sm text-muted-foreground">
          Métricas e estatísticas do sistema
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statsData.map((stat) => (
          <Card key={stat.title}>
            <CardContent className="p-6">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">{stat.title}</p>
                  <p className="text-2xl font-bold">{stat.value}</p>
                </div>
                <div className="text-xs text-muted-foreground">{stat.change}</div>
              </div>
              <div className="flex items-center gap-2">
                <stat.icon className={`h-4 w-4 ${stat.color}`} />
              </div>
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
                {Object.entries(stats.byCategory).map(([category, count]) => (
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
                {Object.keys(stats.byCategory).length === 0 && (
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
                {Object.entries(stats.byImportance).map(([importance, count]) => (
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
                {Object.keys(stats.byImportance).length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-8">
                    Nenhuma memória cadastrada
                  </p>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
