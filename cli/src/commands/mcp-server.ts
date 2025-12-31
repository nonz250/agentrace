import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { patchMake, patchToText } from "diff-match-patch-es";
import { PlanDocumentClient } from "../mcp/plan-document-client.js";

// Tool schemas
const ListPlansSchema = z.object({
  git_remote_url: z.string().describe("Git remote URL to filter plans"),
});

const ReadPlanSchema = z.object({
  id: z.string().describe("Plan document ID"),
});

const CreatePlanSchema = z.object({
  description: z.string().describe("Short description of the plan"),
  body: z.string().describe("Plan content in Markdown format"),
  git_remote_url: z.string().describe("Git remote URL of the repository"),
  session_id: z.string().optional().describe("Claude Code session ID (optional)"),
});

const UpdatePlanSchema = z.object({
  id: z.string().describe("Plan document ID"),
  body: z.string().describe("Updated plan content in Markdown format"),
  session_id: z.string().optional().describe("Claude Code session ID (optional)"),
});

// Tool descriptions with usage guidance
const TOOL_DESCRIPTIONS = {
  list_plans: `List plan documents for a repository.

WHEN TO USE:
- When you need to check existing plans for the current repository
- When the user asks about available plans or implementation documents
- Before creating a new plan to avoid duplicates`,

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

The plan will be saved to Agentrace server and can be reviewed by the team.`,

  update_plan: `Update an existing plan document.

WHEN TO USE:
- When the user asks you to modify a specific plan by ID
- When implementation details change and the plan needs updating
- When you need to add progress notes or completion status to a plan
- When marking a plan as completed after finishing implementation

Changes are tracked with diff patches for history.`,
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

  // list_plans tool
  server.tool(
    "list_plans",
    TOOL_DESCRIPTIONS.list_plans,
    ListPlansSchema.shape,
    async (args) => {
      try {
        const plans = await getClient().listPlans(args.git_remote_url);

        if (plans.length === 0) {
          return {
            content: [
              {
                type: "text" as const,
                text: "No plans found for this repository.",
              },
            ],
          };
        }

        const planList = plans.map((plan) => ({
          id: plan.id,
          description: plan.description,
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
              text: `# ${plan.description}\n\n${plan.body}`,
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
        const plan = await getClient().createPlan({
          description: args.description,
          body: args.body,
          git_remote_url: args.git_remote_url,
          session_id: args.session_id,
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
        // Get current plan to compute patch
        const currentPlan = await getClient().getPlan(args.id);

        // Compute patch using diff-match-patch
        const patches = patchMake(currentPlan.body, args.body);
        const patchText = patchToText(patches);

        const plan = await getClient().updatePlan(args.id, {
          body: args.body,
          patch: patchText,
          session_id: args.session_id,
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

  // Start the server with stdio transport
  const transport = new StdioServerTransport();
  await server.connect(transport);
}
