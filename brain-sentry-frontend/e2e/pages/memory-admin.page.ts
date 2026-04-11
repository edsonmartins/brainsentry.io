import { type Page, type Locator } from "@playwright/test";

export class MemoryAdminPage {
  readonly page: Page;
  readonly searchInput: Locator;
  readonly newMemoryButton: Locator;
  readonly importButton: Locator;
  readonly exportButton: Locator;
  readonly memoryCount: Locator;
  readonly memoryGrid: Locator;
  readonly dialogOverlay: Locator;
  readonly dialogContent: Locator;
  readonly prevPageButton: Locator;
  readonly nextPageButton: Locator;
  readonly emptySearchState: Locator;
  readonly emptyState: Locator;
  readonly loadingSpinner: Locator;

  constructor(page: Page) {
    this.page = page;
    this.searchInput = page.getByPlaceholder("Buscar memórias...");
    this.newMemoryButton = page.getByRole("button", { name: /Nova Memória/i });
    this.importButton = page.getByRole("button", { name: /Importar/i });
    this.exportButton = page.getByRole("button", { name: /Exportar/i });
    this.memoryCount = page.locator("text=/\\d+ memória/");
    this.memoryGrid = page.locator(".grid");
    // Custom dialog (no role=dialog): overlay container with backdrop
    this.dialogOverlay = page.locator("div.fixed.inset-0.z-50");
    this.dialogContent = page.locator("div.fixed.inset-0.z-50 div.bg-background.rounded-lg");
    this.prevPageButton = page.getByRole("button", { name: /Anterior/i });
    this.nextPageButton = page.getByRole("button", { name: /Próxima/i });
    this.emptySearchState = page.getByText(/Nenhuma memória encontrada para/);
    this.emptyState = page.getByText("Nenhuma memória cadastrada.");
    this.loadingSpinner = page.getByText("Carregando...");
  }

  async goto() {
    await this.page.goto("/app/memories");
  }

  async search(query: string) {
    await this.searchInput.fill(query);
    await this.page.waitForTimeout(1000);
    await this.page.waitForLoadState("networkidle");
  }

  async openCreateDialog() {
    await this.newMemoryButton.click();
    await this.dialogContent.waitFor({ state: "visible", timeout: 10_000 });
  }
}
