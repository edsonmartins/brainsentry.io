import { type Page, type Route } from "@playwright/test";
import { DEFAULT_TENANT_ID, DEMO_EMAIL, STORAGE_KEYS } from "./constants";

type Memory = {
  id: string;
  content: string;
  summary: string;
  category: string;
  importance: string;
  tags: string[];
  createdAt: string;
  updatedAt?: string;
  accessCount?: number;
  injectionCount?: number;
  helpfulCount?: number;
};

type User = {
  id: string;
  email: string;
  name?: string;
  tenantId: string;
  roles: string[];
  active: boolean;
  createdAt: string;
  lastLoginAt?: string;
};

type Tenant = {
  id: string;
  name: string;
  slug: string;
  active: boolean;
  maxMemories?: number;
  maxUsers?: number;
  createdAt: string;
};

const API_PATTERN = /^https?:\/\/localhost:(8080|8081)(?:\/api)?\/v1\//;
const JSON_HEADERS = { "access-control-allow-origin": "*", "content-type": "application/json" };

export const MOCK_AUTH = {
  token: "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjQxMDI0NDQ4MDAsInN1YiI6InVzZXItYWRtaW4ifQ.signature",
  user: {
    id: "user-admin",
    email: DEMO_EMAIL,
    name: "Demo Admin",
    tenantId: DEFAULT_TENANT_ID,
    roles: ["ADMIN"],
  },
  tenantId: DEFAULT_TENANT_ID,
};

const baseMemories = (): Memory[] => [
  {
    id: "mem-auth",
    summary: "Autenticacao com refresh token",
    content: "Fluxo de autenticacao com JWT, refresh token e tenant header.",
    category: "INSIGHT",
    importance: "CRITICAL",
    tags: ["auth", "jwt"],
    createdAt: "2026-04-09T10:00:00.000Z",
    accessCount: 12,
    injectionCount: 5,
    helpfulCount: 4,
  },
  {
    id: "mem-search",
    summary: "Busca semantica no admin",
    content: "Planejamento multi-round para recuperar memorias com score trace.",
    category: "KNOWLEDGE",
    importance: "IMPORTANT",
    tags: ["search", "planner"],
    createdAt: "2026-04-08T15:30:00.000Z",
    accessCount: 9,
    injectionCount: 3,
    helpfulCount: 2,
  },
  {
    id: "mem-ops",
    summary: "Monitoramento de tarefas",
    content: "Dashboard de filas, retries e circuit breakers para jobs assíncronos.",
    category: "ACTION",
    importance: "MINOR",
    tags: ["tasks", "ops"],
    createdAt: "2026-04-07T09:15:00.000Z",
    accessCount: 4,
    injectionCount: 1,
    helpfulCount: 1,
  },
];

const baseUsers = (): User[] => [
  {
    id: "user-admin",
    email: DEMO_EMAIL,
    name: "Demo Admin",
    tenantId: DEFAULT_TENANT_ID,
    roles: ["ADMIN"],
    active: true,
    createdAt: "2026-04-01T08:00:00.000Z",
    lastLoginAt: "2026-04-10T09:00:00.000Z",
  },
  {
    id: "user-analyst",
    email: "ana@example.com",
    name: "Ana Analyst",
    tenantId: DEFAULT_TENANT_ID,
    roles: ["USER", "MODERATOR"],
    active: true,
    createdAt: "2026-04-02T08:00:00.000Z",
    lastLoginAt: "2026-04-09T13:00:00.000Z",
  },
];

const baseTenants = (): Tenant[] => [
  {
    id: DEFAULT_TENANT_ID,
    name: "BrainSentry Core",
    slug: "brainsentry-core",
    active: true,
    maxMemories: 5000,
    maxUsers: 50,
    createdAt: "2026-03-20T10:00:00.000Z",
  },
  {
    id: "tenant-labs",
    name: "BrainSentry Labs",
    slug: "brainsentry-labs",
    active: true,
    maxMemories: 2000,
    maxUsers: 20,
    createdAt: "2026-03-25T10:00:00.000Z",
  },
];

