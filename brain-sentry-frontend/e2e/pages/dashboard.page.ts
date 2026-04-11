import { type Page, type Locator } from "@playwright/test";

export class DashboardPage {
  readonly page: Page;
  readonly totalMemoriesCard: Locator;
  readonly categoriesCard: Locator;
  readonly criticalCard: Locator;
  readonly active24hCard: Locator;
  readonly newMemoryButton: Locator;
  readonly searchButton: Locator;
  readonly recentMemoriesSection: Locator;
  readonly categoryBreakdown: Locator;

  constructor(page: Page) {
    this.page = page;
    this.totalMemoriesCard = page.getByText("Total de Memórias");
    this.categoriesCard = page.getByText("Categorias");
    this.criticalCard = page.getByText("Memórias Críticas");
    this.active24hCard = page.getByText("Ativas 24h");
    this.newMemoryButton = page.getByRole("button", { name: /Nova Memória/i });
    this.searchButton = page.getByRole("button", { name: /Buscar/i }).first();
    this.recentMemoriesSection = page.getByText("Memórias Recentes");
    this.categoryBreakdown = page.getByText("Por Categoria");
  }

  async goto() {
    await this.page.goto("/app/dashboard");
  }
}
