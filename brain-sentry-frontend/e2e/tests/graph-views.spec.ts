import { test, expect } from "../fixtures/auth.fixture";
import { mockAdminApis } from "../helpers/admin-mocks";
import { Sidebar } from "../pages/sidebar.page";
import { ROUTES } from "../helpers/constants";

test.describe("Graph Views", () => {
  test.use({ viewport: { width: 1280, height: 900 } });

  test.beforeEach(async ({ authenticatedPage }) => {
    await mockAdminApis(authenticatedPage);
    await authenticatedPage.goto(ROUTES.dashboard);
  });

  test("sidebar exposes all three graph views", async ({ authenticatedPage }) => {
    const sidebar = new Sidebar(authenticatedPage);

    await expect(sidebar.getNavItem("Grafo Global")).toBeVisible();
    await expect(sidebar.getNavItem("Ego-grafo")).toBeVisible();
    await expect(sidebar.getNavItem("Grafo Temporal")).toBeVisible();
  });

  test("global graph: renders canvas, filter chips and stats", async ({ authenticatedPage }) => {
    const sidebar = new Sidebar(authenticatedPage);
    await sidebar.navigateTo("Grafo Global");
    await expect(authenticatedPage).toHaveURL(/\/app\/graph\/global/);

    await expect(authenticatedPage.getByRole("heading", { name: "Grafo Global" })).toBeVisible();

    // Stats strip should show the node count from the mock (3 nodes).
    await expect(authenticatedPage.getByText(/\b3\s+nós\b/)).toBeVisible();
    await expect(authenticatedPage.getByText(/\b2\s+comunidades\b/)).toBeVisible();

    // Filter chips exist
    await expect(authenticatedPage.getByRole("button", { name: "INSIGHT", exact: true })).toBeVisible();
    await expect(authenticatedPage.getByRole("button", { name: "CRITICAL", exact: true })).toBeVisible();

    // The graph container is rendered and contains a canvas
    const container = authenticatedPage.getByTestId("graph-global-canvas");
    await expect(container).toBeVisible();
    await expect(container.locator("canvas")).toBeVisible();

    // Feedback overlay toggle
    await expect(authenticatedPage.getByText("Opacidade por feedback")).toBeVisible();
  });

  test("global graph: clicking a category chip keeps the canvas visible", async ({ authenticatedPage }) => {
    const sidebar = new Sidebar(authenticatedPage);
    await sidebar.navigateTo("Grafo Global");

    const container = authenticatedPage.getByTestId("graph-global-canvas");
    await expect(container.locator("canvas")).toBeVisible();

    await authenticatedPage.getByRole("button", { name: "KNOWLEDGE", exact: true }).click();

    await expect(authenticatedPage.getByText("Falha ao carregar grafo global")).toHaveCount(0);
    await expect(container.locator("canvas")).toBeVisible();
  });

  test("ego graph: empty state before seed, canvas after explore", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.graphEgo);
    await expect(authenticatedPage).toHaveURL(/\/app\/graph\/ego/);

    await expect(authenticatedPage.getByRole("heading", { name: "Ego-grafo" })).toBeVisible();

    const idInput = authenticatedPage.getByPlaceholder("UUID da memória-semente");
    await expect(idInput).toBeVisible();

    // Empty state visible before any seed
    await expect(authenticatedPage.getByText("Sem dados")).toBeVisible();

    await idInput.fill("mem-auth");
    await authenticatedPage.getByRole("button", { name: /Explorar/ }).click();

    const container = authenticatedPage.getByTestId("graph-ego-canvas");
    await expect(container.locator("canvas")).toBeVisible();
  });

  test("ego graph: URL param pre-loads the seed", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(`${ROUTES.graphEgo}?id=mem-auth`);

    const idInput = authenticatedPage.getByPlaceholder("UUID da memória-semente");
    await expect(idInput).toHaveValue("mem-auth");

    const container = authenticatedPage.getByTestId("graph-ego-canvas");
    await expect(container.locator("canvas")).toBeVisible();
  });

  test("timeline graph: renders SVG with markers and supersedes arrow", async ({ authenticatedPage }) => {
    const sidebar = new Sidebar(authenticatedPage);
    await sidebar.navigateTo("Grafo Temporal");
    await expect(authenticatedPage).toHaveURL(/\/app\/graph\/timeline/);

    await expect(authenticatedPage.getByRole("heading", { name: "Grafo Temporal" })).toBeVisible();

    for (const label of ["24h", "7d", "30d"]) {
      await expect(authenticatedPage.getByRole("button", { name: label, exact: true })).toBeVisible();
    }

    const svg = authenticatedPage.getByTestId("timeline-svg");
    await expect(svg).toBeVisible();

    // At least the 2 memory markers from the mock
    await expect.poll(async () => svg.locator("circle").count()).toBeGreaterThanOrEqual(2);

    // Red SUPERSEDES arrow present
    const supArrow = svg.locator("path[marker-end*='sup-arrow']");
    await expect(supArrow.first()).toBeAttached();

    // Stats strip
    await expect(authenticatedPage.getByText(/\b2\s+memórias\b/)).toBeVisible();
    await expect(authenticatedPage.getByText(/\b1\s+substituições\b/)).toBeVisible();
  });

  test("timeline graph: clicking a marker opens detail panel", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.graphTimeline);

    const svg = authenticatedPage.getByTestId("timeline-svg");
    await expect(svg).toBeVisible();

    await svg.locator("circle").first().click({ force: true });

    await expect(authenticatedPage.getByText("Registrada em").first()).toBeVisible();
    await expect(authenticatedPage.getByRole("button", { name: /Abrir Ego-grafo/ })).toBeVisible();
  });

  test("timeline graph: switching range preset keeps the svg rendered", async ({ authenticatedPage }) => {
    await authenticatedPage.goto(ROUTES.graphTimeline);

    await authenticatedPage.getByRole("button", { name: "7d", exact: true }).click();

    await expect(authenticatedPage.getByText("Falha ao carregar timeline")).toHaveCount(0);
    await expect(authenticatedPage.getByTestId("timeline-svg")).toBeVisible();
  });
});
