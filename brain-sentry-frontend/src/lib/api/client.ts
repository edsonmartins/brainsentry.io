import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from "axios";

// Configuração base da API
const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

// Tipos de resposta da API
export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface ApiError {
  message: string;
  statusCode?: number;
  details?: unknown;
}

// Tipos específicos do domínio - alinhados com o backend
// New universal categories
export type MemoryCategory =
  | "INSIGHT"      // Patterns, best practices, preferences
  | "DECISION"     // Decisions (technical or business)
  | "WARNING"      // Anti-patterns, bugs, objections
  | "KNOWLEDGE"    // Domain/customer/product knowledge
  | "ACTION"       // Actions, optimizations, follow-ups
  | "CONTEXT"      // Context, integrations, history
  | "REFERENCE"    // Documentation, materials
  // Legacy categories (deprecated, for backward compatibility)
  | "PATTERN" | "ANTIPATTERN" | "DOMAIN" | "BUG" | "OPTIMIZATION" | "INTEGRATION";
export type ImportanceLevel = "CRITICAL" | "IMPORTANT" | "MINOR";

export interface Memory {
  id: string;
  tenantId?: string;
  content: string;
  summary: string;
  category: MemoryCategory | string;
  importance: ImportanceLevel | string;
  validationStatus?: string;
  metadata?: Record<string, unknown>;
  tags: string[];
  sourceType?: string;
  sourceReference?: string;
  createdBy?: string;
  createdAt: string;
  updatedAt?: string;
  accessCount?: number;
  injectionCount?: number;
  helpfulCount?: number;
  embedding?: number[];
  memoryType?: string;
  emotionalWeight?: number;
  simHash?: string;
  validFrom?: string;
  validTo?: string;
  decayRate?: number;
  supersededBy?: string;
  decayedRelevance?: number;
}

export interface CreateMemoryRequest {
  content: string;
  summary: string;
  category?: MemoryCategory;
  importance?: ImportanceLevel;
  tags?: string[];
}

export interface UpdateMemoryRequest {
  content?: string;
  summary?: string;
  category?: MemoryCategory;
  importance?: ImportanceLevel;
  tags?: string[];
}

export interface MemoryListResponse {
  memories: Memory[];
  total: number;
  totalElements?: number;
  page: number;
  size: number;
  totalPages: number;
  hasNext?: boolean;
  hasPrevious?: boolean;
}

export interface SearchRequest {
  query: string;
  limit?: number;
}

