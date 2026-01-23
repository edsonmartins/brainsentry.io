import { useState, useEffect, useCallback } from "react";
import {
  Network,
  Search,
  ArrowRight,
  Trash2,
  RefreshCw,
  Users,
  Package,
  Building,
  MapPin,
  Calendar,
  Tag,
  CircleDot,
  Link2,
  Zap,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { SearchInput } from "@/components/ui/filter";
import { Spinner, Skeleton } from "@/components/ui/spinner";
import { CategoryTag } from "@/components/ui/tags";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { useDebounce } from "@/hooks";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { api } from "@/lib/api/client";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

interface Memory {
  id: string;
  content: string;
  summary: string;
  category: string;
  importance: string;
  tags: string[];
}

interface Relationship {
  id: string;
  fromMemoryId: string;
  toMemoryId: string;
  type: string;
  strength: number;
  createdAt: string;
  fromMemory?: Memory;
  toMemory?: Memory;
}

// Knowledge Graph types (entities extracted from content)
interface EntityNode {
  id: string;
  name: string;
  type: string;
  sourceMemoryId?: string;
  properties?: Record<string, string>;
}

interface EntityEdge {
  id: string;
  sourceId: string;
  targetId: string;
  sourceName: string;
  targetName: string;
  type: string;
  properties?: Record<string, string>;
}

interface KnowledgeGraphResponse {
  nodes: EntityNode[];
  edges: EntityEdge[];
  totalNodes: number;
  totalEdges: number;
}

const RELATIONSHIP_TYPES = [
  { value: "RELATED", label: "Relacionado" },
  { value: "DEPENDS_ON", label: "Depende de" },
  { value: "DEPENDENT_OF", label: "Dependente" },
  { value: "CONTRADICTS", label: "Contradiz" },
  { value: "SIMILAR_TO", label: "Similar a" },
  { value: "EXTENDS", label: "Estende" },
];

// Entity type colors and icons
const ENTITY_TYPE_CONFIG: Record<string, { color: string; bgColor: string; icon: typeof Users }> = {
  PESSOA: { color: "text-blue-600", bgColor: "bg-blue-100", icon: Users },
  CLIENTE: { color: "text-blue-600", bgColor: "bg-blue-100", icon: Users },
  VENDEDOR: { color: "text-green-600", bgColor: "bg-green-100", icon: Users },
  ORGANIZACAO: { color: "text-purple-600", bgColor: "bg-purple-100", icon: Building },
  EMPRESA: { color: "text-purple-600", bgColor: "bg-purple-100", icon: Building },
  PRODUTO: { color: "text-orange-600", bgColor: "bg-orange-100", icon: Package },
  PEDIDO: { color: "text-red-600", bgColor: "bg-red-100", icon: Tag },
  LOCAL: { color: "text-cyan-600", bgColor: "bg-cyan-100", icon: MapPin },
  ENDERECO: { color: "text-cyan-600", bgColor: "bg-cyan-100", icon: MapPin },
  DATA: { color: "text-yellow-600", bgColor: "bg-yellow-100", icon: Calendar },
  EVENTO: { color: "text-pink-600", bgColor: "bg-pink-100", icon: Calendar },
  CONCEITO: { color: "text-gray-600", bgColor: "bg-gray-100", icon: CircleDot },
};

function getEntityConfig(type: string) {
  return ENTITY_TYPE_CONFIG[type?.toUpperCase()] || {
    color: "text-gray-600",
    bgColor: "bg-gray-100",
    icon: CircleDot,
  };
}

export function RelationshipsPage() {
  const { user } = useAuth();
  const { toast } = useToast();
  const tenantId = user?.tenantId || "default";

  // Tab state
  const [activeTab, setActiveTab] = useState<"knowledge" | "memory">("knowledge");

  // State
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedSearchQuery = useDebounce(searchQuery, 500);
  const [selectedMemory, setSelectedMemory] = useState<Memory | null>(null);
  const [relationshipType, setRelationshipType] = useState("RELATED");
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);

  // Memory relationships state
  const [relationships, setRelationships] = useState<Relationship[]>([]);
  const [totalElements, setTotalElements] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  // Knowledge graph state
  const [knowledgeGraph, setKnowledgeGraph] = useState<KnowledgeGraphResponse | null>(null);
  const [isLoadingGraph, setIsLoadingGraph] = useState(false);

  // Search state
  const [searchResults, setSearchResults] = useState<Memory[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  // Highlighted memory state (for showing connections before creating)
  const [highlightedMemory, setHighlightedMemory] = useState<Memory | null>(null);
  const [highlightedConnections, setHighlightedConnections] = useState<Relationship[]>([]);
  const [isLoadingConnections, setIsLoadingConnections] = useState(false);

  // Dialog state for target memory selection
  const [targetMemory, setTargetMemory] = useState<Memory | null>(null);
  const [dialogSearchQuery, setDialogSearchQuery] = useState("");
  const [dialogSearchResults, setDialogSearchResults] = useState<Memory[]>([]);
  const [isDialogSearching, setIsDialogSearching] = useState(false);
  const debouncedDialogSearch = useDebounce(dialogSearchQuery, 500);

  // Fetch knowledge graph
  const fetchKnowledgeGraph = useCallback(async () => {
    setIsLoadingGraph(true);
    try {
      const response = await api.axiosInstance.get<KnowledgeGraphResponse>(
        `/v1/memories/knowledge-graph?limit=100`
      );
      setKnowledgeGraph(response.data);
    } catch (err) {
      console.error("Error fetching knowledge graph:", err);
      setKnowledgeGraph(null);
    } finally {
      setIsLoadingGraph(false);
    }
  }, []);

  // Fetch memory relationships
  const fetchRelationships = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await api.axiosInstance.get<{ relationships: Relationship[]; totalElements: number }>(
        `/v1/relationships?page=${page - 1}&size=${pageSize}`
      );
      setRelationships(response.data.relationships || []);
      setTotalElements(response.data.totalElements || 0);
    } catch (err) {
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  }, [page, pageSize]);

  // Search memories
  const shouldSearch = debouncedSearchQuery && debouncedSearchQuery.length >= 2;

  const searchMemories = useCallback(async () => {
    if (!shouldSearch) {
      setSearchResults([]);
      return;
    }
    setIsSearching(true);
    try {
      const response = await api.axiosInstance.post<Memory[]>(
        `/v1/memories/search`,
        { query: debouncedSearchQuery, limit: 10 }
      );
      setSearchResults(response.data || []);
    } catch (err) {
      console.error("Search error:", err);
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  }, [shouldSearch, debouncedSearchQuery]);

  // Fetch connections for a specific memory
  const fetchConnectionsForMemory = useCallback(async (memoryId: string) => {
    setIsLoadingConnections(true);
    try {
      // Try to get relationships from the existing list first, or fetch all and filter
      const response = await api.axiosInstance.get<{ relationships: Relationship[]; totalElements: number }>(
        `/v1/relationships?page=0&size=100`
      );
      const allRelationships = response.data.relationships || [];
      const filtered = allRelationships.filter(
        r => r.fromMemoryId === memoryId || r.toMemoryId === memoryId
      );
      setHighlightedConnections(filtered);
    } catch (err) {
      console.error("Error fetching connections:", err);
      setHighlightedConnections([]);
    } finally {
      setIsLoadingConnections(false);
    }
  }, []);

  // Effect to fetch data on mount
  useEffect(() => {
    fetchKnowledgeGraph();
    fetchRelationships();
  }, [fetchKnowledgeGraph, fetchRelationships]);

  // Effect to search memories when query changes
  useEffect(() => {
    searchMemories();
  }, [searchMemories]);

  // Effect to search in dialog when dialog search query changes
  useEffect(() => {
    const searchInDialog = async () => {
      if (!debouncedDialogSearch || debouncedDialogSearch.length < 2) {
        setDialogSearchResults([]);
        return;
      }
      setIsDialogSearching(true);
      try {
        const response = await api.axiosInstance.post<Memory[]>(
          `/v1/memories/search`,
          { query: debouncedDialogSearch, limit: 10 }
        );
        // Filter out the source memory from results
        const filtered = (response.data || []).filter(m => m.id !== selectedMemory?.id);
        setDialogSearchResults(filtered);
      } catch (err) {
        console.error("Dialog search error:", err);
        setDialogSearchResults([]);
      } finally {
        setIsDialogSearching(false);
      }
    };
    searchInDialog();
  }, [debouncedDialogSearch, selectedMemory?.id]);

  // Reset dialog state when dialog closes
  useEffect(() => {
    if (!showCreateDialog) {
      setTargetMemory(null);
      setDialogSearchQuery("");
      setDialogSearchResults([]);
    }
  }, [showCreateDialog]);

  const refetch = () => {
    fetchKnowledgeGraph();
    fetchRelationships();
  };

  // Reprocess all memories to extract entities
  const [isReprocessing, setIsReprocessing] = useState(false);
  const handleReprocessEntities = async () => {
    setIsReprocessing(true);
    try {
      const response = await api.axiosInstance.post<string>(
        `/v1/memories/extract-all-entities`
      );
      toast({
        title: "Reprocessamento concluído",
        description: response.data || "Entidades extraídas com sucesso.",
        variant: "success",
      });
      // Refresh the knowledge graph after reprocessing
      fetchKnowledgeGraph();
    } catch (err) {
      console.error("Error reprocessing:", err);
      toast({
        title: "Erro no reprocessamento",
        description: (err as Error).message || "Não foi possível reprocessar as memórias.",
        variant: "error",
      });
    } finally {
      setIsReprocessing(false);
    }
  };

  // Create relationship
  const handleCreateRelationship = async () => {
    if (!selectedMemory || !targetMemory) return;

    try {
      const response = await fetch(`${API_URL}/v1/relationships`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Tenant-ID": tenantId,
        },
        body: JSON.stringify({
          fromMemoryId: selectedMemory.id,
          toMemoryId: targetMemory.id,
          type: relationshipType,
          strength: 0.5,
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to create relationship");
      }

      toast({
        title: "Relacionamento criado",
        description: `"${selectedMemory.summary}" → ${relationshipType} → "${targetMemory.summary}"`,
        variant: "success",
      });

      setShowCreateDialog(false);
      refetch();
      setSelectedMemory(null);
      setTargetMemory(null);
      setHighlightedMemory(null);
    } catch (err) {
      toast({
        title: "Erro",
        description: (err as Error).message || "Não foi possível criar o relacionamento.",
        variant: "error",
      });
    }
  };

  // Delete relationship
  const handleDeleteRelationship = async (relationshipId: string) => {
    try {
      const response = await fetch(`${API_URL}/v1/relationships/between`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          "X-Tenant-ID": tenantId,
        },
        body: JSON.stringify({
          relationshipId,
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to delete relationship");
      }

      toast({
        title: "Relacionamento removido",
        variant: "success",
      });

      refetch();
    } catch (err) {
      toast({
        title: "Erro",
        description: (err as Error).message || "Não foi possível remover o relacionamento.",
        variant: "error",
      });
    }
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white -mx-0">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Network className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">Grafo de Conhecimento</h1>
                <p className="text-xs text-white/80">
                  Entidades e relacionamentos extraídos automaticamente
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                className="bg-white/20 border-white/30 text-white hover:bg-white/30"
                onClick={handleReprocessEntities}
                disabled={isReprocessing}
              >
                {isReprocessing ? (
                  <Spinner className="h-4 w-4" />
                ) : (
                  <Zap className="h-4 w-4 mr-1" />
                )}
                {isReprocessing ? "Processando..." : "Reprocessar"}
              </Button>
              <Button variant="outline" size="sm" className="bg-white/20 border-white/30 text-white hover:bg-white/30" onClick={refetch}>
                <RefreshCw className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Tab Navigation */}
        <div className="flex gap-2 mb-6">
          <Button
            variant={activeTab === "knowledge" ? "default" : "outline"}
            onClick={() => setActiveTab("knowledge")}
            className={activeTab === "knowledge" ? "bg-gradient-to-r from-brain-primary to-brain-accent" : ""}
          >
            <Network className="h-4 w-4 mr-2" />
            Entidades Extraídas
          </Button>
          <Button
            variant={activeTab === "memory" ? "default" : "outline"}
            onClick={() => setActiveTab("memory")}
            className={activeTab === "memory" ? "bg-gradient-to-r from-brain-primary to-brain-accent" : ""}
          >
            <ArrowRight className="h-4 w-4 mr-2" />
            Memórias Conectadas
          </Button>
        </div>

        {/* Knowledge Graph Tab */}
        {activeTab === "knowledge" && (
          <>
            {/* Stats Cards */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
              <Card>
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold text-brain-primary">
                    {knowledgeGraph?.totalNodes || 0}
                  </div>
                  <p className="text-sm text-muted-foreground">Entidades</p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold text-brain-accent">
                    {knowledgeGraph?.totalEdges || 0}
                  </div>
                  <p className="text-sm text-muted-foreground">Relacionamentos</p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold text-green-600">
                    {new Set(knowledgeGraph?.nodes.map(n => n.type) || []).size}
                  </div>
                  <p className="text-sm text-muted-foreground">Tipos de Entidade</p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold text-purple-600">
                    {new Set(knowledgeGraph?.edges.map(e => e.type) || []).size}
                  </div>
                  <p className="text-sm text-muted-foreground">Tipos de Relação</p>
                </CardContent>
              </Card>
            </div>

            {/* Entities List */}
            <Card className="mb-6">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Users className="h-5 w-5" />
                  Entidades Extraídas
                </CardTitle>
                <p className="text-sm text-muted-foreground">
                  Pessoas, organizações, produtos e conceitos identificados automaticamente nas mensagens
                </p>
              </CardHeader>
              <CardContent>
                {isLoadingGraph ? (
                  <div className="space-y-4">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div key={i} className="flex items-center gap-4 p-4 border rounded-lg">
                        <Skeleton variant="circular" width={40} height={40} />
                        <Skeleton variant="text" width="40%" />
                        <Skeleton variant="text" width="20%" />
                      </div>
                    ))}
                  </div>
                ) : !knowledgeGraph || knowledgeGraph.nodes.length === 0 ? (
                  <div className="text-center py-12 text-muted-foreground">
                    <Network className="h-16 w-16 mx-auto mb-4 opacity-50" />
                    <h3 className="text-lg font-semibold mb-2">
                      Nenhuma entidade extraída ainda
                    </h3>
                    <p className="mb-4 max-w-md mx-auto">
                      Quando você criar memórias, o sistema irá automaticamente extrair
                      entidades como pessoas, empresas, produtos, pedidos, etc.
                      <br />
                      <span className="text-sm">
                        Exemplo: "Cliente João fez pedido #123 de Laptop ProMax"
                        → Extrai: CLIENTE:João, PEDIDO:#123, PRODUTO:Laptop ProMax
                      </span>
                    </p>
                  </div>
                ) : (
                  <div className="space-y-3 max-h-[400px] overflow-y-auto">
                    {knowledgeGraph.nodes.map((entity) => {
                      const config = getEntityConfig(entity.type);
                      const Icon = config.icon;
                      return (
                        <div
                          key={entity.id}
                          className="flex items-center gap-4 p-3 border rounded-lg hover:bg-accent transition-colors"
                        >
                          <div className={`p-2 rounded-full ${config.bgColor}`}>
                            <Icon className={`h-4 w-4 ${config.color}`} />
                          </div>
                          <div className="flex-1">
                            <p className="font-medium">{entity.name}</p>
                            <p className="text-xs text-muted-foreground">
                              Tipo: {entity.type}
                            </p>
                          </div>
                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${config.bgColor} ${config.color}`}>
                            {entity.type}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Entity Relationships */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <ArrowRight className="h-5 w-5" />
                  Relacionamentos entre Entidades
                </CardTitle>
                <p className="text-sm text-muted-foreground">
                  Conexões identificadas automaticamente (ex: REALIZOU, ATENDEU, CONTEM)
                </p>
              </CardHeader>
              <CardContent>
                {isLoadingGraph ? (
                  <div className="space-y-4">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div key={i} className="flex items-center gap-4 p-4 border rounded-lg">
                        <Skeleton variant="text" width="30%" />
                        <Skeleton variant="text" width="20%" />
                        <Skeleton variant="text" width="30%" />
                      </div>
                    ))}
                  </div>
                ) : !knowledgeGraph || knowledgeGraph.edges.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                    <p className="text-sm">
                      Nenhum relacionamento entre entidades encontrado.
                    </p>
                  </div>
                ) : (
                  <div className="space-y-3 max-h-[400px] overflow-y-auto">
                    {knowledgeGraph.edges.map((edge) => (
                      <div
                        key={edge.id}
                        className="flex items-center gap-3 p-3 border rounded-lg hover:bg-accent transition-colors"
                      >
                        <span className="font-medium text-sm">{edge.sourceName}</span>
                        <span className="px-2 py-1 rounded-full bg-brain-primary/10 text-brain-primary text-xs font-medium">
                          {edge.type}
                        </span>
                        <ArrowRight className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium text-sm">{edge.targetName}</span>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </>
        )}

        {/* Memory Relationships Tab */}
        {activeTab === "memory" && (
          <>
            {/* Search Section */}
            <Card className="mb-6">
              <CardContent className="pt-6">
                <div className="flex flex-col gap-4">
                  <div className="flex gap-2">
                    <div className="flex-1">
                      <SearchInput
                        value={searchQuery}
                        onChange={setSearchQuery}
                        placeholder="Busque uma memória para criar conexões manuais..."
                      />
                    </div>
                    <Button className="bg-gradient-to-r from-brain-primary to-brain-accent hover:from-brain-primary-dark hover:to-brain-accent-dark text-white">
                      <Search className="h-4 w-4 mr-2" />
                      Buscar
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Search Results */}
            {shouldSearch && (
              <Card className="mb-6">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Search className="h-4 w-4" />
                    Selecione a memória de origem
                  </CardTitle>
                  <p className="text-sm text-muted-foreground">
                    Clique em uma memória para conectá-la a outra
                  </p>
                </CardHeader>
                <CardContent>
                  {isSearching ? (
                    <div className="flex justify-center py-4">
                      <Spinner />
                    </div>
                  ) : searchResults.length > 0 ? (
                    <div className="space-y-2 max-h-[300px] overflow-y-auto">
                      {searchResults.map((memory) => (
                        <div
                          key={memory.id}
                          className={`flex items-center justify-between p-3 border rounded-lg cursor-pointer transition-colors
                            ${highlightedMemory?.id === memory.id
                              ? 'bg-brain-primary/10 border-brain-primary'
                              : 'hover:bg-accent'}`}
                          onClick={() => {
                            setHighlightedMemory(memory);
                            fetchConnectionsForMemory(memory.id);
                          }}
                        >
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium truncate">{memory.summary}</p>
                            <div className="flex items-center gap-2 mt-1">
                              <CategoryTag category={memory.category} />
                              <span className="text-xs text-muted-foreground">
                                {memory.tags?.slice(0, 2).join(", ") || ""}
                              </span>
                            </div>
                          </div>
                          {highlightedMemory?.id === memory.id && (
                            <div className="flex items-center gap-1 text-brain-primary">
                              <span className="text-xs font-medium">Selecionado</span>
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-center text-muted-foreground py-4">
                      Nenhuma memória encontrada para "{debouncedSearchQuery}"
                    </p>
                  )}
                </CardContent>
              </Card>
            )}

            {/* Selected Memory with Connections */}
            {highlightedMemory && (
              <Card className="mb-6 border-brain-primary/30">
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    <span className="flex items-center gap-2">
                      <Network className="h-5 w-5 text-brain-primary" />
                      Memória Selecionada
                    </span>
                    <Button
                      className="bg-gradient-to-r from-brain-primary to-brain-accent hover:from-brain-primary-dark hover:to-brain-accent-dark text-white"
                      onClick={() => {
                        setSelectedMemory(highlightedMemory);
                        setShowCreateDialog(true);
                      }}
                    >
                      <Link2 className="h-4 w-4 mr-2" />
                      Conectar
                    </Button>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {/* Selected memory info */}
                  <div className="p-4 border rounded-lg bg-accent/50 mb-4">
                    <p className="font-medium">{highlightedMemory.summary}</p>
                    <p className="text-sm text-muted-foreground mt-2 line-clamp-2">
                      {highlightedMemory.content}
                    </p>
                    <div className="flex items-center gap-2 mt-3">
                      <CategoryTag category={highlightedMemory.category} />
                      <span className="text-xs text-muted-foreground">
                        {highlightedMemory.tags?.join(", ") || "-"}
                      </span>
                    </div>
                  </div>

                  {/* Existing connections */}
                  <h4 className="text-sm font-medium mb-2">Conexões existentes:</h4>
                  {isLoadingConnections ? (
                    <div className="flex justify-center py-4">
                      <Spinner />
                    </div>
                  ) : highlightedConnections.length > 0 ? (
                    <div className="space-y-2 max-h-[200px] overflow-y-auto">
                      {highlightedConnections.map((conn) => {
                        const isSource = conn.fromMemoryId === highlightedMemory.id;
                        const otherMemory = isSource ? conn.toMemory : conn.fromMemory;
                        const typeLabels: Record<string, string> = {
                          RELATED: "Relacionado",
                          DEPENDS_ON: "Depende de",
                          DEPENDENT_OF: "É dependência de",
                          CONTRADICTS: "Contradiz",
                          SIMILAR_TO: "Similar a",
                          EXTENDS: "Estende",
                        };
                        return (
                          <div
                            key={conn.id}
                            className="flex items-center gap-2 p-2 border rounded hover:bg-accent transition-colors"
                          >
                            <ArrowRight className="h-3 w-3 text-muted-foreground" />
                            <span className="text-sm flex-1 truncate">
                              {otherMemory?.summary || (isSource ? conn.toMemoryId : conn.fromMemoryId)}
                            </span>
                            <span className="px-2 py-0.5 rounded-full bg-brain-primary/10 text-brain-primary text-xs">
                              {typeLabels[conn.type] || conn.type}
                            </span>
                          </div>
                        );
                      })}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground py-2">
                      Nenhuma conexão encontrada para esta memória.
                    </p>
                  )}
                </CardContent>
              </Card>
            )}

            {/* Memory Relationships List */}
            <Card>
              <CardHeader>
                <CardTitle>Memórias Conectadas</CardTitle>
                <p className="text-sm text-muted-foreground">
                  Conexões manuais entre memórias (depende de, similar a, contradiz, etc.)
                </p>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <div className="space-y-4">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div key={i} className="flex items-center gap-4 p-4 border rounded-lg">
                        <Skeleton variant="circular" width={40} height={40} />
                        <Skeleton variant="text" width="40%" />
                        <Skeleton variant="text" width="30%" />
                      </div>
                    ))}
                  </div>
                ) : relationships.length === 0 ? (
                  <div className="text-center py-12 text-muted-foreground">
                    <Network className="h-16 w-16 mx-auto mb-4 opacity-50" />
                    <h3 className="text-lg font-semibold mb-2">
                      Nenhuma conexão manual
                    </h3>
                    <p className="mb-4 max-w-md mx-auto">
                      Use a busca acima para conectar memórias manualmente.
                      <br />
                      <span className="text-sm">
                        Exemplo: "Padrão JWT" → depende de → "Config Spring Security"
                      </span>
                    </p>
                  </div>
                ) : (
                  <div className="space-y-4 max-h-[500px] overflow-y-auto">
                    {relationships.map((rel) => (
                      <RelationshipItem
                        key={rel.id}
                        relationship={rel}
                        onDelete={handleDeleteRelationship}
                      />
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </>
        )}
      </main>

      {/* Create Relationship Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="sm:max-w-[700px] max-h-[85vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Link2 className="h-5 w-5 text-brain-primary" />
              Criar Conexão entre Memórias
            </DialogTitle>
          </DialogHeader>

          <div className="flex-1 overflow-y-auto space-y-6 py-4">
            {/* Source Memory */}
            {selectedMemory && (
              <div>
                <label className="text-sm font-medium text-muted-foreground mb-2 block">
                  Memória de Origem
                </label>
                <div className="p-4 border rounded-lg bg-brain-primary/5 border-brain-primary/30">
                  <p className="font-medium">{selectedMemory.summary}</p>
                  <div className="flex items-center gap-2 mt-2">
                    <CategoryTag category={selectedMemory.category} />
                    <span className="text-xs text-muted-foreground">
                      {selectedMemory.tags?.join(", ") || "-"}
                    </span>
                  </div>
                </div>
              </div>
            )}

            {/* Relationship Type */}
            <div>
              <label className="text-sm font-medium text-muted-foreground mb-2 block">
                Tipo de Relacionamento
              </label>
              <select
                className="w-full h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={relationshipType}
                onChange={(e) => setRelationshipType(e.target.value)}
              >
                {RELATIONSHIP_TYPES.map((type) => (
                  <option key={type.value} value={type.value}>
                    {type.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Search for Target Memory */}
            <div>
              <label className="text-sm font-medium text-muted-foreground mb-2 block">
                Buscar Memória de Destino
              </label>
              <SearchInput
                value={dialogSearchQuery}
                onChange={setDialogSearchQuery}
                placeholder="Digite para buscar memórias..."
              />

              {/* Search Results */}
              {isDialogSearching ? (
                <div className="flex justify-center py-4">
                  <Spinner />
                </div>
              ) : dialogSearchResults.length > 0 ? (
                <div className="mt-3 space-y-2 max-h-[200px] overflow-y-auto border rounded-lg p-2">
                  {dialogSearchResults.map((memory) => (
                    <div
                      key={memory.id}
                      className={`p-3 border rounded-lg cursor-pointer transition-colors
                        ${targetMemory?.id === memory.id
                          ? 'bg-brain-accent/10 border-brain-accent'
                          : 'hover:bg-accent'}`}
                      onClick={() => setTargetMemory(memory)}
                    >
                      <p className="text-sm font-medium">{memory.summary}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <CategoryTag category={memory.category} />
                        <span className="text-xs text-muted-foreground">
                          {memory.tags?.slice(0, 2).join(", ") || "-"}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              ) : dialogSearchQuery.length >= 2 ? (
                <p className="text-sm text-muted-foreground mt-3 text-center py-4">
                  Nenhuma memória encontrada para "{dialogSearchQuery}"
                </p>
              ) : null}
            </div>

            {/* Selected Target Memory */}
            {targetMemory && (
              <div>
                <label className="text-sm font-medium text-muted-foreground mb-2 block">
                  Memória de Destino Selecionada
                </label>
                <div className="p-4 border rounded-lg bg-brain-accent/5 border-brain-accent/30">
                  <p className="font-medium">{targetMemory.summary}</p>
                  <div className="flex items-center gap-2 mt-2">
                    <CategoryTag category={targetMemory.category} />
                    <span className="text-xs text-muted-foreground">
                      {targetMemory.tags?.join(", ") || "-"}
                    </span>
                  </div>
                </div>
              </div>
            )}

            {/* Connection Preview */}
            {selectedMemory && targetMemory && (
              <div className="p-4 border-2 border-dashed rounded-lg bg-accent/30">
                <p className="text-sm text-center text-muted-foreground mb-2">Prévia da conexão:</p>
                <div className="flex items-center justify-center gap-3 flex-wrap">
                  <span className="font-medium text-sm bg-brain-primary/10 px-3 py-1 rounded">
                    {selectedMemory.summary.substring(0, 30)}...
                  </span>
                  <span className="px-3 py-1 rounded-full bg-brain-primary text-white text-xs font-medium">
                    {RELATIONSHIP_TYPES.find(t => t.value === relationshipType)?.label || relationshipType}
                  </span>
                  <ArrowRight className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium text-sm bg-brain-accent/10 px-3 py-1 rounded">
                    {targetMemory.summary.substring(0, 30)}...
                  </span>
                </div>
              </div>
            )}
          </div>

          <DialogFooter className="border-t pt-4">
            <Button
              variant="outline"
              onClick={() => setShowCreateDialog(false)}
            >
              Cancelar
            </Button>
            <Button
              className="bg-gradient-to-r from-brain-primary to-brain-accent hover:from-brain-primary-dark hover:to-brain-accent-dark text-white"
              onClick={handleCreateRelationship}
              disabled={!selectedMemory || !targetMemory}
            >
              <Link2 className="h-4 w-4 mr-2" />
              Criar Conexão
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

interface RelationshipItemProps {
  relationship: Relationship;
  onDelete: (id: string) => void;
}

function RelationshipItem({ relationship, onDelete }: RelationshipItemProps) {
  const typeLabels: Record<string, string> = {
    RELATED: "Relacionado",
    DEPENDS_ON: "Depende de",
    DEPENDENT_OF: "É dependência de",
    CONTRADICTS: "Contradiz",
    SIMILAR_TO: "Similar a",
    EXTENDS: "Estende",
  };

  return (
    <div className="flex items-center justify-between p-4 border rounded-lg hover:bg-accent transition-colors">
      <div className="flex items-center gap-4">
        <div className="p-2 bg-primary/10 rounded-full">
          <Network className="h-4 w-4 text-primary" />
        </div>

        <div className="flex-1">
          {relationship.fromMemory && relationship.toMemory ? (
            <>
              <div className="text-sm">
                <span className="font-medium">
                  {relationship.fromMemory.summary}
                </span>
              </div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <ArrowRight className="h-3 w-3" />
                <span>{typeLabels[relationship.type] || relationship.type}</span>
                <ArrowRight className="h-3 w-3" />
                <span className="font-medium text-foreground">
                  {relationship.toMemory.summary}
                </span>
              </div>
            </>
          ) : (
            <div className="text-sm text-muted-foreground">
              Relacionamento ID: {relationship.id}
            </div>
          )}

          <div className="flex items-center gap-4 text-xs text-muted-foreground">
            <span>Força: {Math.round((relationship.strength || 0) * 100)}%</span>
          </div>
        </div>

        <Button
          variant="ghost"
          size="icon"
          onClick={() => onDelete(relationship.id)}
        >
          <Trash2 className="h-4 w-4 text-destructive" />
        </Button>
      </div>
    </div>
  );
}
