import { test, expect } from "../fixtures/auth.fixture";
import { mockAdminApis } from "../helpers/admin-mocks";
import { ROUTES } from "../helpers/constants";

test.describe("Memory Admin", () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await mockAdminApis(authenticatedPage);
    await authenticatedPage.goto(ROUTES.memories);
  });

  test("creates a memory from the dialog", async ({ authenticatedPage }) => {
    await authenticatedPage.getByRole("button", { name: /Nova Memória/i }).click();
    await expect(authenticatedPage.getByRole("heading", { name: "Nova Memória" })).toBeVisible();

    await authenticatedPage.locator("#content").fill("Cobertura E2E para todo o admin.");
    await authenticatedPage.locator("#summary").fill("Cobertura E2E do admin");
    await authenticatedPage.getByRole("button", { name: /^Salvar$/i }).click();

    await expect(authenticatedPage.getByText("Cobertura E2E do admin")).toBeVisible();
  });

  test("searches and opens maintenance dialogs", async ({ authenticatedPage }) => {
    await expect(authenticatedPage.getByText("Autenticacao com refresh token", { exact: true })).toBeVisible();

    await authenticatedPage.getByTitle("Histórico de versões").first().click();
    await expect(authenticatedPage.getByText("Histórico de Versões")).toBeVisible();
    await authenticatedPage.keyboard.press("Escape");

    await authenticatedPage.getByTitle("Detectar conflitos").first().click();
    await expect(authenticatedPage.getByText("Detecção de Conflitos")).toBeVisible();
    await expect(authenticatedPage.getByText(/Ambas descrevem comportamento semelhante/i)).toBeVisible();
  });
});
