import { test as base, type Page } from "@playwright/test";
import { ApiHelper } from "../helpers/api.helper";
import { seedAuthenticatedSession } from "../helpers/admin-mocks";

type AuthFixtures = {
  authenticatedPage: Page;
  adminPage: Page;
  apiHelper: ApiHelper;
};

export const test = base.extend<AuthFixtures>({
  apiHelper: async ({}, use) => {
    const helper = new ApiHelper();
    await use(helper);
  },

  authenticatedPage: async ({ page }, use) => {
    await seedAuthenticatedSession(page);
    await use(page);
  },

  adminPage: async ({ page }, use) => {
    await seedAuthenticatedSession(page);
    await use(page);
  },
});

export { expect } from "@playwright/test";
