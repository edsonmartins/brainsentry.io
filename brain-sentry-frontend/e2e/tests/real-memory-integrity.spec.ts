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

const API_BASE_URL = process.env.E2E_API_BASE || process.env.VITE_API_URL || "http://localhost:8081/api";
const DEFAULT_TENANT_ID = "a9f814d2-4dae-41f3-851b-8aa3d4706561";

async function expectOk(response: { ok(): boolean; status(): number; text(): Promise<string> }, label: string) {
  if (!response.ok()) {
    throw new Error(`${label}: ${response.status()} ${await response.text()}`);
  }
}

async function demoLogin(): Promise<AuthResponse> {
  const bootstrap = await playwrightRequest.newContext({
    baseURL: API_BASE_URL,
    extraHTTPHeaders: {
      "Content-Type": "application/json",
      "X-Tenant-ID": DEFAULT_TENANT_ID,
    },
  });

  try {
    const response = await bootstrap.post("/v1/auth/demo");
    await expectOk(response, "demo login failed");
    return (await response.json()) as AuthResponse;
  } finally {
    await bootstrap.dispose();
  }
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
  const response = await api.get(`/v1/memories/${id}`);
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
    const marker = `integrity${Date.now()}`;
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

      const createdResponse = await api.post("/v1/memories", { data: createPayload });
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
          const versionsResponse = await api.get(`/v1/memories/${memoryId}/versions`);
          expect(versionsResponse.ok()).toBeTruthy();
          return ((await versionsResponse.json()) as VersionResponse[]).map((version) => version.version);
        }, { timeout: 10_000 })
        .toEqual(expect.arrayContaining([1]));

      await page.goto("/app/memories");
      await expect(page.getByRole("heading", { name: "Memórias" })).toBeVisible();
      await expect(page.getByText(createPayload.summary)).toBeVisible();
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

      const updatedResponse = await api.put(`/v1/memories/${memoryId}`, { data: updatePayload });
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
          const versionsResponse = await api.get(`/v1/memories/${memoryId}/versions`);
          expect(versionsResponse.ok()).toBeTruthy();
          return ((await versionsResponse.json()) as VersionResponse[]).map((version) => version.version);
        }, { timeout: 10_000 })
        .toEqual(expect.arrayContaining([1, 2]));

      const versionsResponse = await api.get(`/v1/memories/${memoryId}/versions`);
      const versions = (await versionsResponse.json()) as VersionResponse[];
      expect(versions.some((version) => version.summary === createPayload.summary)).toBeTruthy();
      expect(versions.some((version) => version.changeReason === updatePayload.changeReason)).toBeTruthy();

      await page.getByPlaceholder("Buscar memórias...").fill(marker);
      await expect(page.getByText(updatePayload.summary)).toBeVisible();
      await expect(page.getByText("IMPORTANT").first()).toBeVisible();
      await expect(page.getByText("DECISION").first()).toBeVisible();
      await page.getByText(updatePayload.summary).hover();
      await page.getByTitle("Histórico de versões").click();
      await expect(page.getByRole("heading", { name: "Histórico de Versões" })).toBeVisible();
      await expect(page.getByText("Versão 2")).toBeVisible();
      await expect(page.getByText(createPayload.summary)).toBeVisible();

      const deleteResponse = await api.delete(`/v1/memories/${memoryId}`);
      await expectOk(deleteResponse, "delete memory failed");
      memoryId = undefined;

      await expect
        .poll(async () => {
          const response = await api.get(`/v1/memories/${created.id}`);
          return response.status();
        }, { timeout: 10_000 })
        .toBe(404);

      await page.goto("/app/memories");
      await page.getByPlaceholder("Buscar memórias...").fill(marker);
      await expect(page.getByText(`Nenhuma memória encontrada para "${marker}"`)).toBeVisible();
    } finally {
      if (memoryId) {
        await api.delete(`/v1/memories/${memoryId}`);
      }
      await api.dispose();
    }
  });
});
