import { fetch } from "undici";
import { loadConfigWithFallback } from "../config/manager.js";
import { createDispatcher } from "../utils/proxy.js";

export type PlanDocumentStatus = "scratch" | "draft" | "planning" | "pending" | "ready" | "implementation" | "complete";

export interface Project {
  id: string;
  canonical_git_repository: string;
}

export interface PlanDocument {
  id: string;
  description: string;
  body: string;
  project: Project | null;
  status: PlanDocumentStatus;
  collaborators: {
    id: string;
    display_name: string;
  }[];
  url?: string;
  created_at: string;
  updated_at: string;
  active_comments_count: number;
}

export interface PlanDocumentEvent {
  id: string;
  plan_document_id: string;
  claude_session_id: string | null;
  user_id: string | null;
  user_name: string | null;
  patch: string;
  message: string;
  created_at: string;
}

export interface ListPlansResponse {
  plans: PlanDocument[];
  next_cursor?: string;
}

export interface SearchPlansParams {
  gitRemoteUrl?: string;
  status?: string;
  description?: string;
}

export interface ListEventsResponse {
  events: PlanDocumentEvent[];
}

export interface CreatePlanRequest {
  description: string;
  body: string;
  claude_session_id?: string;
  tool_use_id?: string;
}

export interface UpdatePlanRequest {
  description?: string;
  body?: string;
  patch?: string;
  message?: string;
  claude_session_id?: string;
  tool_use_id?: string;
}

// Comment Thread types
export type PlanCommentThreadStatus = "active" | "resolved" | "outdated";

export interface CommentPosition {
  start_offset: number;
  end_offset: number;
  start_line: number;
  start_column: number;
  end_line: number;
  end_column: number;
  found: boolean;
}

export interface PlanCommentMessage {
  id: string;
  thread_id: string;
  user_id: string;
  user_name: string;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface PlanCommentThread {
  id: string;
  plan_document_id: string;
  target_text: string;
  status: PlanCommentThreadStatus;
  position: CommentPosition | null;
  messages: PlanCommentMessage[];
  created_at: string;
  updated_at: string;
}

export interface ListThreadsResponse {
  threads: PlanCommentThread[];
}

export interface CreateThreadRequest {
  target_text: string;
  context_before: string;
  context_after: string;
  content: string;
}

export interface AddMessageRequest {
  content: string;
}

export class PlanDocumentClient {
  private serverUrl: string;
  private apiKey: string;

  constructor(projectDir?: string) {
    const config = loadConfigWithFallback(projectDir);
    if (!config) {
      throw new Error("AgenTrace is not configured. Run 'npx agentrace init' first.");
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
      dispatcher: createDispatcher(),
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

  async searchPlans(params: SearchPlansParams = {}): Promise<PlanDocument[]> {
    const searchParams = new URLSearchParams();
    if (params.gitRemoteUrl) {
      searchParams.set("git_remote_url", params.gitRemoteUrl);
    }
    if (params.status) {
      searchParams.set("status", params.status);
    }
    if (params.description) {
      searchParams.set("description", params.description);
    }

    const query = searchParams.toString();
    const path = query ? `/api/plans?${query}` : "/api/plans";
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

  async setStatus(id: string, status: PlanDocumentStatus, message?: string): Promise<PlanDocument> {
    return this.request<PlanDocument>("PATCH", `/api/plans/${id}/status`, { status, message });
  }

  // Comment Thread methods
  async getThreads(planId: string, status?: PlanCommentThreadStatus | "all"): Promise<PlanCommentThread[]> {
    const params = new URLSearchParams();
    if (status) {
      params.set("status", status);
    }
    const query = params.toString();
    const path = `/api/plans/${planId}/comments${query ? `?${query}` : ""}`;
    const response = await this.request<ListThreadsResponse>("GET", path);
    return response.threads;
  }

  async createThread(planId: string, req: CreateThreadRequest): Promise<PlanCommentThread> {
    return this.request<PlanCommentThread>("POST", `/api/plans/${planId}/comments`, req);
  }

  async resolveThread(planId: string, threadId: string): Promise<PlanCommentThread> {
    return this.request<PlanCommentThread>("POST", `/api/plans/${planId}/comments/${threadId}/resolve`);
  }

  async addMessage(planId: string, threadId: string, req: AddMessageRequest): Promise<PlanCommentMessage> {
    return this.request<PlanCommentMessage>("POST", `/api/plans/${planId}/comments/${threadId}/messages`, req);
  }
}
