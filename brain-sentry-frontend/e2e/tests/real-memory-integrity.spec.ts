import { expect, request as playwrightRequest, test, type APIRequestContext, type Page } from "@playwright/test";

type AuthResponse = {
  token: string;
  tenantId: string;
  user: {
    id: string;
    email: string;
    name?: string;
    roles?: string[];
    tenantId?: string;
  };
};

type MemoryResponse = {
  id: string;
  content: string;
  summary: string;
  category: string;
  importance: string;
  metadata?: Record<string, unknown>;
  tags?: string[];
  sourceType?: string;
  sourceReference?: string;
  createdBy?: string;
  tenantId: string;
  version: number;
  accessCount?: number;
  injectionCount?: number;
  memoryType?: string;
  emotionalWeight?: number;
  simHash?: string;
  validFrom?: string;
  validTo?: string;
  decayRate?: number;
};

type VersionResponse = {
  version: number;
  content: string;
  summary: string;
  category: string;
  importance: string;
  changeReason?: string;
};

type InterceptResponse = {
  enhanced: boolean;
  originalPrompt: string;
  enhancedPrompt?: string;
  contextInjected?: string;
  memoriesUsed?: Array<{
    id: string;
    summary?: string;
    category?: string;
    importance?: string;
    relevanceScore?: number;
    excerpt?: string;
  }>;
  notesUsed?: Array<unknown>;
  latencyMs: number;
  reasoning?: string;
  confidence?: number;
  tokensInjected: number;
  llmCalls: number;
};

type AutoForgetResponse = {
  ttl_expired: number;
  contradictions: number;
  low_value: number;
  total_deleted: number;
  deleted_ids?: string[];
  dry_run: boolean;
};

type SemanticConsolidationResponse = {
  semanticFacts?: unknown[];
  workflows?: unknown[];
  memoriesUsed: number;
  durationMs: number;
};

const RAW_API_BASE_URL = process.env.E2E_API_BASE || process.env.VITE_API_URL || "http://localhost:8081/api";
const RAW_API_BASE_URL_WITHOUT_TRAILING_SLASH = RAW_API_BASE_URL.replace(/\/+$/, "");
const API_BASE_URL = RAW_API_BASE_URL_WITHOUT_TRAILING_SLASH.endsWith("/api")
  ? `${RAW_API_BASE_URL_WITHOUT_TRAILING_SLASH}/`
  : `${RAW_API_BASE_URL_WITHOUT_TRAILING_SLASH}/api/`;
const DEFAULT_TENANT_ID = "a9f814d2-4dae-41f3-851b-8aa3d4706561";

