import { useState, useCallback } from "react";
import { Search, SlidersHorizontal, Sparkles, Clock } from "lucide-react";
import { useFetch, useDebounce } from "@/hooks";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Input, SearchInput, FilterBar, AdvancedFilters } from "@/components/ui/filter";
import { Spinner, Skeleton } from "@/components/ui/spinner";
import { Pagination } from "@/components/ui/pagination";
import { useToast } from "@/components/ui/toast";
import { MemoryCard } from "@/components/memory";
import { useAuth } from "@/contexts/AuthContext";

interface SearchResponse {
  query: string;
  results: Array<{
    id: string;
    content: string;
    summary: string;
    category: string;
    importance: string;
    score: number;
    createdAt: string;
    tags: string[];
  }>;
  totalResults: number;
  searchTimeMs: number;
}

interface CategoryOption {
  value: string;
  label: string;
}

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

const CATEGORY_OPTIONS: CategoryOption[] = [
  { value: "", label: "Todas" },
  { value: "DECISION", label: "Decisões" },
  { value: "PATTERN", label: "Padrões" },
  { value: "ANTIPATTERN", label: "Anti-padrões" },
  { value: "DOMAIN", label: "Domínio" },
  { value: "BUG", label: "Bugs" },
  { value: "OPTIMIZATION", label: "Otimizações" },
  { value: "INTEGRATION", label: "Integrações" },
];

const IMPORTANCE_OPTIONS: CategoryOption[] = [
  { value: "", label: "Todas" },
  { value: "CRITICAL", label: "Crítico" },
  { value: "IMPORTANT", label: "Importante" },
  { value: "MINOR", label: "Menor" },
];

export function SearchPage() {
  const { user } = useAuth();
  const { toast } = useToast();
  const tenantId = user?.tenantId || "default";

  // Search state
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedQuery = useDebounce(searchQuery, 500);
  const [category, setCategory] = useState("");
  const [importance, setImportance] = useState("");
  const [showFilters, setShowFilters] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(12);

  // Build search query
  const shouldSearch = debouncedQuery.length >= 2 || category || importance;

  // Fetch search results
  const {
    data: searchResults,
    isLoading,
    error,
    refetch,
  } = useFetch<SearchResponse>(
    shouldSearch
      ? `${API_URL}/v1/memories/search`
      : null,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        query: debouncedQuery || "*",
        category: category || undefined,
        importance: importance || undefined,
        limit: pageSize,
        offset: (page - 1) * pageSize,
      }),
      skip: !shouldSearch,
    }
  );

  const results = searchResults?.results || [];
  const totalResults = searchResults?.totalResults || 0;
  const totalPages = Math.ceil(totalResults / pageSize);

  const handleSearch = () => {
    if (searchQuery.length < 2 && !category && !importance) {
      toast({
        title: "Busca muito curta",
        description: "Digite pelo menos 2 caracteres ou selecione um filtro.",
        variant: "warning",
      });
      return;
    }
    refetch?.();
  };

  const handleClearFilters = () => {
    setSearchQuery("");
    setCategory("");
    setImportance("");
    setPage(1);
  };

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Sparkles className="h-6 w-6 text-primary" />
              </div>
              <div>
                <h1 className="text-2xl font-bold">Busca Semântica</h1>
                <p className="text-sm text-muted-foreground">
                  Encontre memórias usando IA
                </p>
              </div>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Search Section */}
        <Card className="mb-6">
          <CardContent className="pt-6">
            <div className="flex flex-col gap-4">
              {/* Main Search Bar */}
              <div className="flex gap-2">
                <div className="flex-1">
                  <SearchInput
                    value={searchQuery}
                    onChange={setSearchQuery}
                    placeholder="Digite sua dúvida ou contexto técnico..."
                    className="h-12"
                  />
                </div>
                <Button size="lg" onClick={handleSearch}>
                  <Search className="h-5 w-5" />
                  Buscar
                </Button>
              </div>

              {/* Filters */}
              <FilterBar
                searchValue={searchQuery}
                onSearchChange={setSearchQuery}
                filters={[
                  {
                    key: "category",
                    label: "Categoria",
                    options: CATEGORY_OPTIONS,
                    value: category,
                    onChange: setCategory,
                  },
                  {
                    key: "importance",
                    label: "Importância",
                    options: IMPORTANCE_OPTIONS,
                    value: importance,
                    onChange: setImportance,
                  },
                ]}
                onClearFilters={handleClearFilters}
              />
            </div>
          </CardContent>
        </Card>

        {/* Loading State */}
        {isLoading && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
            {Array.from({ length: 6 }).map((_, i) => (
              <Card key={i}>
                <CardContent className="p-6">
                  <Skeleton variant="text" width="60%" className="mb-4" />
                  <Skeleton variant="rectangular" height={80} className="mb-4" />
                  <div className="flex gap-2">
                    <Skeleton variant="text" width={60} />
                    <Skeleton variant="text" width={60} />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Error State */}
        {error && (
          <Card className="mb-6">
            <CardContent className="p-12 text-center">
              <p className="text-muted-foreground">
                Erro na busca: {(error as Error).message}
              </p>
              <Button className="mt-4" onClick={handleSearch}>
                Tentar novamente
              </Button>
            </CardContent>
          </Card>
        )}

        {/* No Results */}
        {!isLoading && shouldSearch && results.length === 0 && (
          <Card className="mb-6">
            <CardContent className="p-12 text-center">
              <Search className="h-16 w-16 mx-auto mb-4 text-muted-foreground opacity-50" />
              <h3 className="text-lg font-semibold mb-2">
                Nenhum resultado encontrado
              </h3>
              <p className="text-muted-foreground mb-4">
                Tente ajustar sua busca ou filtros
              </p>
              <Button variant="outline" onClick={handleClearFilters}>
                Limpar filtros
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Results */}
        {!isLoading && results.length > 0 && (
          <>
            <div className="mb-4 flex items-center justify-between">
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span>{totalResults} resultados encontrados</span>
                {searchResults?.searchTimeMs && (
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {searchResults.searchTimeMs}ms
                  </span>
                )}
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
              {results.map((memory) => (
                <MemoryCard key={memory.id} memory={memory} />
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <Pagination
                currentPage={page}
                totalPages={totalPages}
                onPageChange={handlePageChange}
                pageSize={pageSize}
                totalItems={totalResults}
                showPageSizeSelector
                pageSizeOptions={[6, 12, 24, 48]}
                onPageSizeChange={(size) => {
                  setPageSize(size);
                  setPage(1);
                }}
              />
            )}
          </>
        )}

        {/* Initial State */}
        {!shouldSearch && (
          <Card>
            <CardContent className="p-12 text-center">
              <div className="max-w-md mx-auto">
                <div className="p-4 bg-primary/10 rounded-full w-16 h-16 mx-auto mb-4">
                  <Search className="h-8 w-8 text-primary mx-auto mt-4" />
                </div>
                <h3 className="text-lg font-semibold mb-2">
                  Busque em suas memórias
                </h3>
                <p className="text-muted-foreground mb-6">
                  Digite uma pergunta, descrição ou contexto técnico para
                  encontrar memórias relevantes usando busca semântica.
                </p>
                <div className="flex flex-wrap gap-2 justify-center">
                  {[
                    "Spring Boot configuration",
                    "React hooks patterns",
                    "API REST best practices",
                    "Docker optimization",
                  ].map((suggestion) => (
                    <Button
                      key={suggestion}
                      variant="outline"
                      size="sm"
                      onClick={() => setSearchQuery(suggestion)}
                    >
                      {suggestion}
                    </Button>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </main>
    </div>
  );
}
