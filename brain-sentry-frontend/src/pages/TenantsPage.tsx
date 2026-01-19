import { useState } from "react";
import {
  Building2,
  Plus,
  Search,
  Edit,
  Trash2,
  Users,
  Database,
  Settings,
  RefreshCw,
  Key,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Input, SearchInput } from "@/components/ui/filter";
import { Pagination, SimplePagination } from "@/components/ui/pagination";
import { Spinner, Skeleton } from "@/components/ui/spinner";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "./ui";
import { useFetch } from "@/hooks";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

interface Tenant {
  id: string;
  name: string;
  slug: string;
  active: boolean;
  maxMemories?: number;
  maxUsers?: number;
  createdAt: string;
  settings?: Record<string, unknown>;
}

interface TenantStats {
  tenantId: string;
  memoryCount: number;
  userCount: number;
  relationshipCount: number;
}

interface CreateTenantRequest {
  name: string;
  slug: string;
  maxMemories?: number;
  maxUsers?: number;
  settings?: Record<string, unknown>;
}

interface UpdateTenantRequest {
  name?: string;
  active?: boolean;
  maxMemories?: number;
  maxUsers?: number;
  settings?: Record<string, unknown>;
}

export function TenantsPage() {
  const { user: currentUser } = useAuth();
  const { toast } = useToast();
  const tenantId = currentUser?.tenantId || "default";

  // State
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null);
  const [formData, setFormData] = useState<CreateTenantRequest>({
    name: "",
    slug: "",
    maxMemories: 1000,
    maxUsers: 10,
  });

  // Fetch tenants
  const {
    data: tenantsData,
    isLoading,
    error,
    refetch,
  } = useFetch<{ content: Tenant[]; totalElements: number }>(
    `${API_URL}/v1/tenants?page=${page - 1}&size=${pageSize}`
  );

  // Fetch stats for each tenant
  const { data: statsData } = useFetch<TenantStats[]>(
    `${API_URL}/v1/tenants/stats`
  );

  const tenants = tenantsData?.content || [];
  const totalElements = tenantsData?.totalElements || 0;
  const totalPages = Math.ceil(totalElements / pageSize);

  const statsMap = new Map(statsData?.map((s) => [s.tenantId, s]) || []);

  // Filter tenants by search
  const filteredTenants = tenants.filter(
    (t) =>
      t.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      t.slug.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Generate slug from name
  const generateSlug = (name: string) => {
    return name
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "");
  };

  // Create tenant
  const handleCreateTenant = async () => {
    try {
      const response = await fetch(`${API_URL}/v1/tenants`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Tenant-ID": tenantId,
        },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || "Failed to create tenant");
      }

      toast({
        title: "Tenant criado",
        description: `O tenant "${formData.name}" foi criado com sucesso.`,
        variant: "success",
      });

      setShowCreateDialog(false);
      resetForm();
      refetch?.();
    } catch (err) {
      toast({
        title: "Erro",
        description: (err as Error).message || "Não foi possível criar o tenant.",
        variant: "error",
      });
    }
  };

  // Update tenant
  const handleUpdateTenant = async () => {
    if (!selectedTenant) return;

    try {
      const response = await fetch(`${API_URL}/v1/tenants/${selectedTenant.id}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          "X-Tenant-ID": tenantId,
        },
        body: JSON.stringify({
          name: formData.name,
          active: formData.active ?? true,
          maxMemories: formData.maxMemories,
          maxUsers: formData.maxUsers,
        } as UpdateTenantRequest),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || "Failed to update tenant");
      }

      toast({
        title: "Tenant atualizado",
        description: `O tenant "${selectedTenant.name}" foi atualizado.`,
        variant: "success",
      });

      setShowEditDialog(false);
      setSelectedTenant(null);
      refetch?.();
    } catch (err) {
      toast({
        title: "Erro",
        description: (err as Error).message || "Não foi possível atualizar o tenant.",
        variant: "error",
      });
    }
  };

  // Delete tenant
  const handleDeleteTenant = async (tenantIdToDelete: string, name: string) => {
    if (
      !confirm(
        `Tem certeza que deseja excluir o tenant "${name}"? Todos os dados associados serão perdidos.`
      )
    ) {
      return;
    }

    try {
      const response = await fetch(`${API_URL}/v1/tenants/${tenantIdToDelete}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          "X-Tenant-ID": tenantId,
        },
      });

      if (!response.ok) {
        throw new Error("Failed to delete tenant");
      }

      toast({
        title: "Tenant excluído",
        description: `O tenant "${name}" foi excluído.`,
        variant: "success",
      });

      refetch?.();
    } catch (err) {
      toast({
        title: "Erro",
        description: "Não foi possível excluir o tenant.",
        variant: "error",
      });
    }
  };

  const resetForm = () => {
    setFormData({
      name: "",
      slug: "",
      maxMemories: 1000,
      maxUsers: 10,
    });
  };

  const openEditDialog = (tenant: Tenant) => {
    setSelectedTenant(tenant);
    setFormData({
      name: tenant.name,
      slug: tenant.slug,
      maxMemories: tenant.maxMemories,
      maxUsers: tenant.maxUsers,
      active: tenant.active,
    });
    setShowEditDialog(true);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString("pt-BR");
  };

  const getTenantStats = (tenant: Tenant) => {
    return statsMap.get(tenant.id);
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Building2 className="h-6 w-6 text-primary" />
              </div>
              <div>
                <h1 className="text-2xl font-bold">Tenants</h1>
                <p className="text-sm text-muted-foreground">
                  Gerencie as organizações do sistema
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={() => refetch?.()}>
                <RefreshCw className="h-4 w-4" />
              </Button>
              <Button onClick={() => setShowCreateDialog(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Novo Tenant
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Search */}
        <div className="mb-6">
          <SearchInput
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="Buscar tenants por nome ou slug..."
          />
        </div>

        {/* Tenants Grid */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 6 }).map((_, i) => (
              <Card key={i}>
                <CardContent className="p-6">
                  <Skeleton variant="text" width="60%" />
                  <Skeleton variant="text" width="40%" />
                  <Skeleton variant="text" width="80%" />
                </CardContent>
              </Card>
            ))}
          </div>
        ) : filteredTenants.length === 0 ? (
          <Card>
            <CardContent className="text-center py-12 text-muted-foreground">
              <Building2 className="h-16 w-16 mx-auto mb-4 opacity-50" />
              <h3 className="text-lg font-semibold mb-2">
                {searchQuery ? "Nenhum tenant encontrado" : "Nenhum tenant cadastrado"}
              </h3>
              <p className="mb-4">
                {searchQuery
                  ? "Tente buscar com outro termo."
                  : "Comece adicionando um novo tenant ao sistema."}
              </p>
              {!searchQuery && (
                <Button onClick={() => setShowCreateDialog(true)}>
                  <Plus className="h-4 w-4 mr-2" />
                  Adicionar Tenant
                </Button>
              )}
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredTenants.map((tenant) => {
              const stats = getTenantStats(tenant);
              const isCurrentTenant = tenant.id === tenantId;

              return (
                <Card
                  key={tenant.id}
                  className={`transition-all hover:shadow-md ${
                    isCurrentTenant ? "border-primary" : ""
                  }`}
                >
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className="p-2 bg-primary/10 rounded-lg">
                          <Building2 className="h-5 w-5 text-primary" />
                        </div>
                        <div>
                          <CardTitle className="text-lg">{tenant.name}</CardTitle>
                          <p className="text-xs text-muted-foreground font-mono">
                            @{tenant.slug}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-1">
                        <div
                          className={`w-2 h-2 rounded-full ${
                            tenant.active ? "bg-green-500" : "bg-red-500"
                          }`}
                        />
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {/* Stats */}
                    {stats && (
                      <div className="grid grid-cols-3 gap-2 text-center">
                        <div className="p-2 bg-muted/50 rounded-lg">
                          <Database className="h-4 w-4 mx-auto mb-1 text-muted-foreground" />
                          <p className="text-lg font-semibold">{stats.memoryCount}</p>
                          <p className="text-xs text-muted-foreground">Memórias</p>
                        </div>
                        <div className="p-2 bg-muted/50 rounded-lg">
                          <Users className="h-4 w-4 mx-auto mb-1 text-muted-foreground" />
                          <p className="text-lg font-semibold">{stats.userCount}</p>
                          <p className="text-xs text-muted-foreground">Usuários</p>
                        </div>
                        <div className="p-2 bg-muted/50 rounded-lg">
                          <Key className="h-4 w-4 mx-auto mb-1 text-muted-foreground" />
                          <p className="text-lg font-semibold">{stats.relationshipCount}</p>
                          <p className="text-xs text-muted-foreground">Relações</p>
                        </div>
                      </div>
                    )}

                    {/* Limits */}
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Limite de memórias</span>
                        <span className="font-medium">
                          {tenant.maxMemories || "Ilimitado"}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Limite de usuários</span>
                        <span className="font-medium">{tenant.maxUsers || "Ilimitado"}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Criado em</span>
                        <span className="font-medium text-xs">
                          {formatDate(tenant.createdAt)}
                        </span>
                      </div>
                    </div>

                    {/* Actions */}
                    <div className="flex gap-2 pt-2 border-t">
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1"
                        onClick={() => openEditDialog(tenant)}
                      >
                        <Edit className="h-3 w-3 mr-1" />
                        Editar
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleDeleteTenant(tenant.id, tenant.name)}
                        disabled={isCurrentTenant}
                      >
                        <Trash2 className="h-3 w-3 text-destructive" />
                      </Button>
                    </div>

                    {isCurrentTenant && (
                      <div className="text-xs text-center text-primary font-medium bg-primary/10 py-1 rounded">
                        Tenant atual
                      </div>
                    )}
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="mt-6">
            <SimplePagination
              currentPage={page}
              totalPages={totalPages}
              onPageChange={setPage}
            />
          </div>
        )}
      </main>

      {/* Create Tenant Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Novo Tenant</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <label className="text-sm font-medium mb-2">Nome</label>
              <Input
                type="text"
                value={formData.name}
                onChange={(e) => {
                  const name = e.target.value;
                  setFormData({
                    ...formData,
                    name,
                    slug: generateSlug(name),
                  });
                }}
                placeholder="Minha Organização"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2">Slug (identificador único)</label>
              <Input
                type="text"
                value={formData.slug}
                onChange={(e) =>
                  setFormData({ ...formData, slug: e.target.value.toLowerCase() })
                }
                placeholder="minha-organizacao"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Usado em URLs e IDs. Deve ser único.
              </p>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium mb-2">Limite de Memórias</label>
                <Input
                  type="number"
                  value={formData.maxMemories || ""}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      maxMemories: Number(e.target.value) || undefined,
                    })
                  }
                  placeholder="1000"
                />
              </div>
              <div>
                <label className="text-sm font-medium mb-2">Limite de Usuários</label>
                <Input
                  type="number"
                  value={formData.maxUsers || ""}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      maxUsers: Number(e.target.value) || undefined,
                    })
                  }
                  placeholder="10"
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
              Cancelar
            </Button>
            <Button onClick={handleCreateTenant} disabled={!formData.name || !formData.slug}>
              Criar Tenant
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Tenant Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Editar Tenant</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <label className="text-sm font-medium mb-2">Nome</label>
              <Input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2">Slug</label>
              <Input type="text" value={formData.slug} disabled />
              <p className="text-xs text-muted-foreground mt-1">
                O slug não pode ser alterado após a criação.
              </p>
            </div>
            <div>
              <label className="text-sm font-medium mb-2">Status</label>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, active: true })}
                  className={`px-3 py-1.5 rounded-full text-sm transition-colors ${
                    formData.active !== false
                      ? "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                      : "bg-muted hover:bg-muted/80"
                  }`}
                >
                  Ativo
                </button>
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, active: false })}
                  className={`px-3 py-1.5 rounded-full text-sm transition-colors ${
                    formData.active === false
                      ? "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"
                      : "bg-muted hover:bg-muted/80"
                  }`}
                >
                  Inativo
                </button>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium mb-2">Limite de Memórias</label>
                <Input
                  type="number"
                  value={formData.maxMemories || ""}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      maxMemories: Number(e.target.value) || undefined,
                    })
                  }
                />
              </div>
              <div>
                <label className="text-sm font-medium mb-2">Limite de Usuários</label>
                <Input
                  type="number"
                  value={formData.maxUsers || ""}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      maxUsers: Number(e.target.value) || undefined,
                    })
                  }
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEditDialog(false)}>
              Cancelar
            </Button>
            <Button onClick={handleUpdateTenant}>Salvar Alterações</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
