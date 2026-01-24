import { useState, useEffect } from "react";
import {
  Shield,
  Activity,
  AlertTriangle,
  Clock,
  User,
  Filter,
  Download,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  Info,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Input, FilterBar } from "@/components/ui/filter";
import { Pagination, SimplePagination } from "@/components/ui/pagination";
import { Spinner, Skeleton } from "@/components/ui/spinner";
import { useFetch, useDebounce } from "@/hooks";
import { CategoryTag, ImportanceTag, ReadOnlyTags } from "@/components/ui/tags";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

interface AuditLog {
  id: string;
  eventType: string;
  timestamp: string;
  userId: string;
  sessionId: string;
  outcome: string;
  errorMessage?: string;
  memoriesAccessed?: string[];
  memoriesCreated?: string[];
  memoriesModified?: string[];
  inputData?: Record<string, unknown>;
  outputData?: Record<string, unknown>;
}

interface StatsResponse {
  totalEvents: number;
  eventsByType: Record<string, number>;
  eventsByUser: Record<string, number>;
  recentActivity: number;
}

const EVENT_TYPE_LABELS: Record<string, string> = {
  context_injection: "Injeção de Contexto",
  memory_created: "Memória Criada",
  memory_updated: "Memória Atualizada",
  memory_deleted: "Memória Excluída",
  relationship_created: "Relacionamento Criado",
  error: "Erro",
};

const OUTCOME_COLORS: Record<string, string> = {
  success: "text-green-600 dark:text-green-400",
  failed: "text-red-600 dark:text-red-400",
  rejected: "text-yellow-600 dark:text-yellow-400",
  partial: "text-orange-600 dark:text-orange-400",
};

