import { test, expect } from "@playwright/test";
import { LoginPage } from "../pages/login.page";
import { mockAdminApis, mockAuthApis } from "../helpers/admin-mocks";
import { DEMO_EMAIL, ROUTES, STORAGE_KEYS } from "../helpers/constants";

test.describe("Authentication", () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthApis(page);
    await mockAdminApis(page);
  });

  test("renders login form", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    await expect(loginPage.emailInput).toBeVisible();
    await expect(loginPage.passwordInput).toBeVisible();
    await expect(loginPage.submitButton).toBeVisible();
    await expect(loginPage.demoButton).toBeVisible();
  });

  test("logs in with demo credentials", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login(DEMO_EMAIL, "demo123");

    await expect(page).toHaveURL(/\/app\/dashboard/);
    await expect(page.getByText("Total Memories")).toBeVisible();
    await expect(page.locator("body")).toContainText(DEMO_EMAIL);
  });

  test("logs in via demo button", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginWithDemo();

    await expect(page).toHaveURL(/\/app\/dashboard/);
    await expect(page.getByText("Recent Memories")).toBeVisible();
  });

  test("shows error for invalid credentials", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login("wrong@example.com", "invalid");

    await expect(loginPage.errorMessage).toContainText("Credenciais inválidas");
    await expect(page).toHaveURL(/\/login/);
  });

  test("redirects protected route to login when not authenticated", async ({ page }) => {
    await page.goto(ROUTES.dashboard);
    await expect(page).toHaveURL(/\/login/);
  });

  test("persists authenticated session after reload", async ({ page }) => {
    await page.addInitScript(({ keys }) => {
      localStorage.setItem(keys.token, "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjQxMDI0NDQ4MDAsInN1YiI6InVzZXItYWRtaW4ifQ.signature");
      localStorage.setItem(keys.user, JSON.stringify({
        id: "user-admin",
        email: "demo@example.com",
        name: "Demo Admin",
        tenantId: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
        roles: ["ADMIN"],
      }));
      localStorage.setItem(keys.tenantId, "a9f814d2-4dae-41f3-851b-8aa3d4706561");
    }, { keys: STORAGE_KEYS });

    await page.goto(ROUTES.dashboard);
    await expect(page).toHaveURL(/\/app\/dashboard/);

    await page.reload();
    await expect(page).toHaveURL(/\/app\/dashboard/);
    await expect(page.getByText("Recent Memories")).toBeVisible();
  });
});
