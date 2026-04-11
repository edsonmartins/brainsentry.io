import { defineConfig, devices } from "@playwright/test";

const apiBaseURL = process.env.E2E_API_BASE || process.env.VITE_API_URL || "http://localhost:8081/api";
const webPort = process.env.E2E_WEB_PORT || "4602";

export default defineConfig({
  testDir: "./e2e/tests",
  testMatch: /real-.*\.spec\.ts/,
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  reporter: process.env.CI ? "github" : [["html", { open: "never" }], ["list"]],
  timeout: 60_000,
  expect: { timeout: 15_000 },

  use: {
    baseURL: `http://127.0.0.1:${webPort}`,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],

  webServer: {
    command: `VITE_API_URL=${apiBaseURL} npm run dev -- --host 127.0.0.1 --port ${webPort} --strictPort`,
    url: `http://127.0.0.1:${webPort}`,
    reuseExistingServer: false,
    timeout: 30_000,
  },
});
