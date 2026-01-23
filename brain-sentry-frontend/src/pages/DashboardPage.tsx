import { useState, useEffect } from "react";
import {
  Brain,
  Database,
  Search,
  TrendingUp,
  Activity,
  Clock,
  Tag,
  Plus,
  Filter,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { MemoryCard } from "@/components/memory";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { api } from "@/lib/api";

interface Stats {
  totalMemories: number;
  memoriesByCategory: Record<string, number>;
  memoriesByImportance: Record<string, number>;
  requestsToday: number;
  injectionRate: number;
  avgLatencyMs: number;
  helpfulnessRate: number;
  totalInjections: number;
  activeMemories24h: number;
}

interface RecentMemory {
  id: string;
  content: string;
  summary: string;
  category: string;
  importance: string;
  createdAt: string;
  tags: string[];
}

interface CategoryStat {
  category: string;
  count: number;
}

export function DashboardPage() {
  const { user } = useAuth();
  const { toast } = useToast();
  const [tenantId, setTenantId] = useState(user?.tenantId || "default");

  const [stats, setStats] = useState<Stats | null>(null);
  const [memories, setMemories] = useState<RecentMemory[]>([]);
  const [statsLoading, setStatsLoading] = useState(true);
  const [memoriesLoading, setMemoriesLoading] = useState(true);

  const fetchData = async () => {
    setStatsLoading(true);
    setMemoriesLoading(true);
    try {
      const [statsData, memoriesData] = await Promise.all([
        api.getStats(),
        api.getMemories(0, 6),
      ]);
      setStats(statsData);
      setMemories(memoriesData.memories);
    } catch (error) {
      console.error("Error fetching dashboard data:", error);
    } finally {
      setStatsLoading(false);
      setMemoriesLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleRefresh = () => {
    fetchData();
    toast({
      title: "Dashboard atualizado",
      description: "Os dados foram atualizados com sucesso.",
      variant: "info",
    });
  };

  // Build categories from stats API
  const categories = Object.entries(stats?.memoriesByCategory || {})
    .map(([category, count]) => ({ category, count }))
    .sort((a, b) => b.count - a.count);

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white -mx-0">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Brain className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">Brain Sentry</h1>
                <p className="text-xs text-white/80">
                  Sistema de Memória para Desenvolvedores
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="text-xs text-white/80">
                Tenant: <span className="font-medium text-white">{tenantId}</span>
              </div>
              <Button variant="outline" size="sm" className="bg-white/20 border-white/30 text-white hover:bg-white/30" onClick={handleRefresh}>
                <Activity className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Quick Actions */}
        <div className="mb-8 flex items-center justify-between">
          <div className="flex gap-2">
            <Button size="sm" className="bg-gradient-to-r from-brain-primary to-brain-accent hover:from-brain-primary-dark hover:to-brain-accent-dark text-white">
              <Plus className="h-4 w-4 mr-2" />
              Nova Memória
            </Button>
            <Button size="sm" variant="outline">
              <Search className="h-4 w-4 mr-2" />
              Buscar
            </Button>
          </div>
          <Button size="sm" variant="outline">
            <Filter className="h-4 w-4 mr-2" />
            Filtros
          </Button>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <StatsCard
            title="Total de Memórias"
            value={stats?.totalMemories || 0}
            icon={<Database className="h-5 w-5" />}
            loading={statsLoading}
          />
          <StatsCard
            title="Categorias"
            value={Object.keys(stats?.memoriesByCategory || {}).length}
            icon={<Tag className="h-5 w-5" />}
            loading={statsLoading}
          />
          <StatsCard
            title="Memórias Críticas"
            value={stats?.memoriesByImportance?.CRITICAL || 0}
            icon={<TrendingUp className="h-5 w-5" />}
            loading={statsLoading}
          />
          <StatsCard
            title="Ativas 24h"
            value={stats?.activeMemories24h || 0}
            icon={<Clock className="h-5 w-5" />}
            loading={statsLoading}
          />
        </div>

        {/* Main Content */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Recent Memories */}
          <div className="lg:col-span-2">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Memórias Recentes</CardTitle>
                <Button variant="ghost" size="sm">Ver todas</Button>
              </CardHeader>
              <CardContent>
                {memoriesLoading ? (
                  <div className="flex justify-center py-8">
                    <Spinner size="lg" />
                  </div>
                ) : memories.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                    <Brain className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p>Nenhuma memória encontrada</p>
                    <Button className="mt-4" size="sm">
                      Criar primeira memória
                    </Button>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {memories.map((memory) => (
                      <MemoryCard key={memory.id} memory={memory} />
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Categories Breakdown */}
          <div>
            <Card>
              <CardHeader>
                <CardTitle>Por Categoria</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {categories.map((cat) => (
                    <div key={cat.category} className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div
                          className="h-2 w-2 rounded-full"
                          style={{
                            backgroundColor: getCategoryColor(cat.category),
                          }}
                        />
                        <span className="text-sm capitalize">{cat.category.toLowerCase()}</span>
                      </div>
                      <span className="text-sm font-medium">{cat.count}</span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            {/* Quick Tips */}
            <Card className="mt-4">
              <CardHeader>
                <CardTitle className="text-sm">Dica Rápida</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Use hashtags nos seus prompts para contextuais automáticos.
                  Exemplo: "#spring-boot #rest #api"
                </p>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  );
}

interface StatsCardProps {
  title: string;
  value: number | string;
  icon?: React.ReactNode;
  loading?: boolean;
  suffix?: string;
  trend?: string;
}

function StatsCard({ title, value, icon, loading, suffix, trend }: StatsCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-muted-foreground">{title}</p>
            {loading ? (
              <Spinner size="sm" />
            ) : (
              <p className="text-base font-bold leading-tight">
                {value}
                {suffix}
              </p>
            )}
            {trend && (
              <p className="text-xs text-green-600 dark:text-green-400 mt-1">
                {trend}
              </p>
            )}
          </div>
          {icon && (
            <div className="p-3 bg-gradient-to-br from-brain-primary to-brain-accent rounded-lg text-white">
              {icon}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function getCategoryColor(category: string): string {
  const colors: Record<string, string> = {
    DECISION: "#3b82f6",
    PATTERN: "#10b981",
    ANTIPATTERN: "#ef4444",
    DOMAIN: "#f59e0b",
    BUG: "#dc2626",
    OPTIMIZATION: "#8b5cf6",
    INTEGRATION: "#06b6d4",
  };
  return colors[category] || "#6b7280";
}
