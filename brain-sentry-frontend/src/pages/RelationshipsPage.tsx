import { useState } from "react";
import {
  Network,
  Plus,
  Search,
  Filter,
  ArrowRight,
  Trash2,
  RefreshCw,
  AlertCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Input, SearchInput } from "@/components/ui/filter";
import { Spinner, Skeleton } from "@/components/ui/spinner";
import { CategoryTag, ReadOnlyTags } from "@/components/ui/tags";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { useFetch } from "@/hooks";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";

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

const RELATIONSHIP_TYPES = [
  { value: "RELATED", label: "Relacionado" },
  { value: "DEPENDS_ON", label: "Depende de" },
  { value: "DEPENDENT_OF", label: "Dependente" },
  { value: "CONTRADICTS", label: "Contradiz" },
  { value: "SIMILAR_TO", label: "Similar a" },
  { value: "EXTENDS", label: "Estende" },
];

export function RelationshipsPage() {
  const { user } = useAuth();
  const { toast } = useToast();
  const tenantId = user?.tenantId || "default";

  // State
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedMemory, setSelectedMemory] = useState<Memory | null>(null);
  const [relationshipType, setRelationshipType] = useState("RELATED");
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);

  // Fetch relationships
  const {
    data: relationshipsData,
    isLoading,
    error,
    refetch,
  } = useFetch<{ relationships: Relationship[]; totalElements: number }>(
    `${API_URL}/v1/relationships?page=${page - 1}&size=${pageSize}`
  );

  // Search for memories to link
  const shouldSearch = searchQuery && searchQuery.length >= 2;
  const {
    data: searchResults,
    isLoading: isSearching,
  } = useFetch<{ memories: Memory[] }>(
    shouldSearch
      ? `${API_URL}/v1/memories/search`
      : `${API_URL}/v1/memories/search`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        query: searchQuery,
        limit: 10,
      }),
      skip: !shouldSearch,
    }
  );

  const relationships = relationshipsData?.relationships || [];
  const totalElements = relationshipsData?.totalElements || 0;

  // Create relationship
  const handleCreateRelationship = async (toMemoryId: string) => {
    if (!selectedMemory) return;

    try {
      const response = await fetch(`${API_URL}/v1/relationships`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Tenant-ID": tenantId,
        },
        body: JSON.stringify({
          fromMemoryId: selectedMemory.id,
          toMemoryId: toMemoryId,
          type: relationshipType,
          strength: 0.5,
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to create relationship");
      }

      toast({
        title: "Relacionamento criado",
        description: `Memória "${selectedMemory.summary}" foi conectada.`,
        variant: "success",
      });

      setShowCreateDialog(false);
      refetch?.();
      setSelectedMemory(null);
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

      refetch?.();
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
        <div className="px-6 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-white/20 rounded-lg backdrop-blur-sm">
                <Network className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-2xl font-bold">Relacionamentos</h1>
                <p className="text-sm text-white/80">
                  Visualize e gerencie conexões entre memórias
                </p>
              </div>
            </div>
            <Button variant="outline" size="sm" className="bg-white/20 border-white/30 text-white hover:bg-white/30" onClick={() => refetch?.()}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Action Bar */}
        <div className="mb-6 flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <SearchInput
              value={searchQuery}
              onChange={setSearchQuery}
              placeholder="Buscar memórias para conectar..."
            />
          </div>
          <Button className="bg-white text-brain-primary hover:bg-white/90" onClick={() => setShowCreateDialog(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Novo Relacionamento
          </Button>
        </div>

        {/* Search Results */}
        {isSearching && (
          <Card className="mb-6">
            <CardHeader>
              <CardTitle>Resultados da Busca</CardTitle>
            </CardHeader>
            <CardContent>
              {searchResults?.memories && searchResults.memories.length > 0 ? (
                <div className="space-y-2">
                  {searchResults.memories.map((memory) => (
                    <div
                      key={memory.id}
                      className="flex items-center justify-between p-3 border rounded-lg hover:bg-accent cursor-pointer"
                      onClick={() => {
                        setSelectedMemory(memory);
                        setShowCreateDialog(true);
                      }}
                    >
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium truncate">{memory.summary}</p>
                        <p className="text-xs text-muted-foreground truncate">{memory.category}</p>
                      </div>
                      <ArrowRight className="h-4 w-4 text-muted-foreground" />
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-center text-muted-foreground py-4">
                  Nenhuma memória encontrada
                </p>
              )}
            </CardContent>
          </Card>
        )}

        {/* Relationships List */}
        <Card>
          <CardHeader>
            <CardTitle>Relacionamentos</CardTitle>
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
                  Nenhum relacionamento encontrado
                </h3>
                <p className="mb-4">
                  Conecte memórias relacionadas para criar um grafo de conhecimento.
                </p>
                <Button className="bg-gradient-to-r from-brain-primary to-brain-accent hover:from-brain-primary-dark hover:to-brain-accent-dark text-white" onClick={() => setShowCreateDialog(true)}>
                  <Plus className="h-4 w-4 mr-2" />
                  Criar Relacionamento
                </Button>
              </div>
            ) : (
              <div className="space-y-4">
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
      </main>

      {/* Create Relationship Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {selectedMemory
                ? `Conectar "${selectedMemory.summary}"`
                : "Selecionar Memória"}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 px-6 py-4">
            {!selectedMemory ? (
              <p className="text-sm text-muted-foreground">
                Busque e selecione uma memória para conectar.
              </p>
            ) : (
              <>
                <div>
                  <label className="text-sm font-medium mb-2">Tipo de Relacionamento</label>
                  <select
                    className="w-full h-9 rounded-md border border-input bg-background px-3 py-1"
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

                {searchQuery && searchResults?.memories && (
                  <div>
                    <label className="text-sm font-medium mb-2">
                      Selecionar memória destino
                    </label>
                    <div className="space-y-2 max-h-48 overflow-y-auto">
                      {searchResults.memories.map((memory) => (
                        <button
                          key={memory.id}
                          onClick={() => handleCreateRelationship(memory.id)}
                          className="w-full text-left p-3 border rounded-lg hover:bg-accent text-left"
                        >
                          <div className="flex flex-col gap-1">
                            <p className="text-sm font-medium">{memory.summary}</p>
                            <div className="flex items-center gap-2">
                              <CategoryTag category={memory.category} />
                              <span className="text-xs text-muted-foreground">
                                {memory.tags?.slice(0, 2).join(", ") || "-"}
                              </span>
                            </div>
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
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
