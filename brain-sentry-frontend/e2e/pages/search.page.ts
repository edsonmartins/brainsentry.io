import { type Page, type Locator } from "@playwright/test";

export class SearchPage {
  readonly page: Page;
  readonly searchInput: Locator;
  readonly searchButton: Locator;
  readonly advancedToggle: Locator;
  readonly advancedInfo: Locator;
  readonly categoryFilter: Locator;
  readonly importanceFilter: Locator;
  readonly clearFiltersButton: Locator;
  readonly resultsCount: Locator;
  readonly noResults: Locator;
  readonly initialState: Locator;
  readonly pageSizeSelect: Locator;
  readonly suggestions: Locator;

  constructor(page: Page) {
    this.page = page;
    this.searchInput = page.getByPlaceholder(/Digite sua dúvida/i);
    this.searchButton = page.getByRole("button", { name: "Buscar" }).first();
    this.advancedToggle = page.getByRole("button", { name: /Busca Avançada/i });
    this.advancedInfo = page.getByText(/Retrieval Planner/i);
    this.categoryFilter = page.locator("text=Categoria").locator("..");
    this.importanceFilter = page.locator("text=Importância").locator("..");
    this.clearFiltersButton = page.getByRole("button", { name: /Limpar filtros/i });
    this.resultsCount = page.locator("text=/\\d+ resultados encontrados/");
    this.noResults = page.getByText("Nenhum resultado encontrado");
    this.initialState = page.getByText("Busque em suas memórias");
    this.pageSizeSelect = page.locator("select");
    this.suggestions = page.getByRole("button", { name: /Spring Boot|React hooks|API REST|Docker/i });
  }

  async goto() {
    await this.page.goto("/app/search");
  }

  async search(query: string) {
    await this.searchInput.fill(query);
    await this.searchButton.click();
  }
}
