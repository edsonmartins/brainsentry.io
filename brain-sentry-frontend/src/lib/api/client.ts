import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from "axios";

// Configuração base da API
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api";

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
export type MemoryCategory = "DECISION" | "PATTERN" | "ANTIPATTERN" | "DOMAIN" | "BUG" | "OPTIMIZATION" | "INTEGRATION";
export type ImportanceLevel = "CRITICAL" | "IMPORTANT" | "MINOR";

export interface Memory {
  id: string;
  tenantId?: string;
  content: string;
  summary: string;
  category: MemoryCategory;
  importance: ImportanceLevel;
  tags: string[];
  createdAt: string;
  updatedAt?: string;
  accessCount?: number;
  injectionCount?: number;
  helpfulCount?: number;
  embedding?: number[];
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
  page: number;
  size: number;
  totalPages: number;
}

export interface SearchRequest {
  query: string;
  limit?: number;
}

export interface MemoryStats {
  totalMemories: number;
  byCategory: Record<string, number>;
  byImportance: Record<string, number>;
  avgInjectionRate: number;
  avgHelpfulnessRate: number;
}

// Interceptador para adicionar headers de autenticação
const authRequestInterceptor = (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
  // Adicionar tenant ID se disponível
  const tenantId = localStorage.getItem("tenant_id") || "default";
  if (config.headers) {
    config.headers["X-Tenant-ID"] = tenantId;
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
    return response.data;
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
    const response = await this.client.post<Memory[]>("/v1/memories/search", {
      query,
      limit,
    });
    return response.data;
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
