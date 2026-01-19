import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Plus, Search, Loader2 } from "lucide-react";
import { MemoryCard, MemoryDialog } from "@/components/memory";
import { api, type Memory, getErrorMessage } from "@/lib/api";

export default function MemoryAdminPage() {
  const [memories, setMemories] = useState<Memory[]>([]);
  const [filteredMemories, setFilteredMemories] = useState<Memory[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedMemory, setSelectedMemory] = useState<Memory | undefined>(undefined);

  // Buscar memórias do backend
  const fetchMemories = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.getMemories(0, 100);
      setMemories(response.memories || []);
      setFilteredMemories(response.memories || []);
    } catch (err) {
      setError(getErrorMessage(err));
      console.error("Error fetching memories:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMemories();
  }, []);

  // Filtrar memórias localmente
  useEffect(() => {
    if (searchQuery === "") {
      setFilteredMemories(memories);
    } else {
      const filtered = memories.filter((memory) =>
        memory.content.toLowerCase().includes(searchQuery.toLowerCase()) ||
        memory.summary.toLowerCase().includes(searchQuery.toLowerCase()) ||
        memory.tags?.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()))
      );
      setFilteredMemories(filtered);
    }
  }, [searchQuery, memories]);

  const handleDelete = async (id: string) => {
    if (!confirm("Tem certeza que deseja excluir esta memória?")) return;

    try {
      await api.deleteMemory(id);
      setMemories(memories.filter(m => m.id !== id));
    } catch (err) {
      alert("Erro ao excluir: " + getErrorMessage(err));
    }
  };

  const handleView = (id: string) => {
    const memory = memories.find(m => m.id === id);
    if (memory) {
      setSelectedMemory(memory);
      setDialogOpen(true);
    }
  };

  const handleEdit = (id: string) => {
    const memory = memories.find(m => m.id === id);
    if (memory) {
      setSelectedMemory(memory);
      setDialogOpen(true);
    }
  };

  const handleCreate = () => {
    setSelectedMemory(undefined);
    setDialogOpen(true);
  };

  const handleDialogClose = () => {
    setDialogOpen(false);
    setSelectedMemory(undefined);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Memories</h1>
        <p className="text-sm text-muted-foreground">
          Gerencie as memórias do sistema
        </p>
      </div>

      {/* Actions Bar */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex-1 max-w-sm">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Buscar memórias..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-4 py-2 text-sm bg-background border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
        </div>
        <Button size="sm" onClick={handleCreate}>
          <Plus className="h-4 w-4 mr-2" />
          Nova Memória
        </Button>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          <span className="ml-2 text-muted-foreground">Carregando...</span>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-destructive/10 text-destructive p-4 rounded-md">
          <p className="font-medium">Erro ao carregar memórias</p>
          <p className="text-sm">{error}</p>
          <Button
            size="sm"
            variant="outline"
            className="mt-2"
            onClick={fetchMemories}
          >
            Tentar novamente
          </Button>
        </div>
      )}

      {/* Memories Grid */}
      {!loading && !error && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredMemories.map((memory) => (
              <MemoryCard
                key={memory.id}
                memory={memory}
                onView={handleView}
                onEdit={handleEdit}
                onDelete={handleDelete}
              />
            ))}
          </div>

          {filteredMemories.length === 0 && memories.length === 0 && (
            <div className="text-center py-12">
              <p className="text-muted-foreground">Nenhuma memória cadastrada.</p>
              <p className="text-sm text-muted-foreground mt-1">
                Clique em "Nova Memória" para criar a primeira.
              </p>
            </div>
          )}

          {filteredMemories.length === 0 && memories.length > 0 && (
            <div className="text-center py-12">
              <p className="text-muted-foreground">Nenhuma memória encontrada para "{searchQuery}"</p>
            </div>
          )}
        </>
      )}

      {/* Memory Dialog */}
      <MemoryDialog
        open={dialogOpen}
        onOpenChange={handleDialogClose}
        memory={selectedMemory}
        onSuccess={fetchMemories}
      />
    </div>
  );
}
