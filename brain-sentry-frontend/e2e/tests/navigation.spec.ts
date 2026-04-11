import { test, expect } from "../fixtures/auth.fixture";
import { mockAdminApis } from "../helpers/admin-mocks";
import { Sidebar } from "../pages/sidebar.page";
import { ROUTES } from "../helpers/constants";

test.describe("Navigation", () => {
  test.use({ viewport: { width: 1280, height: 900 } });

  test.beforeEach(async ({ authenticatedPage }) => {
    await mockAdminApis(authenticatedPage);
    await authenticatedPage.goto(ROUTES.dashboard);
  });

  test("renders every admin item in the sidebar", async ({ authenticatedPage }) => {
    const sidebar = new Sidebar(authenticatedPage);

    for (const item of sidebar.getAllNavItems()) {
      await expect(sidebar.getNavItem(item)).toBeVisible();
    }

    await expect(sidebar.getNavItem("Timeline")).toBeVisible();
    await expect(sidebar.userEmail).toContainText("demo@example.com");
  });

  test("switches between key admin routes", async ({ authenticatedPage }) => {
    const sidebar = new Sidebar(authenticatedPage);

    await sidebar.navigateTo("Memórias");
    await expect(authenticatedPage).toHaveURL(/\/app\/memories/);

    await sidebar.navigateTo("Relacionamentos");
    await expect(authenticatedPage).toHaveURL(/\/app\/relationships/);

    await sidebar.navigateTo("Configurações");
    await expect(authenticatedPage).toHaveURL(/\/app\/configuration/);

    await sidebar.navigateTo("Dashboard");
    await expect(authenticatedPage).toHaveURL(/\/app\/dashboard/);
  });
});
