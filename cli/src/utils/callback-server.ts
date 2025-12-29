import * as http from "node:http";
import * as crypto from "node:crypto";

export interface CallbackResult {
  apiKey: string;
}

export interface CallbackServerOptions {
  token: string;
  timeout: number; // milliseconds
}

export function getRandomPort(): number {
  // Use dynamic port range (49152-65535)
  return Math.floor(Math.random() * (65535 - 49152 + 1)) + 49152;
}

export function generateToken(): string {
  return crypto.randomUUID();
}

export function startCallbackServer(
  port: number,
  options: CallbackServerOptions
): Promise<CallbackResult> {
  return new Promise((resolve, reject) => {
    let resolved = false;

    const server = http.createServer((req, res) => {
      // Set CORS headers
      res.setHeader("Access-Control-Allow-Origin", "*");
      res.setHeader("Access-Control-Allow-Methods", "POST, OPTIONS");
      res.setHeader("Access-Control-Allow-Headers", "Content-Type");

      // Handle preflight
      if (req.method === "OPTIONS") {
        res.writeHead(204);
        res.end();
        return;
      }

      // Only accept POST to /callback
      if (req.method !== "POST" || req.url !== "/callback") {
        res.writeHead(404, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "Not found" }));
        return;
      }

      let body = "";
      req.on("data", (chunk) => {
        body += chunk.toString();
      });

      req.on("end", () => {
        try {
          const data = JSON.parse(body);

          // Validate token
          if (data.token !== options.token) {
            res.writeHead(401, { "Content-Type": "application/json" });
            res.end(JSON.stringify({ error: "Invalid token" }));
            return;
          }

          // Validate api_key
          if (!data.api_key || typeof data.api_key !== "string") {
            res.writeHead(400, { "Content-Type": "application/json" });
            res.end(JSON.stringify({ error: "Invalid api_key" }));
            return;
          }

          // Success response
          res.writeHead(200, { "Content-Type": "application/json" });
          res.end(JSON.stringify({ success: true }));

          resolved = true;
          server.close();
          resolve({ apiKey: data.api_key });
        } catch {
          res.writeHead(400, { "Content-Type": "application/json" });
          res.end(JSON.stringify({ error: "Invalid JSON" }));
        }
      });
    });

    server.on("error", (err) => {
      if (!resolved) {
        reject(err);
      }
    });

    // Timeout handling
    const timeoutId = setTimeout(() => {
      if (!resolved) {
        server.close();
        reject(new Error("Timeout: No callback received"));
      }
    }, options.timeout);

    server.on("close", () => {
      clearTimeout(timeoutId);
    });

    server.listen(port, "127.0.0.1", () => {
      // Server started successfully
    });
  });
}