function apiPath(path: string) {
  return path.replace(/^\//, "");
}

function uniqueMarker(prefix: string) {
  const alphaTimestamp = Date.now()
    .toString(36)
    .replace(/\d/g, (digit) => "abcdefghij"[Number(digit)]);
  return `${prefix}${alphaTimestamp}`;
}

async function expectOk(response: { ok(): boolean; status(): number; text(): Promise<string> }, label: string) {
  if (!response.ok()) {
    throw new Error(`${label}: ${response.status()} ${await response.text()}`);
  }
}

async function demoLogin(): Promise<AuthResponse> {
  const candidates = [API_BASE_URL];

  const failures: string[] = [];
  for (const baseURL of candidates) {
    const bootstrap = await playwrightRequest.newContext({
      baseURL,
      extraHTTPHeaders: {
        "Content-Type": "application/json",
        "X-Tenant-ID": DEFAULT_TENANT_ID,
      },
    });

    try {
      const response = await bootstrap.post(apiPath("/v1/auth/demo"));
      if (response.ok()) {
        return (await response.json()) as AuthResponse;
      }
      failures.push(`${baseURL}: ${response.status()} ${await response.text()}`);
    } finally {
      await bootstrap.dispose();
    }
  }

  throw new Error(`demo login failed for all API bases: ${failures.join(" | ")}`);
}

async function newAuthenticatedApi(auth: AuthResponse): Promise<APIRequestContext> {
  return playwrightRequest.newContext({
    baseURL: API_BASE_URL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${auth.token}`,
      "Content-Type": "application/json",
      "X-Tenant-ID": auth.tenantId || auth.user.tenantId || DEFAULT_TENANT_ID,
    },
  });
}

async function seedAuthenticatedBrowser(page: Page, auth: AuthResponse) {
  const tenantId = auth.tenantId || auth.user.tenantId || DEFAULT_TENANT_ID;
  const user = { ...auth.user, tenantId };

  await page.addInitScript(
    ({ token, storedUser, storedTenantId }) => {
      window.localStorage.setItem("brain_sentry_token", token);
      window.localStorage.setItem("brain_sentry_user", JSON.stringify(storedUser));
      window.localStorage.setItem("tenant_id", storedTenantId);
    },
    { token: auth.token, storedUser: user, storedTenantId: tenantId }
  );
}

async function getMemory(api: APIRequestContext, id: string): Promise<MemoryResponse> {
  const response = await api.get(apiPath(`/v1/memories/${id}`));
  await expectOk(response, "get memory failed");
  return (await response.json()) as MemoryResponse;
}

async function expectRawMemory(memory: MemoryResponse, expected: Partial<MemoryResponse> & { metadata?: Record<string, unknown> }) {
  for (const [key, value] of Object.entries(expected)) {
    if (key === "metadata") {
      continue;
    }
    expect(memory[key as keyof MemoryResponse]).toEqual(value);
  }

  if (expected.metadata) {
    expect(memory.metadata).toMatchObject(expected.metadata);
  }
}

test.describe("real backend memory integrity", () => {
  test.skip(process.env.E2E_REAL_API !== "1", "Set E2E_REAL_API=1 and use playwright.real.config.ts to run real backend tests.");

  test("creates, renders, edits, versions, deletes, and validates raw memory fields", async ({ page }) => {
    const auth = await demoLogin();
    const api = await newAuthenticatedApi(auth);
    const marker = uniqueMarker("integrity");
    let memoryId: string | undefined;

    await seedAuthenticatedBrowser(page, auth);

    try {
      const createPayload = {
        content: `Real backend memory integrity content ${marker}. PostgreSQL raw fields must remain intact.`,
        summary: `Real memory integrity ${marker}`,
        category: "KNOWLEDGE",
        importance: "CRITICAL",
        memoryType: "SEMANTIC",
        tags: ["e2e-real", "integrity", marker],
        metadata: {
          suite: "real-memory-integrity",
          marker,
          rawFieldCheck: true,
        },
        sourceType: "playwright",
        sourceReference: "admin-real-e2e",
        createdBy: auth.user.email,
        tenantId: auth.tenantId,
        emotionalWeight: 0.42,
      };

      const createdResponse = await api.post(apiPath("/v1/memories"), { data: createPayload });
      await expectOk(createdResponse, "create memory failed");
      const created = (await createdResponse.json()) as MemoryResponse;
      memoryId = created.id;

      await expect.poll(async () => getMemory(api, memoryId!)).toMatchObject({
        id: memoryId,
        version: 1,
        summary: createPayload.summary,
      });

      const rawCreated = await getMemory(api, memoryId);
      await expectRawMemory(rawCreated, {
        content: createPayload.content,
        summary: createPayload.summary,
        category: createPayload.category,
        importance: createPayload.importance,
        memoryType: createPayload.memoryType,
        sourceType: createPayload.sourceType,
        sourceReference: createPayload.sourceReference,
        createdBy: createPayload.createdBy,
        tenantId: createPayload.tenantId,
        metadata: createPayload.metadata,
      });
      expect(rawCreated.tags).toEqual(expect.arrayContaining(createPayload.tags));
      expect(rawCreated.simHash, "simHash must be persisted for dedup/integrity").toBeTruthy();
      expect(rawCreated.decayRate, "decay rate must be derived from memory type").toBeGreaterThan(0);
      expect(rawCreated.emotionalWeight).toBeCloseTo(createPayload.emotionalWeight, 2);

      await expect
        .poll(async () => {
          const versionsResponse = await api.get(apiPath(`/v1/memories/${memoryId}/versions`));
          expect(versionsResponse.ok()).toBeTruthy();
          return ((await versionsResponse.json()) as VersionResponse[]).map((version) => version.version);
        }, { timeout: 10_000 })
        .toEqual(expect.arrayContaining([1]));

      await page.goto("/app/memories");
      await expect(page.getByRole("heading", { name: "Memórias" })).toBeVisible();
      await expect(page.getByText(createPayload.summary, { exact: true })).toBeVisible();
      await expect(page.getByText("CRITICAL").first()).toBeVisible();
      await expect(page.getByText("KNOWLEDGE").first()).toBeVisible();
      await expect(page.getByText(marker).first()).toBeVisible();

      const updatePayload = {
        content: `Updated real backend memory integrity content ${marker}. Version history must preserve the original state.`,
        summary: `Updated real memory integrity ${marker}`,
        category: "DECISION",
        importance: "IMPORTANT",
        tags: ["e2e-real", "integrity-updated", marker],
        metadata: {
          suite: "real-memory-integrity",
          marker,
          edited: true,
        },
        changeReason: "playwright real integrity update",
      };

      const updatedResponse = await api.put(apiPath(`/v1/memories/${memoryId}`), { data: updatePayload });
      await expectOk(updatedResponse, "update memory failed");

      await expect
        .poll(async () => getMemory(api, memoryId!), { timeout: 10_000 })
        .toMatchObject({
          version: 2,
          summary: updatePayload.summary,
          category: updatePayload.category,
          importance: updatePayload.importance,
        });

      const updated = await getMemory(api, memoryId);
      await expectRawMemory(updated, {
        content: updatePayload.content,
        summary: updatePayload.summary,
        category: updatePayload.category,
        importance: updatePayload.importance,
        tenantId: createPayload.tenantId,
        metadata: updatePayload.metadata,
      });
      expect(updated.tags).toEqual(expect.arrayContaining(updatePayload.tags));
      expect(updated.version).toBe(2);

      await expect
        .poll(async () => {
          const versionsResponse = await api.get(apiPath(`/v1/memories/${memoryId}/versions`));
          expect(versionsResponse.ok()).toBeTruthy();
          return ((await versionsResponse.json()) as VersionResponse[]).map((version) => version.version);
        }, { timeout: 10_000 })
        .toEqual(expect.arrayContaining([1, 2]));

      const versionsResponse = await api.get(apiPath(`/v1/memories/${memoryId}/versions`));
      const versions = (await versionsResponse.json()) as VersionResponse[];
      expect(versions.some((version) => version.summary === createPayload.summary)).toBeTruthy();
      expect(versions.some((version) => version.changeReason === updatePayload.changeReason)).toBeTruthy();

      await page.getByPlaceholder("Buscar memórias...").fill(marker);
      await expect(page.getByText(updatePayload.summary, { exact: true })).toBeVisible();
      await expect(page.getByText("IMPORTANT").first()).toBeVisible();
      await expect(page.getByText("DECISION").first()).toBeVisible();
      await page.getByText(updatePayload.summary).hover();
      await page.getByTitle("Histórico de versões").click();
      await expect(page.getByRole("heading", { name: "Histórico de Versões" })).toBeVisible();
      await expect(page.getByText("Versão 2")).toBeVisible();
      await expect(page.getByText(createPayload.summary, { exact: true })).toBeVisible();

      const deleteResponse = await api.delete(apiPath(`/v1/memories/${memoryId}`));
      await expectOk(deleteResponse, "delete memory failed");
      memoryId = undefined;

      await expect
        .poll(async () => {
          const response = await api.get(apiPath(`/v1/memories/${created.id}`));
          return response.status();
        }, { timeout: 10_000 })
        .toBe(404);

      await page.goto("/app/memories");
      await page.getByPlaceholder("Buscar memórias...").fill(marker);
      await expect(page.getByText(`Nenhuma memória encontrada para "${marker}"`)).toBeVisible();
    } finally {
      if (memoryId) {
        await api.delete(apiPath(`/v1/memories/${memoryId}`));
      }
      await api.dispose();
    }
  });

  test("intercepts prompts with only active relevant memories", async () => {
    const auth = await demoLogin();
    const api = await newAuthenticatedApi(auth);
    const marker = uniqueMarker("intercept");
    const past = new Date(Date.now() - 60 * 60 * 1000).toISOString();
    const future = new Date(Date.now() + 60 * 60 * 1000).toISOString();
    const createdMemoryIds: string[] = [];

    try {
      const createMemory = async (payload: Record<string, unknown>) => {
        const response = await api.post(apiPath("/v1/memories"), { data: payload });
        await expectOk(response, "create intercept fixture memory failed");
        const memory = (await response.json()) as MemoryResponse;
        createdMemoryIds.push(memory.id);
        return memory;
      };

      const active = await createMemory({
        content: `When asked to implement ${marker} repository, use the event-driven memory gateway and preserve tenant isolation.`,
        summary: `Active intercept guidance ${marker}`,
        category: "KNOWLEDGE",
        importance: "CRITICAL",
        memoryType: "SEMANTIC",
        tags: ["e2e-real", "intercept", marker],
        metadata: {
          suite: "real-intercept",
          marker,
          expectedInContext: true,
        },
        sourceType: "playwright",
        sourceReference: "intercept-real-e2e",
        createdBy: auth.user.email,
        tenantId: auth.tenantId,
        validFrom: past,
        validTo: future,
      });

      const expired = await createMemory({
        content: `Expired rule to implement ${marker} repository with a stale direct-call pattern.`,
        summary: `Expired intercept guidance ${marker}`,
        category: "KNOWLEDGE",
        importance: "CRITICAL",
        memoryType: "SEMANTIC",
        tags: ["e2e-real", "intercept-expired", marker],
        metadata: {
          suite: "real-intercept",
          marker,
          expectedInContext: false,
        },
        sourceType: "playwright",
        sourceReference: "intercept-real-e2e",
        createdBy: auth.user.email,
        tenantId: auth.tenantId,
        validFrom: past,
        validTo: past,
      });

      const minor = await createMemory({
        content: `Minor note to implement ${marker} repository with a non-critical detail.`,
        summary: `Minor intercept note ${marker}`,
        category: "KNOWLEDGE",
        importance: "MINOR",
        memoryType: "SEMANTIC",
        tags: ["e2e-real", "intercept-minor", marker],
        metadata: {
          suite: "real-intercept",
          marker,
          expectedInContext: false,
        },
        sourceType: "playwright",
        sourceReference: "intercept-real-e2e",
        createdBy: auth.user.email,
        tenantId: auth.tenantId,
        validFrom: past,
        validTo: future,
      });

      const prompt = `implement ${marker} repository`;
      const interceptResponse = await api.post(apiPath("/v1/intercept"), {
        data: {
          prompt,
          userId: auth.user.id,
          tenantId: auth.tenantId,
          maxTokens: 600,
        },
      });
      await expectOk(interceptResponse, "intercept failed");
      const intercepted = (await interceptResponse.json()) as InterceptResponse;

      expect(intercepted.originalPrompt).toBe(prompt);
      expect(intercepted.enhanced).toBe(true);
      expect(intercepted.contextInjected).toContain(active.summary);
      expect(intercepted.contextInjected).not.toContain(expired.summary);
      expect(intercepted.contextInjected).not.toContain(minor.summary);
      expect(intercepted.enhancedPrompt).toContain(active.summary);
      expect(intercepted.enhancedPrompt).toContain(prompt);
      expect(intercepted.tokensInjected).toBeGreaterThan(0);
      expect(intercepted.memoriesUsed?.map((memory) => memory.id)).toContain(active.id);
      expect(intercepted.memoriesUsed?.map((memory) => memory.id)).not.toContain(expired.id);
      expect(intercepted.memoriesUsed?.map((memory) => memory.id)).not.toContain(minor.id);

      await expect
        .poll(async () => (await getMemory(api, active.id)).injectionCount || 0, { timeout: 10_000 })
        .toBeGreaterThan(0);
      expect((await getMemory(api, expired.id)).injectionCount || 0).toBe(0);
      expect((await getMemory(api, minor.id)).injectionCount || 0).toBe(0);
    } finally {
      for (const memoryId of createdMemoryIds.reverse()) {
        await api.delete(apiPath(`/v1/memories/${memoryId}`));
      }
      await api.dispose();
    }
  });

  test("validates learning lifecycle endpoints without mutating dry-run fixtures", async () => {
    const auth = await demoLogin();
    const api = await newAuthenticatedApi(auth);
    const marker = uniqueMarker("learning");
    const past = new Date(Date.now() - 60 * 60 * 1000).toISOString();
    let expiredMemoryId: string | undefined;
    let semanticFixtureId: string | undefined;

    try {
      const expiredResponse = await api.post(apiPath("/v1/memories"), {
        data: {
          content: `Expired auto-forget dry-run fixture ${marker}`,
          summary: `Expired auto-forget fixture ${marker}`,
          category: "KNOWLEDGE",
          importance: "MINOR",
          memoryType: "SEMANTIC",
          tags: ["e2e-real", "auto-forget", marker],
          metadata: { suite: "real-learning-lifecycle", marker, shouldRemainAfterDryRun: true },
          sourceType: "playwright",
          sourceReference: "learning-real-e2e",
          createdBy: auth.user.email,
          tenantId: auth.tenantId,
          validTo: past,
        },
      });
      await expectOk(expiredResponse, "create expired auto-forget fixture failed");
      expiredMemoryId = ((await expiredResponse.json()) as MemoryResponse).id;

      const autoForgetResponse = await api.post(apiPath("/v1/auto-forget?dryRun=true"));
      await expectOk(autoForgetResponse, "auto-forget dry-run failed");
      const autoForget = (await autoForgetResponse.json()) as AutoForgetResponse;
      expect(autoForget.dry_run).toBe(true);
      expect(autoForget.ttl_expired).toBeGreaterThan(0);
      expect(autoForget.total_deleted).toBeGreaterThan(0);
      expect(autoForget.deleted_ids || []).toContain(expiredMemoryId);

      const stillPresent = await getMemory(api, expiredMemoryId);
      expect(stillPresent.summary).toBe(`Expired auto-forget fixture ${marker}`);

      const semanticFixtureResponse = await api.post(apiPath("/v1/memories"), {
        data: {
          content: `Semantic consolidate minimum fixture ${marker}`,
          summary: `Semantic consolidate fixture ${marker}`,
          category: "KNOWLEDGE",
          importance: "IMPORTANT",
          memoryType: "SEMANTIC",
          tags: ["e2e-real", "semantic-consolidate", marker],
          metadata: { suite: "real-learning-lifecycle", marker },
          sourceType: "playwright",
          sourceReference: "learning-real-e2e",
          createdBy: auth.user.email,
          tenantId: auth.tenantId,
        },
      });
      await expectOk(semanticFixtureResponse, "create semantic consolidate fixture failed");
      semanticFixtureId = ((await semanticFixtureResponse.json()) as MemoryResponse).id;

      const semanticResponse = await api.post(apiPath("/v1/semantic/consolidate?minMemories=999999"));
      await expectOk(semanticResponse, "semantic consolidate minimum check failed");
      const semantic = (await semanticResponse.json()) as SemanticConsolidationResponse;
      expect(semantic.memoriesUsed).toBeGreaterThan(0);
      expect(semantic.semanticFacts || []).toHaveLength(0);
      expect(semantic.workflows || []).toHaveLength(0);
      expect(await getMemory(api, semanticFixtureId)).toMatchObject({
        id: semanticFixtureId,
        summary: `Semantic consolidate fixture ${marker}`,
      });
    } finally {
      for (const memoryId of [expiredMemoryId, semanticFixtureId]) {
        if (memoryId) {
          await api.delete(apiPath(`/v1/memories/${memoryId}`));
        }
      }
      await api.dispose();
    }
  });
});