export function AuditPage() {
  const { user } = useAuth();
  const { toast } = useToast();
  const tenantId = user?.tenantId || "a9f814d2-4dae-41f3-851b-8aa3d4706561";

  // Filters
  const [searchTerm, setSearchTerm] = useState("");
  const [eventType, setEventType] = useState("");
  const [dateRange, setDateRange] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);

  // Fetch audit logs
  const { data, isLoading, error, refetch } = useFetch<{ content: AuditLog[]; totalElements: number }>(
    `${API_URL}/v1/audit-logs?page=${page - 1}&size=${pageSize}`
  );

  // Fetch stats
  const { data: stats } = useFetch<StatsResponse>(`${API_URL}/v1/audit-logs/stats`);

  const logs = data?.content || [];
  const totalElements = data?.totalElements || 0;
  const totalPages = Math.ceil(totalElements / pageSize);

  const handleExport = async () => {
    try {
      const response = await fetch(`${API_URL}/v1/audit-logs/export`, {
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) throw new Error("Failed to export");

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `audit-logs-${new Date().toISOString()}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);

      toast({
        title: "Exportação concluída",
        description: "Os logs foram exportados com sucesso.",
        variant: "success",
      });
    } catch {
      toast({
        title: "Erro na exportação",
        description: "Não foi possível exportar os logs.",
        variant: "error",
      });
    }
  };

  const getEventIcon = (eventType: string) => {
    if (eventType.includes("error")) return AlertTriangle;
    if (eventType.includes("created")) return Shield;
    return Activity;
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString("pt-BR");
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white -mx-0">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Shield className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">Auditoria</h1>
                <p className="text-xs text-white/80">
                  Logs de operações do sistema
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" className="bg-white/20 text-white border-white/30 hover:bg-white/30" onClick={() => refetch?.()}>
                <RefreshCw className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="sm" className="bg-white text-brain-primary hover:bg-white/90 border-0" onClick={handleExport}>
                <Download className="h-4 w-4 mr-2" />
                Exportar
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
          <StatCard
            title="Total de Eventos"
            value={stats?.totalEvents || 0}
            icon={<Activity className="h-5 w-5" />}
            loading={!stats}
          />
          <StatCard
            title="Últimas 24h"
            value={stats?.recentActivity || 0}
            icon={<Clock className="h-5 w-5" />}
            loading={!stats}
          />
          <StatCard
            title="Usuários Ativos"
            value={Object.keys(stats?.eventsByUser || {}).length}
            icon={<User className="h-5 w-5" />}
            loading={!stats}
          />
          <StatCard
            title="Tipos de Eventos"
            value={Object.keys(stats?.eventsByType || {}).length}
            icon={<Filter className="h-5 w-5" />}
            loading={!stats}
          />
        </div>

        {/* Filters */}
        <div className="mb-6">
          <FilterBar
            searchValue={searchTerm}
            onSearchChange={setSearchTerm}
            filters={[
              {
                key: "eventType",
                label: "Tipo",
                options: [
                  { value: "", label: "Todos" },
                  { value: "context_injection", label: "Injeção" },
                  { value: "memory_created", label: "Criação" },
                  { value: "memory_updated", label: "Atualização" },
                  { value: "memory_deleted", label: "Exclusão" },
                  { value: "relationship_created", label: "Relacionamento" },
                ],
                value: eventType,
                onChange: setEventType,
              },
            ]}
          />
        </div>

        {/* Logs Table */}
        <Card>
          <CardHeader>
            <CardTitle>Histórico de Eventos</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-2">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="flex items-center gap-4 p-3 border-b">
                    <Skeleton variant="circular" width={40} height={40} />
                    <Skeleton variant="text" width="30%" />
                    <Skeleton variant="text" width="20%" />
                  </div>
                ))}
              </div>
            ) : logs.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <Shield className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>Nenhum evento de auditoria encontrado</p>
              </div>
            ) : (
              <>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b text-left text-sm text-muted-foreground">
                        <th className="p-3">Tipo</th>
                        <th className="p-3">Usuário</th>
                        <th className="p-3">Sessão</th>
                        <th className="p-3">Resultado</th>
                        <th className="p-3">Dados</th>
                        <th className="p-3">Horário</th>
                      </tr>
                    </thead>
                    <tbody>
                      {logs.map((log) => (
                        <tr key={log.id} className="border-b hover:bg-muted/50">
                          <td className="p-3">
                            <div className="flex items-center gap-2">
                              {(() => {
                                const Icon = getEventIcon(log.eventType);
                                return <Icon className="h-4 w-4" />;
                              })()}
                              <span className="text-sm">
                                {EVENT_TYPE_LABELS[log.eventType] || log.eventType}
                              </span>
                            </div>
                          </td>
                          <td className="p-3 text-sm">{log.userId || "-"}</td>
                          <td className="p-3 text-sm font-mono">{log.sessionId || "-"}</td>
                          <td className="p-3">
                            <span className={`text-sm font-medium ${OUTCOME_COLORS[log.outcome] || ""}`}>
                              {log.outcome}
                            </span>
                          </td>
                          <td className="p-3 text-sm">
                            {log.errorMessage ? (
                              <span className="text-destructive">{log.errorMessage}</span>
                            ) : log.memoriesAccessed && log.memoriesAccessed.length > 0 ? (
                              <span>{log.memoriesAccessed.length} acessos</span>
                            ) : log.memoriesCreated && log.memoriesCreated.length > 0 ? (
                              <span>{log.memoriesCreated.length} criadas</span>
                            ) : log.memoriesModified && log.memoriesModified.length > 0 ? (
                              <span>{log.memoriesModified.length} modificadas</span>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </td>
                          <td className="p-3 text-sm text-muted-foreground">
                            {formatTimestamp(log.timestamp)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* Pagination */}
                {totalPages > 1 && (
                  <div className="mt-4">
                    <SimplePagination
                      currentPage={page}
                      totalPages={totalPages}
                      onPageChange={setPage}
                    />
                  </div>
                )}
              </>
            )}
          </CardContent>
        </Card>

        {/* Events by Type Chart */}
        {stats?.eventsByType && (
          <Card className="mt-6">
            <CardHeader>
              <CardTitle>Eventos por Tipo</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {Object.entries(stats.eventsByType)
                  .sort(([, a], [, b]) => b - a)
                  .map(([type, count]) => (
                    <div key={type} className="flex items-center gap-3">
                      <span className="flex-1 text-sm">{EVENT_TYPE_LABELS[type] || type}</span>
                      <div className="w-full bg-muted rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full"
                          style={{ width: `${(count / stats.totalEvents) * 100}%` }}
                        />
                      </div>
                      <span className="text-sm font-medium w-12 text-right">{count}</span>
                    </div>
                  ))}
              </div>
            </CardContent>
          </Card>
        )}
      </main>
    </div>
  );
}

interface StatCardProps {
  title: string;
  value: number;
  icon?: React.ReactNode;
  loading?: boolean;
}

function StatCard({ title, value, icon, loading }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-muted-foreground">{title}</p>
            {loading ? (
              <Skeleton variant="text" width={80} />
            ) : (
              <p className="text-base font-bold leading-tight">{value}</p>
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
