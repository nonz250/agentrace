import { defineConfig } from "tsup";
import pkg from "./package.json";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm"],
  target: "node18",
  outDir: "dist",
  clean: true,
  splitting: false,
  sourcemap: false,
  dts: false,
  define: {
    __CLI_VERSION__: JSON.stringify(pkg.version),
  },
  // Keep external dependencies (don't bundle them)
  external: [
    "@modelcontextprotocol/sdk",
    "commander",
    "diff-match-patch-es",
    "undici",
    "zod",
  ],
});
