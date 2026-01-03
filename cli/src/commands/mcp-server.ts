import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { patchMake, patchToText } from "diff-match-patch-es";
import * as fs from "node:fs";
import * as path from "node:path";
import * as os from "node:os";
import { PlanDocumentClient } from "../mcp/plan-document-client.js";

interface SessionInfo {
  session_id?: string;
  tool_use_id?: string;
}

// Read session_id and tool_use_id from file written by PreToolUse hook
function getSessionInfoFromFile(): SessionInfo {
  try {
    const sessionFile = path.join(os.homedir(), ".agentrace", "current-session.json");
    if (fs.existsSync(sessionFile)) {
      const content = fs.readFileSync(sessionFile, "utf-8");
      const data = JSON.parse(content);
      return {
        session_id: data.session_id,
        tool_use_id: data.tool_use_id,
      };
    }
  } catch {
    // Ignore errors, return empty object
  }
  return {};
}

// Tool schemas
const SearchPlansSchema = z.object({
  git_remote_url: z.string().optional().describe("Git remote URL to filter by project"),
  status: z.string().optional().describe("Comma-separated statuses to filter (e.g., 'planning,implementation')"),
  description: z.string().optional().describe("Partial match search on plan description"),
});

const ReadPlanSchema = z.object({
  id: z.string().describe("Plan document ID"),
});

const CreatePlanSchema = z.object({
  description: z.string().describe("Short description of the plan"),
  body: z.string().describe("Plan content in Markdown format"),
});

const UpdatePlanSchema = z.object({
  id: z.string().describe("Plan document ID"),
  body: z.string().describe("Updated plan content in Markdown format"),
});

const SetPlanStatusSchema = z.object({
  id: z.string().describe("Plan document ID"),
  status: z.enum(["scratch", "draft", "planning", "pending", "implementation", "complete"]).describe("New status for the plan"),
});

// Tool descriptions with usage guidance
const TOOL_DESCRIPTIONS = {
  search_plans: `Search plan documents with filtering options.

WHEN TO USE:
- When you need to check existing plans for the current repository
- When the user asks about available plans or implementation documents
- Before creating a new plan to avoid duplicates
- When searching for plans by status (e.g., find all plans in 'scratch' status)
- When searching for plans by description keyword`,

  read_plan: `Read a plan document by ID.

WHEN TO USE:
- When the user asks you to check or review a specific plan by ID
- When you need to understand an existing plan before making changes
- When the user references a plan ID in their request`,

  create_plan: `Create a new plan document to record implementation or design plans.

WHEN TO USE:
- ALWAYS use this when you create a design or implementation plan
- When entering plan mode and documenting your approach
- When the user asks you to save or persist a plan
- When planning significant features, refactoring, or architectural changes

The plan will be saved to Agentrace server and can be reviewed by the team.
The project is automatically determined from the session's git repository.`,

  update_plan: `Update an existing plan document.

WHEN TO USE:
- When the user asks you to modify a specific plan by ID
- When implementation details change and the plan needs updating
- When you need to add progress notes or completion status to a plan

Changes are tracked with diff patches for history.`,

  set_plan_status: `Set the status of a plan document.

WHEN TO USE:
- When transitioning a plan from planning to implementation phase
- When marking a plan as complete after finishing the work
- When the user explicitly asks to change the status of a plan

Available statuses:
- scratch: Initial rough notes, starting point for discussion with AI
- draft: Plan not yet fully considered (optional intermediate status)
- planning: Plan is being designed/refined through discussion
- pending: Waiting for approval or blocked
- implementation: Active development is in progress
- complete: The work described in the plan is finished

BASIC FLOW: scratch → planning → implementation → complete
(draft and pending are optional auxiliary statuses)

STATUS TRANSITION GUIDELINES:
- scratch → planning: When you read a scratch plan (usually written by human), review its content and rewrite it into a more concrete plan, then change status to planning
- planning → implementation: When the plan is finalized after discussion, change status to implementation before starting work
- implementation → complete: When all work described in the plan is finished, change status to complete

CAUTION:
- When a plan is in "implementation" status, someone else might already be working on it. Check with the team before starting work on such plans.`,
};