function json(route: Route, body: unknown, status = 200) {
  return route.fulfill({
    status,
    headers: JSON_HEADERS,
    body: JSON.stringify(body),
  });
}

function text(route: Route, body: string, contentType = "text/plain") {
  return route.fulfill({
    status: 200,
    headers: {
      "access-control-allow-origin": "*",
      "content-type": contentType,
    },
    body,
  });
}

function normalizePath(pathname: string) {
  return pathname.startsWith("/api/") ? pathname.slice(4) : pathname;
}

export async function seedAuthenticatedSession(page: Page) {
  await page.addInitScript(
    ({ auth, keys }) => {
      localStorage.setItem(keys.token, auth.token);
      localStorage.setItem(keys.user, JSON.stringify(auth.user));
      localStorage.setItem(keys.tenantId, auth.tenantId);
    },
    { auth: MOCK_AUTH, keys: STORAGE_KEYS }
  );
}

export async function mockAuthApis(page: Page) {
  await page.route(API_PATTERN, async (route) => {
    const url = new URL(route.request().url());
    const path = normalizePath(url.pathname);
    const method = route.request().method();

    if (path === "/v1/auth/demo" && method === "POST") {
      return json(route, { ok: true });
    }

    if (path === "/v1/auth/login" && method === "POST") {
      const payload = route.request().postDataJSON() as { email?: string; password?: string };
      if (payload.email === DEMO_EMAIL && payload.password === "demo123") {
        return json(route, MOCK_AUTH);
      }
      return json(route, { message: "Credenciais inválidas" }, 401);
    }

    if (path === "/v1/auth/logout" && method === "POST") {
      return json(route, { ok: true });
    }

    if (path === "/v1/auth/refresh" && method === "POST") {
      return json(route, { token: MOCK_AUTH.token });
    }

    return route.fallback();
  });
}

