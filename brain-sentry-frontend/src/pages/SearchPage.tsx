import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Search, Sparkles, Brain, ChevronDown, ChevronUp, Info } from "lucide-react";
import { useDebounce } from "@/hooks";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Input, FilterSelect } from "@/components/ui/filter";
import { Spinner, Skeleton } from "@/components/ui/spinner";
import { useToast } from "@/components/ui/toast";
import { MemoryCard } from "@/components/memory";
import { useAuth } from "@/contexts/AuthContext";
import { api } from "@/lib/api/client";

interface SearchResultItem {
  id: string;
  content: string;
  summary: string;
  category: string;
  importance: string;
  score: number;
  tags: string[];
  createdAt: string;
  updatedAt?: string;
  accessCount?: number;
  injectionCount?: number;
  helpfulCount?: number;
  scoreTrace?: {
    vectorScore?: number;
    graphScore?: number;
    recencyScore?: number;
    importanceBoost?: number;
    totalScore?: number;
    explanation?: string;
  };
}

interface PlanSearchResult {
  query: string;
  rounds: Array<{
    round: number;
    subQuery: string;
    results: SearchResultItem[];
    coverage: number;
  }>;
  finalResults: SearchResultItem[];
  totalCoverage: number;
  searchTimeMs: number;
}

const CATEGORY_KEYS = ["", "INSIGHT", "WARNING", "KNOWLEDGE", "ACTION", "CONTEXT", "REFERENCE", "GENERAL"] as const;
const IMPORTANCE_KEYS = ["", "CRITICAL", "IMPORTANT", "MINOR"] as const;

