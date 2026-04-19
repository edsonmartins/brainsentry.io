import { test, expect } from "../fixtures/auth.fixture";
import { mockAdminApis } from "../helpers/admin-mocks";
import { ROUTES } from "../helpers/constants";

test.describe("Admin Coverage", () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await mockAdminApis(authenticatedPage);
  });

  test("covers relationships workflows", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.relationships);

    await expect(authenticatedPage.getByRole("heading", { name: "Grafo de Conhecimento" }).first()).toBeVisible();
    await authenticatedPage.getByPlaceholder(/Quais clientes compraram/i).fill("Quais memórias falam de autenticação?");
    await authenticatedPage.getByRole("button", { name: /Consultar/i }).click();
    await expect(authenticatedPage.locator("pre").filter({ hasText: /MATCH \(m:Memory\)/i })).toBeVisible();

    await authenticatedPage.getByRole("button", { name: /Memórias Conectadas/i }).click();
    await authenticatedPage.getByPlaceholder(/Busque uma memória/i).fill("auth");
    await expect(authenticatedPage.getByText("Autenticacao com refresh token")).toBeVisible();
  });

  test("covers timeline filters", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.timeline);

    await expect(authenticatedPage.getByRole("heading", { name: "Linha do Tempo" })).toBeVisible();
    await expect(authenticatedPage.getByText(/Mostrando 3 de 3 eventos/i)).toBeVisible();

    await authenticatedPage.getByRole("button", { name: "CRITICAL" }).click();
    await expect(authenticatedPage.getByText(/Mostrando 1 de 1 eventos/i)).toBeVisible();
  });

  test("covers audit list", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.audit);

    await expect(authenticatedPage.getByText("Histórico de Eventos")).toBeVisible();
    await expect(authenticatedPage.getByText("Injeção de Contexto")).toBeVisible();
    await expect(authenticatedPage.getByText("1 criadas")).toBeVisible();
  });

  test("covers users management", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.users);

    await expect(authenticatedPage.getByText(/Usuários \(2 usuários\)/i)).toBeVisible();
    await authenticatedPage.getByRole("button", { name: /Novo Usuário/i }).click();
    await authenticatedPage.getByRole("heading", { name: "Novo Usuário" }).waitFor();

    await authenticatedPage.getByPlaceholder("usuario@exemplo.com").fill("novo@empresa.com");
    await authenticatedPage.getByPlaceholder("Nome do usuário").fill("Novo QA");
    await authenticatedPage.locator('input[type="password"]').fill("demo123");
    await authenticatedPage.getByRole("button", { name: /Criar Usuário/i }).click();

    await expect(authenticatedPage.getByText("novo@empresa.com", { exact: true }).first()).toBeVisible();
  });

  test("covers tenants management", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.tenants);

    await expect(authenticatedPage.getByText("BrainSentry Labs")).toBeVisible();
    await authenticatedPage.getByRole("button", { name: /Novo Tenant/i }).click();

    await authenticatedPage.getByPlaceholder("Minha Organização").fill("BrainSentry QA");
    await expect(authenticatedPage.getByPlaceholder("minha-organizacao")).toHaveValue("brainsentry-qa");
    await authenticatedPage.getByRole("button", { name: /Criar Tenant/i }).click();

    await expect(authenticatedPage.getByRole("heading", { name: "BrainSentry QA" })).toBeVisible();
  });

  test("covers configuration sections", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.configuration);

    await authenticatedPage.getByLabel("Nome da Aplicação").fill("Brain Sentry Admin");
    await authenticatedPage.getByRole("button", { name: /Salvar Alterações/i }).click();

    await authenticatedPage.getByRole("button", { name: /Webhooks/i }).click();
    await authenticatedPage.getByPlaceholder("https://example.com/webhook").fill("https://hooks.example.com/new");
    await authenticatedPage.getByRole("button", { name: /Adicionar/i }).click();
    await expect(authenticatedPage.getByText("https://hooks.example.com/new")).toBeVisible();

    await authenticatedPage.getByRole("button", { name: /PII Scanner/i }).click();
    await authenticatedPage.getByPlaceholder(/Cole o texto para escanear/i).fill("Contato: cliente@empresa.com");
    await authenticatedPage.getByRole("button", { name: /Escanear PII/i }).click();
    await expect(authenticatedPage.locator("span").filter({ hasText: "cliente@empresa.com" })).toBeVisible();
  });

  test("covers analytics benchmark", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.analytics);

    await expect(authenticatedPage.getByText("Benchmark de Retrieval")).toBeVisible();
    await authenticatedPage.getByRole("button", { name: /Executar Benchmark/i }).click();
    await expect(authenticatedPage.getByText("Recall Médio")).toBeVisible();
    await expect(authenticatedPage.getByText("7.2 QPS")).toBeVisible();
  });

  test("covers profile and playground", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.profile);
    await expect(authenticatedPage.getByText("Perfil Estático")).toBeVisible();
    await expect(authenticatedPage.getByText("Playwright")).toBeVisible();

    await authenticatedPage.goto(ROUTES.playground);
    await authenticatedPage.getByPlaceholder(/Digite um prompt/i).fill("Como melhorar o admin?");
    await authenticatedPage.getByRole("button", { name: /Interceptar/i }).click();
    await expect(authenticatedPage.getByText("Prompt Enhanced")).toBeVisible();

    await authenticatedPage.getByPlaceholder(/Quais memórias estão relacionadas/i).fill("Quais memórias falam de autenticação?");
    await authenticatedPage.getByRole("button", { name: /Perguntar/i }).click();
    await expect(authenticatedPage.getByText("Cypher Gerado")).toBeVisible();
  });

  test("covers connectors, notes and tasks", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.connectors);
    await expect(authenticatedPage.getByText("GitHub")).toBeVisible();
    await authenticatedPage.getByRole("button", { name: /Sincronizar/i }).first().click();
    await expect(authenticatedPage.getByText("Documentos", { exact: true }).first()).toBeVisible();

    await authenticatedPage.goto(ROUTES.notes);
    await expect(authenticatedPage.getByText("Sessão de autenticação")).toBeVisible();
    await authenticatedPage.getByRole("button", { name: /Hindsight/i }).click();
    await expect(authenticatedPage.getByText(/Mocks centralizados reduziram flakes/i)).toBeVisible();

    await authenticatedPage.goto(ROUTES.tasks);
    await expect(authenticatedPage.getByText("Pendentes")).toBeVisible();
    await expect(authenticatedPage.getByText("Taxa de Sucesso")).toBeVisible();
  });
});
