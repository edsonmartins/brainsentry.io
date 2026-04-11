import { test, expect } from "../fixtures/auth.fixture";
import { mockAdminApis } from "../helpers/admin-mocks";
import { ROUTES } from "../helpers/constants";

test.describe("Search", () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await mockAdminApis(authenticatedPage);
    await authenticatedPage.goto(ROUTES.search);
  });

  test("runs semantic search and clears filters", async ({ authenticatedPage }) => {
    const input = authenticatedPage.getByPlaceholder(/Digite sua dúvida/i);

    await expect(input).toBeVisible();
    await input.fill("auth");
    await authenticatedPage.getByRole("button", { name: /^Buscar$/i }).click();

    await expect(authenticatedPage.getByText("Autenticacao com refresh token")).toBeVisible();
    await expect(authenticatedPage.getByText(/resultados encontrados/i)).toBeVisible();

    await authenticatedPage.getByRole("button", { name: /Limpar filtros/i }).click();
    await expect(input).toHaveValue("");
  });

  test("supports advanced mode", async ({ authenticatedPage }) => {
    await authenticatedPage.getByRole("button", { name: /Busca Avançada/i }).click();
    await expect(authenticatedPage.getByText(/Retrieval Planner/i)).toBeVisible();

    await authenticatedPage.getByPlaceholder(/Digite sua dúvida/i).fill("busca");
    await authenticatedPage.getByRole("button", { name: /^Buscar$/i }).click();

    await expect(authenticatedPage.getByText(/resultados encontrados/i)).toBeVisible();
  });
});
