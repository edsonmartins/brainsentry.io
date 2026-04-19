import { test, expect } from "../fixtures/auth.fixture";
import { mockAdminApis } from "../helpers/admin-mocks";
import { ROUTES } from "../helpers/constants";

test.describe("Dashboard", () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await mockAdminApis(authenticatedPage);
    await authenticatedPage.goto(ROUTES.dashboard);
  });

  test("shows core stats and recent memories", async ({ authenticatedPage }) => {
    await expect(authenticatedPage.getByText("Total de Memórias")).toBeVisible();
    await expect(authenticatedPage.getByText("Categorias", { exact: true })).toBeVisible();
    await expect(authenticatedPage.getByText("Críticas", { exact: true }).first()).toBeVisible();
    await expect(authenticatedPage.getByText("Ativas 24h")).toBeVisible();
    await expect(authenticatedPage.getByText("Memórias Recentes")).toBeVisible();
    await expect(authenticatedPage.getByText("Autenticacao com refresh token")).toBeVisible();
  });

  test("navigates through quick actions", async ({ authenticatedPage }) => {
    await authenticatedPage.getByRole("button", { name: /Nova Memória/i }).click();
    await expect(authenticatedPage).toHaveURL(/\/app\/memories/);

    await authenticatedPage.goto(ROUTES.dashboard);
    await authenticatedPage.getByRole("button", { name: /^Buscar$/i }).click();
    await expect(authenticatedPage).toHaveURL(/\/app\/search/);
  });
});