export function SearchPage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const { toast } = useToast();

  const CATEGORY_OPTIONS = CATEGORY_KEYS.map((k) => ({
    value: k,
    label: k === "" ? t("search.categoryOptions.all") : t(`search.categoryOptions.${k}`),
  }));
  const IMPORTANCE_OPTIONS = IMPORTANCE_KEYS.map((k) => ({
    value: k,
    label: k === "" ? t("search.importanceOptions.all") : t(`search.importanceOptions.${k}`),
  }));

  // Search state
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedQuery = useDebounce(searchQuery, 500);
  const [category, setCategory] = useState("");
  const [importance, setImportance] = useState("");
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(12);

  // Results state
  const [searchResults, setSearchResults] = useState<SearchResultItem[]>([]);
  const [totalResults, setTotalResults] = useState(0);
  const [searchTimeMs, setSearchTimeMs] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  // Advanced search (retrieval planner)
  const [advancedMode, setAdvancedMode] = useState(false);
  const [planResult, setPlanResult] = useState<PlanSearchResult | null>(null);
  const [planLoading, setPlanLoading] = useState(false);

  // ScoreTrace expanded
  const [expandedScores, setExpandedScores] = useState<Set<string>>(new Set());

  const shouldSearch = debouncedQuery.length >= 2 || category || importance;

  // Server-side search
  const performSearch = useCallback(async () => {
    if (!shouldSearch) {
      setSearchResults([]);
      setTotalResults(0);
      setError(null);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await api.axiosInstance.post<SearchResultItem[] | { results: SearchResultItem[]; total: number; searchTimeMs: number }>(
        "/v1/memories/search",
        {
          query: debouncedQuery || "*",
          category: category || undefined,
          importance: importance || undefined,
          limit: pageSize,
          offset: page * pageSize,
        }
      );

      // Handle both array and object response formats
      if (Array.isArray(response.data)) {
        setSearchResults(response.data);
        setTotalResults(response.data.length);
      } else {
        setSearchResults(response.data.results || []);
        setTotalResults(response.data.total || response.data.results?.length || 0);
        setSearchTimeMs(response.data.searchTimeMs || 0);
      }
    } catch (err) {
      setError(err as Error);
      toast({ title: t("search.searchError"), description: (err as Error).message, variant: "error" });
    } finally {
      setIsLoading(false);
    }
  }, [shouldSearch, debouncedQuery, category, importance, pageSize, page, toast, t]);

  // Auto-search when params change
  useEffect(() => {
    if (!advancedMode) performSearch();
  }, [performSearch, advancedMode]);

  // Reset page on filter change
  useEffect(() => { setPage(0); }, [debouncedQuery, category, importance]);

  const totalPages = Math.ceil(totalResults / pageSize);

  // Advanced search with retrieval planner
  const handlePlanSearch = async () => {
    if (!searchQuery.trim()) return;
    setPlanLoading(true);
    setPlanResult(null);
    try {
      const data = await api.planSearch(searchQuery, pageSize);
      setPlanResult(data);
      if (data?.finalResults) {
        setSearchResults(data.finalResults);
        setTotalResults(data.finalResults.length);
      }
    } catch (err) {
      toast({ title: t("search.advancedError"), description: (err as Error).message, variant: "error" });
    } finally {
      setPlanLoading(false);
    }
  };

  const handleSearch = () => {
    if (searchQuery.length < 2 && !category && !importance) {
      toast({ title: t("search.shortSearch"), description: t("search.shortSearchDesc"), variant: "warning" });
      return;
    }
    if (advancedMode) {
      handlePlanSearch();
    } else {
      performSearch();
    }
  };

  const handleClearFilters = () => {
    setSearchQuery("");
    setCategory("");
    setImportance("");
    setPage(0);
    setPlanResult(null);
  };

  const toggleScoreTrace = (id: string) => {
    setExpandedScores(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  };

  const results = searchResults;

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Sparkles className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("search.semanticTitle")}</h1>
                <p className="text-xs text-white/80">{t("search.subtitle")}</p>
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
                <div className="flex-1 relative">
                  <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input type="text" value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder={t("search.semanticPlaceholder")}
                    className="pl-10 h-12"
                    onKeyDown={(e) => { if (e.key === "Enter") handleSearch(); }} />
                </div>
                <Button size="lg" className="bg-gradient-to-r from-brain-primary to-brain-accent text-white" onClick={handleSearch}>
                  <Search className="h-5 w-5 mr-2" /> {t("search.runSearch")}
                </Button>
              </div>

              {/* Filters + Advanced Toggle */}
              <div className="flex flex-wrap gap-4 items-end">
                <FilterSelect label={t("search.categoryFilter")} options={CATEGORY_OPTIONS} value={category} onChange={setCategory} />
                <FilterSelect label={t("search.importanceFilter")} options={IMPORTANCE_OPTIONS} value={importance} onChange={setImportance} />

                <Button
                  variant={advancedMode ? "default" : "outline"}
                  size="sm"
                  onClick={() => setAdvancedMode(!advancedMode)}
                  className={advancedMode ? "bg-brain-accent hover:bg-brain-accent/90" : ""}
                >
                  <Brain className="h-4 w-4 mr-1" />
                  {t("search.advanced")}
                </Button>

                {(searchQuery || category || importance) && (
                  <Button variant="ghost" size="sm" onClick={handleClearFilters}>{t("search.clearFilters")}</Button>
                )}
              </div>

              {advancedMode && (
                <div className="p-3 bg-brain-accent/10 rounded-md text-sm text-muted-foreground flex items-center gap-2">
                  <Info className="h-4 w-4 shrink-0" />
                  <span>
                    {t("search.advancedInfo")}
                  </span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Plan Search Rounds */}
        {planResult && planResult.rounds && (
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="text-sm flex items-center gap-2">
                <Brain className="h-4 w-4 text-brain-accent" />
                {t("search.planTitle", { rounds: planResult.rounds.length, coverage: ((planResult.totalCoverage || 0) * 100).toFixed(0) })}
                {planResult.searchTimeMs > 0 && <span className="text-xs text-muted-foreground">({planResult.searchTimeMs}ms)</span>}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {planResult.rounds.map((round) => (
                  <div key={round.round} className="flex items-center gap-3 p-2 bg-accent rounded-md">
                    <span className="text-xs font-bold text-brain-primary bg-brain-primary/10 px-2 py-1 rounded">
                      {t("search.round", { n: round.round })}
                    </span>
                    <span className="text-sm flex-1">"{round.subQuery}"</span>
                    <span className="text-xs text-muted-foreground">{t("search.resultsCount", { count: round.results?.length || 0 })}</span>
                    <div className="w-16 h-2 bg-gray-200 rounded-full overflow-hidden">
                      <div className="h-full bg-brain-accent rounded-full" style={{ width: `${(round.coverage || 0) * 100}%` }} />
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Loading State */}
        {(isLoading || planLoading) && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
            {Array.from({ length: 6 }).map((_, i) => (
              <Card key={i}>
                <CardContent className="p-6">
                  <Skeleton variant="text" width="60%" className="mb-4" />
                  <Skeleton variant="rectangular" height={80} className="mb-4" />
                  <div className="flex gap-2"><Skeleton variant="text" width={60} /><Skeleton variant="text" width={60} /></div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Error */}
        {error && (
          <Card className="mb-6">
            <CardContent className="p-12 text-center">
              <p className="text-muted-foreground">{t("search.errorMessage", { message: error.message })}</p>
              <Button className="mt-4" onClick={handleSearch}>{t("search.tryAgain")}</Button>
            </CardContent>
          </Card>
        )}

        {/* No Results */}
        {!isLoading && !planLoading && shouldSearch && results.length === 0 && (
          <Card className="mb-6">
            <CardContent className="p-12 text-center">
              <Search className="h-16 w-16 mx-auto mb-4 text-muted-foreground opacity-50" />
              <h3 className="text-lg font-semibold mb-2">{t("search.noResults")}</h3>
              <p className="text-muted-foreground mb-4">{t("search.noResultsDesc")}</p>
              <Button variant="outline" onClick={handleClearFilters}>{t("search.clearFilters")}</Button>
            </CardContent>
          </Card>
        )}

        {/* Results */}
        {!isLoading && !planLoading && results.length > 0 && (
          <>
            <div className="mb-4 flex items-center justify-between">
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span>{t("search.resultsFound", { total: totalResults })}</span>
                {searchTimeMs > 0 && <span>{t("search.in", { time: searchTimeMs })}</span>}
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground">{t("search.perPage")}</span>
                <select value={pageSize} onChange={(e) => { setPageSize(Number(e.target.value)); setPage(0); }}
                  className="h-8 rounded-md border border-input bg-background px-2 text-xs">
                  {[6, 12, 24, 48].map(s => <option key={s} value={s}>{s}</option>)}
                </select>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
              {results.map((memory) => (
                <div key={memory.id}>
                  <MemoryCard memory={memory} />
                  {/* ScoreTrace */}
                  {(memory.score !== undefined || memory.scoreTrace) && (
                    <div className="mt-1 px-3">
                      <button
                        onClick={() => toggleScoreTrace(memory.id)}
                        className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
                      >
                        {expandedScores.has(memory.id) ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                        Score: {(memory.score || memory.scoreTrace?.totalScore || 0).toFixed(3)}
                      </button>
                      {expandedScores.has(memory.id) && (
                        <div className="mt-1 p-2 bg-accent rounded-md text-xs space-y-1">
                          {memory.scoreTrace ? (
                            <>
                              {memory.scoreTrace.vectorScore !== undefined && (
                                <div className="flex justify-between">
                                  <span>Vector:</span>
                                  <span className="font-mono">{memory.scoreTrace.vectorScore.toFixed(4)}</span>
                                </div>
                              )}
                              {memory.scoreTrace.graphScore !== undefined && (
                                <div className="flex justify-between">
                                  <span>Graph:</span>
                                  <span className="font-mono">{memory.scoreTrace.graphScore.toFixed(4)}</span>
                                </div>
                              )}
                              {memory.scoreTrace.recencyScore !== undefined && (
                                <div className="flex justify-between">
                                  <span>Recency:</span>
                                  <span className="font-mono">{memory.scoreTrace.recencyScore.toFixed(4)}</span>
                                </div>
                              )}
                              {memory.scoreTrace.importanceBoost !== undefined && (
                                <div className="flex justify-between">
                                  <span>Importance:</span>
                                  <span className="font-mono">{memory.scoreTrace.importanceBoost.toFixed(4)}</span>
                                </div>
                              )}
                              {memory.scoreTrace.explanation && (
                                <p className="text-muted-foreground mt-1 border-t pt-1">{memory.scoreTrace.explanation}</p>
                              )}
                            </>
                          ) : (
                            <div className="flex justify-between">
                              <span>Total Score:</span>
                              <span className="font-mono">{(memory.score || 0).toFixed(4)}</span>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* Server-side Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-4">
                <Button size="sm" variant="outline" disabled={page === 0}
                  onClick={() => setPage(p => Math.max(0, p - 1))}>
                  {t("search.pagePrev")}
                </Button>
                <span className="text-sm text-muted-foreground">
                  {t("memory.page", { current: page + 1, total: totalPages })}
                </span>
                <Button size="sm" variant="outline" disabled={page >= totalPages - 1}
                  onClick={() => setPage(p => p + 1)}>
                  {t("search.pageNext")}
                </Button>
              </div>
            )}
          </>
        )}

        {/* Initial State */}
        {!shouldSearch && !planResult && (
          <Card>
            <CardContent className="p-12 text-center">
              <div className="max-w-md mx-auto">
                <div className="p-4 bg-gradient-to-br from-brain-primary/20 to-brain-accent/20 rounded-full w-16 h-16 mx-auto mb-4">
                  <Search className="h-8 w-8 text-brain-primary mx-auto mt-4" />
                </div>
                <h3 className="text-lg font-semibold mb-2">{t("search.searchInMemories")}</h3>
                <p className="text-muted-foreground mb-6">
                  {t("search.searchTip")}
                </p>
                <div className="flex flex-wrap gap-2 justify-center">
                  {["Spring Boot configuration", "React hooks patterns", "API REST best practices", "Docker optimization"].map((s) => (
                    <Button key={s} variant="outline" size="sm" onClick={() => setSearchQuery(s)}>{s}</Button>
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
