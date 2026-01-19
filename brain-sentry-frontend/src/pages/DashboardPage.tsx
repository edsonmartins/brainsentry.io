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
import { useFetch } from "@/hooks";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { MemoryCard } from "@/components/memory";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";

interface Stats {
  totalMemories: number;
  totalCategories: number;
  recentActivity: number;
  avgImportance: string;
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

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export function DashboardPage() {
  const { user } = useAuth();
  const { toast } = useToast();
  const [tenantId, setTenantId] = useState(user?.tenantId || "default");

  const {
    data: stats,
    isLoading: statsLoading,
    refetch: refetchStats,
  } = useFetch<Stats>(`${API_URL}/v1/stats/overview`);

  const {
    data: memoriesData,
    isLoading: memoriesLoading,
  } = useFetch<{ memories: RecentMemory[]; totalElements: number }>(
    `${API_URL}/v1/memories?page=0&size=6`
  );

  const {
    data: categoriesData,
  } = useFetch<{ content: CategoryStat[] }>(
    `${API_URL}/v1/stats/by-category`
  );

  const memories = memoriesData?.memories || [];
  const categories = categoriesData?.content || [];

  const handleRefresh = () => {
    refetchStats();
    toast({
      title: "Dashboard atualizado",
      description: "Os dados foram atualizados com sucesso.",
      variant: "info",
    });
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Brain className="h-6 w-6 text-primary" />
              </div>
              <div>
                <h1 className="text-2xl font-bold">Brain Sentry</h1>
                <p className="text-sm text-muted-foreground">
                  Sistema de Memória para Desenvolvedores
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="text-sm text-muted-foreground">
                Tenant: <span className="font-medium">{tenantId}</span>
              </div>
              <Button variant="outline" size="sm" onClick={handleRefresh}>
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
            <Button size="sm">
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
            trend="+12%"
          />
          <StatsCard
            title="Categorias"
            value={stats?.totalCategories || 0}
            icon={<Tag className="h-5 w-5" />}
            loading={statsLoading}
          />
          <StatsCard
            title="Atividade Recente"
            value={stats?.recentActivity || 0}
            suffix="/24h"
            icon={<Clock className="h-5 w-5" />}
            loading={statsLoading}
          />
          <StatsCard
            title="Importância Média"
            value={stats?.avgImportance || "-"}
            icon={<TrendingUp className="h-5 w-5" />}
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
              <p className="text-2xl font-bold">
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
            <div className="p-3 bg-primary/10 rounded-lg text-primary">
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