export async function mcpServerCommand(): Promise<void> {
  const server = new McpServer({
    name: "agentrace",
    version: "1.0.0",
    description: `Agentrace Plan Document Management Server.

This server provides tools to manage implementation and design plans.
Plans are stored on the Agentrace server and can be reviewed by the team.

IMPORTANT GUIDELINES:
- When you create a design or implementation plan, ALWAYS save it using create_plan
- When the user asks you to check or modify a plan by ID, use the appropriate tool
- Plans help track what you're working on and enable team collaboration`,
  });

  let client: PlanDocumentClient | null = null;

  function getClient(): PlanDocumentClient {
    if (!client) {
      client = new PlanDocumentClient();
    }
    return client;
  }

  // search_plans tool
  server.tool(
    "search_plans",
    TOOL_DESCRIPTIONS.search_plans,
    SearchPlansSchema.shape,
    async (args) => {
      try {
        const plans = await getClient().searchPlans({
          gitRemoteUrl: args.git_remote_url,
          status: args.status,
          description: args.description,
        });

        if (plans.length === 0) {
          return {
            content: [
              {
                type: "text" as const,
                text: "No plans found matching the search criteria.",
              },
            ],
          };
        }

        const planList = plans.map((plan) => ({
          id: plan.id,
          description: plan.description,
          status: plan.status,
          updated_at: plan.updated_at,
          collaborators: plan.collaborators.map((c) => c.display_name).join(", "),
        }));

        return {
          content: [
            {
              type: "text" as const,
              text: JSON.stringify(planList, null, 2),
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: "text" as const,
              text: `Error: ${error instanceof Error ? error.message : String(error)}`,
            },
          ],
          isError: true,
        };
      }
    }
  );

  // read_plan tool
  server.tool(
    "read_plan",
    TOOL_DESCRIPTIONS.read_plan,
    ReadPlanSchema.shape,
    async (args) => {
      try {
        const plan = await getClient().getPlan(args.id);

        return {
          content: [
            {
              type: "text" as const,
              text: `# ${plan.description}\n\nStatus: ${plan.status}\n\n${plan.body}`,
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: "text" as const,
              text: `Error: ${error instanceof Error ? error.message : String(error)}`,
            },
          ],
          isError: true,
        };
      }
    }
  );

  // create_plan tool
  server.tool(
    "create_plan",
    TOOL_DESCRIPTIONS.create_plan,
    CreatePlanSchema.shape,
    async (args) => {
      try {
        // Read session_id and tool_use_id from file written by PreToolUse hook
        const sessionInfo = getSessionInfoFromFile();

        const plan = await getClient().createPlan({
          description: args.description,
          body: args.body,
          claude_session_id: sessionInfo.session_id,
          tool_use_id: sessionInfo.tool_use_id,
        });

        return {
          content: [
            {
              type: "text" as const,
              text: `Plan created successfully.\n\nID: ${plan.id}\nDescription: ${plan.description}`,
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: "text" as const,
              text: `Error: ${error instanceof Error ? error.message : String(error)}`,
            },
          ],
          isError: true,
        };
      }
    }
  );

  // update_plan tool
  server.tool(
    "update_plan",
    TOOL_DESCRIPTIONS.update_plan,
    UpdatePlanSchema.shape,
    async (args) => {
      try {
        // Read session_id and tool_use_id from file written by PreToolUse hook
        const sessionInfo = getSessionInfoFromFile();

        // Get current plan to compute patch
        const currentPlan = await getClient().getPlan(args.id);

        // Compute patch using diff-match-patch
        const patches = patchMake(currentPlan.body, args.body);
        const patchText = patchToText(patches);

        const plan = await getClient().updatePlan(args.id, {
          body: args.body,
          patch: patchText,
          claude_session_id: sessionInfo.session_id,
          tool_use_id: sessionInfo.tool_use_id,
        });

        return {
          content: [
            {
              type: "text" as const,
              text: `Plan updated successfully.\n\nID: ${plan.id}\nDescription: ${plan.description}`,
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: "text" as const,
              text: `Error: ${error instanceof Error ? error.message : String(error)}`,
            },
          ],
          isError: true,
        };
      }
    }
  );

  // set_plan_status tool
  server.tool(
    "set_plan_status",
    TOOL_DESCRIPTIONS.set_plan_status,
    SetPlanStatusSchema.shape,
    async (args) => {
      try {
        const plan = await getClient().setStatus(args.id, args.status);

        return {
          content: [
            {
              type: "text" as const,
              text: `Plan status updated successfully.\n\nID: ${plan.id}\nDescription: ${plan.description}\nStatus: ${plan.status}`,
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: "text" as const,
              text: `Error: ${error instanceof Error ? error.message : String(error)}`,
            },
          ],
          isError: true,
        };
      }
    }
  );

  // Start the server with stdio transport
  const transport = new StdioServerTransport();
  await server.connect(transport);
}