export interface MemoryStats {
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

type SearchResponse = Memory[] | { results?: Memory[]; total?: number; searchTimeMs?: number };
type RawMemoryListResponse = Omit<MemoryListResponse, "total"> & { total?: number };

function normalizeMemoryListResponse(data: RawMemoryListResponse): MemoryListResponse {
  const total = data.total ?? data.totalElements ?? data.memories?.length ?? 0;
  return {
    ...data,
    total,
    totalElements: data.totalElements ?? total,
  };
}

// Interceptador para adicionar headers de autenticação
const authRequestInterceptor = (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
  // Adicionar tenant ID se disponível
  const tenantId = localStorage.getItem("tenant_id") || "a9f814d2-4dae-41f3-851b-8aa3d4706561";

  // Adicionar token JWT se disponível
  const token = localStorage.getItem("brain_sentry_token");

  if (config.headers) {
    config.headers["X-Tenant-ID"] = tenantId;
    if (token) {
      config.headers["Authorization"] = `Bearer ${token}`;
    }
  }
  return config;
};

// Interceptador para tratamento de erros
const errorResponseInterceptor = (error: unknown): Promise<ApiError> => {
  const apiError: ApiError = {
    message: "Erro desconhecido",
    statusCode: 0,
    details: error,
  };

  if (axios.isAxiosError(error)) {
    const responseData = error.response?.data as { message?: string } | undefined;
    apiError.message = responseData?.message || error.message || "Erro desconhecido";
    apiError.statusCode = error.response?.status;
    apiError.details = error.response?.data;
  } else if (error instanceof Error) {
    apiError.message = error.message;
  }

  return Promise.reject(apiError);
};

// Interceptador para tratamento de respostas de sucesso
const successResponseInterceptor = (response: AxiosResponse): AxiosResponse => {
  return response;
};

// Classe principal do API Client
class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        "Content-Type": "application/json",
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    this.client.interceptors.request.use(authRequestInterceptor);
    this.client.interceptors.response.use(successResponseInterceptor, errorResponseInterceptor);
  }

  // Memory endpoints - alinhados com MemoryController do backend
  async getMemories(page: number = 0, size: number = 20): Promise<MemoryListResponse> {
    const response = await this.client.get<MemoryListResponse>("/v1/memories", {
      params: { page, size },
    });
    return normalizeMemoryListResponse(response.data);
  }

  async getMemory(id: string): Promise<Memory> {
    const response = await this.client.get<Memory>(`/v1/memories/${id}`);
    return response.data;
  }

  async createMemory(data: CreateMemoryRequest): Promise<Memory> {
    const response = await this.client.post<Memory>("/v1/memories", data);
    return response.data;
  }

  async updateMemory(id: string, data: UpdateMemoryRequest): Promise<Memory> {
    const response = await this.client.put<Memory>(`/v1/memories/${id}`, data);
    return response.data;
  }

  async deleteMemory(id: string): Promise<void> {
    await this.client.delete(`/v1/memories/${id}`);
  }

  async searchMemories(query: string, limit: number = 10): Promise<Memory[]> {
    const response = await this.client.post<SearchResponse>("/v1/memories/search", {
      query,
      limit,
    });
    return Array.isArray(response.data) ? response.data : response.data.results || [];
  }

  async getMemoriesByCategory(category: string): Promise<Memory[]> {
    const response = await this.client.get<Memory[]>(`/v1/memories/by-category/${category}`);
    return response.data;
  }

  async getMemoriesByImportance(importance: string): Promise<Memory[]> {
    const response = await this.client.get<Memory[]>(`/v1/memories/by-importance/${importance}`);
    return response.data;
  }

  async getRelatedMemories(id: string, depth: number = 2): Promise<Memory[]> {
    const response = await this.client.get<Memory[]>(`/v1/memories/${id}/related`, {
      params: { depth },
    });
    return response.data;
  }

  async recordFeedback(id: string, helpful: boolean): Promise<void> {
    await this.client.post(`/v1/memories/${id}/feedback`, null, {
      params: { helpful },
    });
  }

  // Stats endpoints
  async getStats(): Promise<MemoryStats> {
    const response = await this.client.get<MemoryStats>("/v1/stats/overview");
    return response.data;
  }

  // Health check
  async healthCheck(): Promise<{ status: string; timestamp: string }> {
    const response = await this.client.get("/v1/stats/health");
    return response.data;
  }

  // Profile
  async getProfile(): Promise<any> {
    const response = await this.client.get("/v1/profile");
    return response.data;
  }

  // NL Graph Query
  async nlQuery(question: string): Promise<any> {
    const response = await this.client.post("/v1/graph/nl-query", { question });
    return response.data;
  }

  // Reflection
  async runReflection(): Promise<any> {
    const response = await this.client.post("/v1/reflect");
    return response.data;
  }

  // Reconciliation
  async reconcileFacts(content: string, sessionId?: string): Promise<any> {
    const response = await this.client.post("/v1/reconcile", { content, sessionId });
    return response.data;
  }

  // Retrieval Planner
  async planSearch(query: string, limit: number = 10): Promise<any> {
    const response = await this.client.post("/v1/memories/plan-search", { query, limit });
    return response.data;
  }

  // Spreading Activation
  async activateMemories(seedIds: string[], seedActivations?: number[]): Promise<any> {
    const response = await this.client.post("/v1/memories/activate", { seedIds, seedActivations });
    return response.data;
  }

  // Graph Communities
  async getCommunities(): Promise<any> {
    const response = await this.client.get("/v1/graph/communities");
    return response.data;
  }

  // Interception
  async intercept(prompt: string, sessionId?: string): Promise<any> {
    const response = await this.client.post("/v1/intercept", { prompt, sessionId });
    return response.data;
  }

  // Compression
  async compress(messages: any[], options?: any): Promise<any> {
    const response = await this.client.post("/v1/compression/compress", { messages, ...options });
    return response.data;
  }

  // Connectors
  async getConnectors(): Promise<any> {
    const response = await this.client.get("/v1/connectors");
    return response.data;
  }

  async syncConnector(name: string): Promise<any> {
    const response = await this.client.post(`/v1/connectors/${name}/sync`);
    return response.data;
  }

  // Tasks
  async getTaskMetrics(): Promise<any> {
    const response = await this.client.get("/v1/tasks/metrics");
    return response.data;
  }

  // Consolidation
  async consolidate(similarityThreshold: number = 0.85): Promise<any> {
    const response = await this.client.post("/v1/consolidate", { similarityThreshold });
    return response.data;
  }

  // Benchmark
  async runBenchmark(queryCount: number = 10, k: number = 10): Promise<any> {
    const response = await this.client.post("/v1/benchmark/run", { queryCount, k });
    return response.data;
  }

  // Admin
  async getCircuitBreakers(): Promise<any> {
    const response = await this.client.get("/v1/admin/circuit-breakers");
    return response.data;
  }

  async getLLMMetrics(): Promise<any> {
    const response = await this.client.get("/v1/admin/llm-metrics");
    return response.data;
  }

  async scanPII(text: string): Promise<any> {
    const response = await this.client.post("/v1/pii/scan", { text });
    return response.data;
  }

  // Memory Versions
  async getMemoryVersions(id: string): Promise<any> {
    const response = await this.client.get(`/v1/memories/${id}/versions`);
    return response.data;
  }

  // Memory Correction
  async flagMemory(id: string, reason: string): Promise<any> {
    const response = await this.client.post(`/v1/memories/${id}/flag`, { reason });
    return response.data;
  }

  async reviewCorrection(id: string, action: string): Promise<any> {
    const response = await this.client.post(`/v1/memories/${id}/review`, { action });
    return response.data;
  }

  async rollbackMemory(id: string, version: number): Promise<any> {
    const response = await this.client.post(`/v1/memories/${id}/rollback`, { version });
    return response.data;
  }

  // Batch
  async importBatch(memories: any[]): Promise<any> {
    const response = await this.client.post("/v1/batch/import", { memories });
    return response.data;
  }

  async exportBatch(): Promise<any> {
    const response = await this.client.get("/v1/batch/export");
    return response.data;
  }

  // Webhooks
  async listWebhooks(): Promise<any> {
    const response = await this.client.get("/v1/webhooks");
    return response.data;
  }

  async createWebhook(url: string, events: string[]): Promise<any> {
    const response = await this.client.post("/v1/webhooks", { url, events });
    return response.data;
  }

  async deleteWebhook(id: string): Promise<void> {
    await this.client.delete(`/v1/webhooks/${id}`);
  }

  // Conflicts
  async detectConflicts(memoryId: string): Promise<any> {
    const response = await this.client.post(`/v1/conflicts/detect/${memoryId}`);
    return response.data;
  }

  async scanConflicts(): Promise<any> {
    const response = await this.client.post("/v1/conflicts/scan");
    return response.data;
  }

  // Notes
  async getNotes(): Promise<any> {
    const response = await this.client.get("/v1/notes");
    return response.data;
  }

  async getHindsightNotes(): Promise<any> {
    const response = await this.client.get("/v1/notes/hindsight");
    return response.data;
  }

  async analyzeSession(sessionId: string): Promise<any> {
    const response = await this.client.post("/v1/notes/analyze", { sessionId });
    return response.data;
  }

  // Sessions
  async getSessionEvents(sessionId: string): Promise<any> {
    const response = await this.client.get(`/v1/sessions/${sessionId}/events`);
    return response.data;
  }

  // Knowledge Graph
  async getKnowledgeGraph(limit: number = 100): Promise<any> {
    const response = await this.client.get("/v1/entity-graph/knowledge-graph", {
      params: { limit },
    });
    return response.data;
  }

  // Audit Logs
  async getAuditLogs(limit: number = 100): Promise<any> {
    const response = await this.client.get("/v1/audit-logs", {
      params: { limit },
    });
    return response.data;
  }

  // Getter para o cliente axios bruto (para casos específicos)
  get axiosInstance(): AxiosInstance {
    return this.client;
  }
}

// Instância singleton
export const api = new ApiClient();

// Funções auxiliares para tratamento de erros
export function isApiError(error: unknown): error is ApiError {
  return typeof error === "object" && error !== null && "message" in error;
}

export function getErrorMessage(error: unknown): string {
  if (isApiError(error)) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Erro desconhecido";
}
