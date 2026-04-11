import { API_BASE, DEMO_EMAIL, DEMO_PASSWORD, DEFAULT_TENANT_ID } from "./constants";

export interface AuthResponse {
  token: string;
  user: { id: string; email: string; name?: string; tenantId?: string };
  tenantId?: string;
}

// Cache auth response to avoid rate limiting (429)
let cachedAuth: AuthResponse | null = null;
let cacheExpiry = 0;

export class ApiHelper {
  private token: string | null = null;

  async login(
    email = DEMO_EMAIL,
    password = DEMO_PASSWORD
  ): Promise<AuthResponse> {
    // Return cached auth if still valid (cache for 5 minutes)
    if (cachedAuth && Date.now() < cacheExpiry && email === DEMO_EMAIL) {
      this.token = cachedAuth.token;
      return cachedAuth;
    }

    const res = await this.fetchWithRetry(`${API_BASE}/v1/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Tenant-ID": DEFAULT_TENANT_ID,
      },
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) {
      throw new Error(`Login failed: ${res.status} ${await res.text()}`);
    }

    const data: AuthResponse = await res.json();
    this.token = data.token;

    // Cache for demo user
    if (email === DEMO_EMAIL) {
      cachedAuth = data;
      cacheExpiry = Date.now() + 5 * 60 * 1000;
    }

    return data;
  }

  async ensureDemoUser(): Promise<void> {
    const res = await fetch(`${API_BASE}/v1/auth/demo`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Tenant-ID": DEFAULT_TENANT_ID,
      },
    });
    if (!res.ok) {
      console.warn(`Demo user setup returned ${res.status}`);
    }
  }

  async createMemory(content: string): Promise<{ id: string }> {
    if (!this.token) await this.login();
    const res = await fetch(`${API_BASE}/v1/memories`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.token}`,
        "X-Tenant-ID": DEFAULT_TENANT_ID,
      },
      body: JSON.stringify({ content }),
    });
    if (!res.ok) {
      throw new Error(`Create memory failed: ${res.status}`);
    }
    return res.json();
  }

  async deleteMemory(id: string): Promise<void> {
    if (!this.token) await this.login();
    await fetch(`${API_BASE}/v1/memories/${id}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${this.token}`,
        "X-Tenant-ID": DEFAULT_TENANT_ID,
      },
    });
  }

  private async fetchWithRetry(
    url: string,
    init: RequestInit,
    retries = 3
  ): Promise<Response> {
    for (let i = 0; i < retries; i++) {
      const res = await fetch(url, init);
      if (res.status === 429 && i < retries - 1) {
        // Wait with exponential backoff
        await new Promise((r) => setTimeout(r, 1000 * (i + 1)));
        continue;
      }
      return res;
    }
    return fetch(url, init);
  }
}