export async function mockAdminApis(page: Page) {
  const memories = baseMemories();
  const users = baseUsers();
  const tenants = baseTenants();
  let webhooks = [
    {
      id: "wh-1",
      url: "https://hooks.example.com/brain",
      events: ["memory.created", "memory.updated"],
      active: true,
      createdAt: "2026-04-01T12:00:00.000Z",
    },
  ];

  await page.route(API_PATTERN, async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = normalizePath(url.pathname);
    const method = request.method();

    if (path === "/v1/stats/overview" && method === "GET") {
      return json(route, {
        totalMemories: memories.length,
        memoriesByCategory: { INSIGHT: 1, KNOWLEDGE: 1, ACTION: 1 },
        memoriesByImportance: { CRITICAL: 1, IMPORTANT: 1, MINOR: 1 },
        requestsToday: 42,
        injectionRate: 0.34,
        avgLatencyMs: 182,
        helpfulnessRate: 0.88,
        totalInjections: 19,
        activeMemories24h: 2,
      });
    }

    if (path === "/v1/memories" && method === "GET") {
      const pageParam = Number(url.searchParams.get("page") ?? 0);
      const sizeParam = Number(url.searchParams.get("size") ?? 20);
      const start = pageParam * sizeParam;
      const slice = memories.slice(start, start + sizeParam);
      return json(route, {
        memories: slice,
        total: memories.length,
        page: pageParam,
        size: sizeParam,
        totalPages: Math.max(1, Math.ceil(memories.length / sizeParam)),
      });
    }

    if (path === "/v1/memories" && method === "POST") {
      const payload = request.postDataJSON() as Partial<Memory>;
      const created: Memory = {
        id: `mem-${Date.now()}`,
        summary: payload.summary ?? "Nova memória",
        content: payload.content ?? "",
        category: payload.category ?? "INSIGHT",
        importance: payload.importance ?? "IMPORTANT",
        tags: payload.tags ?? [],
        createdAt: "2026-04-10T10:00:00.000Z",
        accessCount: 0,
        injectionCount: 0,
        helpfulCount: 0,
      };
      memories.unshift(created);
      return json(route, created, 201);
    }

    if (path === "/v1/memories/search" && method === "POST") {
      const payload = request.postDataJSON() as { query?: string; category?: string; importance?: string; limit?: number };
      const query = (payload.query ?? "").toLowerCase();
      let results = memories.filter((memory) =>
        query === "*" ||
        memory.summary.toLowerCase().includes(query) ||
        memory.content.toLowerCase().includes(query) ||
        memory.tags.some((tag) => tag.toLowerCase().includes(query))
      );

      if (payload.category) {
        results = results.filter((memory) => memory.category === payload.category);
      }
      if (payload.importance) {
        results = results.filter((memory) => memory.importance === payload.importance);
      }

      const limited = results.slice(0, payload.limit ?? 10).map((memory, index) => ({
        ...memory,
        score: 0.91 - index * 0.08,
        scoreTrace: {
          vectorScore: 0.82,
          graphScore: 0.61,
          recencyScore: 0.55,
          importanceBoost: 0.2,
          totalScore: 0.91 - index * 0.08,
          explanation: "Combina similaridade vetorial e contexto do grafo.",
        },
      }));

      return json(route, {
        results: limited,
        total: results.length,
        searchTimeMs: 48,
      });
    }

    if (path === "/v1/memories/plan-search" && method === "POST") {
      const payload = request.postDataJSON() as { query?: string; limit?: number };
      const results = memories
        .filter((memory) =>
          memory.summary.toLowerCase().includes((payload.query ?? "").toLowerCase()) ||
          memory.content.toLowerCase().includes((payload.query ?? "").toLowerCase()) ||
          memory.tags.some((tag) => tag.toLowerCase().includes((payload.query ?? "").toLowerCase()))
        )
        .slice(0, payload.limit ?? 10);

      return json(route, {
        query: payload.query ?? "",
        rounds: [
          {
            round: 1,
            subQuery: payload.query ?? "",
            results,
            coverage: 0.82,
          },
        ],
        finalResults: results,
        totalCoverage: 0.82,
        searchTimeMs: 77,
      });
    }

    if (/^\/v1\/memories\/[^/]+$/.test(path) && method === "PUT") {
      const id = path.split("/")[3];
      const payload = request.postDataJSON() as Partial<Memory>;
      const index = memories.findIndex((memory) => memory.id === id);
      if (index >= 0) {
        memories[index] = { ...memories[index], ...payload, updatedAt: "2026-04-10T11:00:00.000Z" };
        return json(route, memories[index]);
      }
      return json(route, { message: "Memory not found" }, 404);
    }

    if (/^\/v1\/memories\/[^/]+$/.test(path) && method === "DELETE") {
      const id = path.split("/")[3];
      const index = memories.findIndex((memory) => memory.id === id);
      if (index >= 0) memories.splice(index, 1);
      return json(route, { ok: true });
    }

    if (/^\/v1\/memories\/[^/]+\/versions$/.test(path) && method === "GET") {
      return json(route, {
        versions: [
          {
            version: 3,
            content: "Conteudo atual",
            summary: "Autenticacao com refresh token",
            category: "INSIGHT",
            importance: "CRITICAL",
            tags: ["auth"],
            changedAt: "2026-04-10T09:00:00.000Z",
            changedBy: "demo@example.com",
          },
          {
            version: 2,
            content: "Conteudo anterior",
            summary: "Autenticacao inicial",
            category: "INSIGHT",
            importance: "IMPORTANT",
            tags: ["auth"],
            changedAt: "2026-04-09T09:00:00.000Z",
            changedBy: "demo@example.com",
          },
        ],
      });
    }

    if (/^\/v1\/memories\/[^/]+\/flag$/.test(path) && method === "POST") {
      return json(route, { ok: true });
    }

    if (/^\/v1\/memories\/[^/]+\/review$/.test(path) && method === "POST") {
      return json(route, { ok: true });
    }

    if (/^\/v1\/memories\/[^/]+\/rollback$/.test(path) && method === "POST") {
      return json(route, { ok: true });
    }

    if (/^\/v1\/conflicts\/detect\/[^/]+$/.test(path) && method === "POST") {
      const memoryId = path.split("/")[4];
      return json(route, {
        memoryId,
        conflicts: [
          {
            conflictingMemoryId: "mem-search",
            conflictingSummary: "Busca semantica no admin",
            reason: "Ambas descrevem comportamento semelhante com termos divergentes.",
            severity: "medium",
          },
        ],
      });
    }

    if (path === "/v1/batch/export" && method === "GET") {
      return text(route, JSON.stringify({ memories }, null, 2), "application/json");
    }

    if (path === "/v1/batch/import" && method === "POST") {
      return json(route, { imported: 2 });
    }

    if (path === "/v1/profile" && method === "GET") {
      return json(route, {
        staticProfile: {
          traits: ["analitico", "orientado a qualidade"],
          preferences: ["testes determinísticos", "tipagem forte"],
          expertise: ["Playwright", "React", "Observabilidade"],
          summary: "Perfil consistente com foco em confiabilidade e cobertura.",
        },
        dynamicProfile: {
          recentFocus: ["Cobertura E2E do admin", "Métricas operacionais"],
          goals: ["Reduzir regressões no painel", "Automatizar smoke flows"],
          activity: "Alta atividade nas áreas de admin e busca semântica.",
          summary: "Nas últimas sessões o foco esteve em estabilidade do painel administrativo.",
        },
        generatedAt: "2026-04-10T10:00:00.000Z",
      });
    }

    if (path === "/v1/intercept" && method === "POST") {
      return json(route, {
        enhanced: true,
        originalPrompt: "Como melhorar o admin?",
        enhancedPrompt: "Como melhorar o admin com foco em confiabilidade, observabilidade e testes?",
        contextInjected: "Memórias de auth, busca e operações foram usadas.",
        memoriesUsed: memories.slice(0, 2).map((memory) => ({ id: memory.id, summary: memory.summary })),
        notesUsed: [{ id: "note-1", title: "Lição de rollout" }],
        latencyMs: 123,
        reasoning: "Priorizei experiências com maior risco de regressão.",
        confidence: 0.92,
        tokensInjected: 188,
        llmCalls: 2,
      });
    }

    if (path === "/v1/graph/nl-query" && method === "POST") {
      return json(route, {
        question: "Quais memórias falam de autenticação?",
        cypher: "MATCH (m:Memory)-[:RELATES_TO]->(t:Topic {name:'auth'}) RETURN m",
        results: [{ summary: "Autenticacao com refresh token" }],
        attempts: 1,
      });
    }

    if (path === "/v1/connectors" && method === "GET") {
      return json(route, { connectors: ["github", "notion"] });
    }

    if (/^\/v1\/connectors\/[^/]+\/sync$/.test(path) && method === "POST") {
      const connector = path.split("/")[3];
      return json(route, {
        connector,
        documentsFound: connector === "github" ? 14 : 6,
        chunksCreated: connector === "github" ? 32 : 11,
        tasksSubmitted: connector === "github" ? 5 : 2,
      });
    }

    if (path === "/v1/connectors/sync-all" && method === "POST") {
      return json(route, {
        github: { connector: "github", documentsFound: 14, chunksCreated: 32, tasksSubmitted: 5 },
        notion: { connector: "notion", documentsFound: 6, chunksCreated: 11, tasksSubmitted: 2 },
      });
    }

    if (path === "/v1/notes" && method === "GET") {
      return json(route, [
        {
          id: "note-1",
          title: "Sessão de autenticação",
          content: "Validar tenant header em todos os requests do admin.",
          noteType: "SESSION",
          sessionId: "sess-1",
          createdAt: "2026-04-10T08:00:00.000Z",
        },
      ]);
    }

    if (path === "/v1/notes/hindsight" && method === "GET") {
      return json(route, [
        {
          id: "hs-1",
          sessionId: "sess-2",
          content: "Mocks centralizados reduziram flakes na suíte.",
          impact: "Acelera feedback e elimina dependência do backend local.",
          createdAt: "2026-04-10T08:30:00.000Z",
        },
      ]);
    }

    if (path === "/v1/tasks/metrics" && method === "GET") {
      return json(route, { processed: 34, failed: 2, recovered: 1 });
    }

    if (path === "/v1/tasks/pending" && method === "GET") {
      return json(route, { pending: 7 });
    }

    if (path.startsWith("/v1/audit/logs/export") && method === "GET") {
      return text(route, "id,eventType\n1,memory_created\n", "text/csv");
    }

    if (path.startsWith("/v1/audit/logs/stats") && method === "GET") {
      return json(route, {
        totalEvents: 120,
        eventsByType: { memory_created: 20, memory_updated: 15, context_injection: 85 },
        eventsByUser: { "demo@example.com": 110, "ana@example.com": 10 },
        recentActivity: 18,
      });
    }

    if (path.startsWith("/v1/audit/logs") && method === "GET") {
      return json(route, {
        content: [
          {
            id: "audit-1",
            eventType: "memory_created",
            timestamp: "2026-04-10T09:00:00.000Z",
            userId: "user-admin",
            sessionId: "sess-1",
            outcome: "success",
            memoriesCreated: ["mem-auth"],
          },
        ],
        totalElements: 1,
      });
    }

    if (path === "/v1/users" && method === "GET") {
      return json(route, { content: users, totalElements: users.length });
    }

    if (path === "/v1/users" && method === "POST") {
      const payload = request.postDataJSON() as Partial<User> & { password?: string };
      const created: User = {
        id: `user-${Date.now()}`,
        email: payload.email ?? "novo@example.com",
        name: payload.name ?? "",
        tenantId: DEFAULT_TENANT_ID,
        roles: payload.roles ?? ["USER"],
        active: true,
        createdAt: "2026-04-10T10:10:00.000Z",
      };
      users.push(created);
      return json(route, created, 201);
    }

    if (/^\/v1\/users\/[^/]+$/.test(path) && method === "PATCH") {
      const id = path.split("/")[3];
      const payload = request.postDataJSON() as Partial<User>;
      const index = users.findIndex((user) => user.id === id);
      if (index >= 0) {
        users[index] = { ...users[index], ...payload };
        return json(route, users[index]);
      }
      return json(route, { message: "User not found" }, 404);
    }

    if (/^\/v1\/users\/[^/]+$/.test(path) && method === "DELETE") {
      const id = path.split("/")[3];
      const index = users.findIndex((user) => user.id === id);
      if (index >= 0) users.splice(index, 1);
      return json(route, { ok: true });
    }

    if (path === "/v1/tenants/stats" && method === "GET") {
      return json(route, [
        { tenantId: DEFAULT_TENANT_ID, memoryCount: 120, userCount: 8, relationshipCount: 14 },
        { tenantId: "tenant-labs", memoryCount: 64, userCount: 4, relationshipCount: 9 },
      ]);
    }

    if (path === "/v1/tenants" && method === "GET") {
      return json(route, { content: tenants, totalElements: tenants.length });
    }

    if (path === "/v1/tenants" && method === "POST") {
      const payload = request.postDataJSON() as Partial<Tenant>;
      const created: Tenant = {
        id: `tenant-${Date.now()}`,
        name: payload.name ?? "Novo Tenant",
        slug: payload.slug ?? "novo-tenant",
        active: payload.active ?? true,
        maxMemories: payload.maxMemories ?? 1000,
        maxUsers: payload.maxUsers ?? 10,
        createdAt: "2026-04-10T10:20:00.000Z",
      };
      tenants.push(created);
      return json(route, created, 201);
    }

    if (/^\/v1\/tenants\/[^/]+$/.test(path) && method === "PATCH") {
      const id = path.split("/")[3];
      const payload = request.postDataJSON() as Partial<Tenant>;
      const index = tenants.findIndex((tenant) => tenant.id === id);
      if (index >= 0) {
        tenants[index] = { ...tenants[index], ...payload };
        return json(route, tenants[index]);
      }
      return json(route, { message: "Tenant not found" }, 404);
    }

    if (/^\/v1\/tenants\/[^/]+$/.test(path) && method === "DELETE") {
      const id = path.split("/")[3];
      const index = tenants.findIndex((tenant) => tenant.id === id);
      if (index >= 0) tenants.splice(index, 1);
      return json(route, { ok: true });
    }

    if (path === "/v1/config" && method === "PUT") {
      return json(route, { ok: true });
    }

    if (path === "/v1/webhooks" && method === "GET") {
      return json(route, webhooks);
    }

    if (path === "/v1/webhooks" && method === "POST") {
      const payload = request.postDataJSON() as { url?: string; events?: string[] };
      const created = {
        id: `wh-${Date.now()}`,
        url: payload.url ?? "https://hooks.example.com/new",
        events: payload.events ?? ["memory.created"],
        active: true,
        createdAt: "2026-04-10T10:30:00.000Z",
      };
      webhooks = [...webhooks, created];
      return json(route, created, 201);
    }

    if (/^\/v1\/webhooks\/[^/]+$/.test(path) && method === "DELETE") {
      const id = path.split("/")[3];
      webhooks = webhooks.filter((webhook) => webhook.id !== id);
      return json(route, { ok: true });
    }

    if (path === "/v1/admin/circuit-breakers" && method === "GET") {
      return json(route, {
        breakers: [
          { name: "OpenRouter", state: "closed", failures: 0 },
          { name: "Embeddings", state: "half-open", failures: 2 },
        ],
      });
    }

    if (path === "/v1/admin/llm-metrics" && method === "GET") {
      return json(route, {
        metrics: [
          { model: "gpt-4.1-mini", totalRequests: 18, totalTokens: 12450, totalCost: 1.248, avgLatencyMs: 640, errorRate: 0.01 },
        ],
      });
    }

    if (path === "/v1/pii/scan" && method === "POST") {
      return json(route, {
        found: true,
        entities: [{ type: "EMAIL", value: "cliente@empresa.com", start: 9, end: 28 }],
        maskedText: "Contato: [EMAIL]",
      });
    }

    if (path === "/v1/benchmark/run" && method === "POST") {
      return json(route, {
        queryCount: 10,
        k: 10,
        avgLatencyMs: 128,
        p50LatencyMs: 110,
        p95LatencyMs: 180,
        p99LatencyMs: 220,
        avgRecall: 0.86,
        avgPrecision: 0.8,
        throughputQps: 7.2,
      });
    }

    if (path === "/v1/entity-graph/knowledge-graph" && method === "GET") {
      return json(route, {
        nodes: [
          { id: "node-auth", name: "Autenticacao", type: "CONCEITO" },
          { id: "node-user", name: "Usuario", type: "PESSOA" },
        ],
        edges: [
          {
            id: "edge-1",
            sourceId: "node-user",
            targetId: "node-auth",
            sourceName: "Usuario",
            targetName: "Autenticacao",
            type: "USES",
          },
        ],
        totalNodes: 2,
        totalEdges: 1,
      });
    }

    if (path === "/v1/relationships" && method === "GET") {
      return json(route, {
        relationships: [
          {
            id: "rel-1",
            fromMemoryId: "mem-auth",
            toMemoryId: "mem-search",
            type: "RELATED",
            strength: 0.9,
            createdAt: "2026-04-10T07:00:00.000Z",
            fromMemory: memories[0],
            toMemory: memories[1],
          },
        ],
        totalElements: 1,
      });
    }

    if (path === "/v1/relationships" && method === "POST") {
      return json(route, { ok: true }, 201);
    }

    if (path === "/v1/relationships/between" && method === "DELETE") {
      return json(route, { ok: true });
    }

    if (path === "/v1/graph/communities" && method === "GET") {
      return json(route, {
        communities: [
          { id: 1, members: ["node-auth", "node-user"], label: "Core auth", size: 2 },
        ],
      });
    }

    if (path === "/v1/memories/activate" && method === "POST") {
      return json(route, {
        activations: [
          { memoryId: "mem-search", score: 0.87 },
          { memoryId: "mem-ops", score: 0.65 },
        ],
      });
    }

    if (path === "/v1/memories/extract-all-entities" && method === "POST") {
      return text(route, "Entidades reprocessadas");
    }

    return route.fallback();
  });
}
