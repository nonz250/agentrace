import { loadConfig } from "../config/manager.js";

export interface IngestPayload {
  session_id: string;
  transcript_lines: unknown[];
  cwd?: string;
  git_remote_url?: string;
  git_branch?: string;
}

export interface IngestResponse {
  ok: boolean;
  events_created?: number;
  error?: string;
}

export interface WebSessionResponse {
  url: string;
  expires_at: string;
}

function getBaseUrl(config: { server_url: string }): string {
  return config.server_url.replace(/\/+$/, '');
}

export async function sendIngest(
  payload: IngestPayload
): Promise<IngestResponse> {
  const config = loadConfig();
  if (!config) {
    return { ok: false, error: "Config not found" };
  }

  const url = `${getBaseUrl(config)}/api/ingest`;

  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${config.api_key}`,
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const text = await response.text();
      return { ok: false, error: `HTTP ${response.status}: ${text}` };
    }

    return (await response.json()) as IngestResponse;
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { ok: false, error: message };
  }
}

export async function createWebSession(): Promise<
  { ok: true; data: WebSessionResponse } | { ok: false; error: string }
> {
  const config = loadConfig();
  if (!config) {
    return { ok: false, error: "Config not found. Run 'agentrace init' first." };
  }

  const url = `${getBaseUrl(config)}/api/auth/web-session`;

  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${config.api_key}`,
      },
    });

    if (!response.ok) {
      const text = await response.text();
      return { ok: false, error: `HTTP ${response.status}: ${text}` };
    }

    const data = (await response.json()) as WebSessionResponse;
    return { ok: true, data };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { ok: false, error: message };
  }
}
