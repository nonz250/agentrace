import { loadConfig } from "../config/manager.js";

export type PlanDocumentStatus = "draft" | "planning" | "pending" | "implementation" | "complete";

export interface PlanDocument {
  id: string;
  description: string;
  body: string;
  git_remote_url: string;
  status: PlanDocumentStatus;
  collaborators: {
    id: string;
    display_name: string;
  }[];
  created_at: string;
  updated_at: string;
}

export interface PlanDocumentEvent {
  id: string;
  plan_document_id: string;
  session_id: string | null;
  user_id: string | null;
  user_name: string | null;
  patch: string;
  created_at: string;
}

export interface ListPlansResponse {
  plans: PlanDocument[];
}

export interface ListEventsResponse {
  events: PlanDocumentEvent[];
}

export interface CreatePlanRequest {
  description: string;
  body: string;
  git_remote_url: string;
  session_id?: string;
}

export interface UpdatePlanRequest {
  description?: string;
  body?: string;
  patch?: string;
  session_id?: string;
}

export class PlanDocumentClient {
  private serverUrl: string;
  private apiKey: string;

  constructor() {
    const config = loadConfig();
    if (!config) {
      throw new Error("Agentrace is not configured. Run 'npx agentrace init' first.");
    }
    this.serverUrl = config.server_url;
    this.apiKey = config.api_key;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const url = `${this.serverUrl}${path}`;
    const headers: Record<string, string> = {
      "Authorization": `Bearer ${this.apiKey}`,
      "Content-Type": "application/json",
    };

    const response = await fetch(url, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API request failed: ${response.status} ${errorText}`);
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return undefined as T;
    }

    return response.json() as Promise<T>;
  }

  async listPlans(gitRemoteUrl?: string): Promise<PlanDocument[]> {
    let path = "/api/plans";
    if (gitRemoteUrl) {
      path += `?git_remote_url=${encodeURIComponent(gitRemoteUrl)}`;
    }
    const response = await this.request<ListPlansResponse>("GET", path);
    return response.plans;
  }

  async getPlan(id: string): Promise<PlanDocument> {
    return this.request<PlanDocument>("GET", `/api/plans/${id}`);
  }

  async getPlanEvents(id: string): Promise<PlanDocumentEvent[]> {
    const response = await this.request<ListEventsResponse>(
      "GET",
      `/api/plans/${id}/events`
    );
    return response.events;
  }

  async createPlan(req: CreatePlanRequest): Promise<PlanDocument> {
    return this.request<PlanDocument>("POST", "/api/plans", req);
  }

  async updatePlan(id: string, req: UpdatePlanRequest): Promise<PlanDocument> {
    return this.request<PlanDocument>("PATCH", `/api/plans/${id}`, req);
  }

  async deletePlan(id: string): Promise<void> {
    await this.request<void>("DELETE", `/api/plans/${id}`);
  }

  async setStatus(id: string, status: PlanDocumentStatus): Promise<PlanDocument> {
    return this.request<PlanDocument>("PATCH", `/api/plans/${id}/status`, { status });
  }
}
