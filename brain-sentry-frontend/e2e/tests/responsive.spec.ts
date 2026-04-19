import { test, expect } from "../fixtures/auth.fixture";
import { mockAdminApis } from "../helpers/admin-mocks";
import { Sidebar } from "../pages/sidebar.page";
import { ROUTES } from "../helpers/constants";

test.describe("Responsive", () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test.beforeEach(async ({ authenticatedPage }) => {
    await mockAdminApis(authenticatedPage);
    await authenticatedPage.goto(ROUTES.dashboard);
  });

  test("opens mobile menu and navigates", async ({ authenticatedPage }) => {
    const sidebar = new Sidebar(authenticatedPage);

    await expect(sidebar.sidebar).toBeHidden();
    await expect(sidebar.mobileMenuToggle).toBeVisible();

    await sidebar.openMobileMenu();
    await expect(sidebar.getMobileNavItem("Dashboard")).toBeVisible();
    await expect(sidebar.getMobileNavItem("Linha do Tempo")).toBeVisible();

    await sidebar.getMobileNavItem("Memórias").click();
    await expect(authenticatedPage).toHaveURL(/\/app\/memories/);
  });
});
