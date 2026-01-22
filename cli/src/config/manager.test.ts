import { describe, it, expect, beforeEach, afterEach } from "vitest";
import * as fs from "node:fs";
import * as path from "node:path";
import * as os from "node:os";
import {
  getConfigPath,
  loadLocalConfig,
  saveLocalConfig,
  deleteLocalConfig,
  getLocalConfigPath,
  loadConfigWithFallback,
  type AgentraceConfig,
} from "./manager.js";

describe("config/manager", () => {
  const testConfig: AgentraceConfig = {
    server_url: "http://localhost:8080",
    api_key: "agtr_test_key",
  };

  describe("global config", () => {
    it("getConfigPath returns expected path", () => {
      expect(getConfigPath()).toBe(
        path.join(os.homedir(), ".agentrace", "config.json")
      );
    });
  });

  describe("local config", () => {
    let tempProjectDir: string;

    beforeEach(() => {
      tempProjectDir = fs.mkdtempSync(path.join(os.tmpdir(), "agentrace-project-"));
    });

    afterEach(() => {
      if (fs.existsSync(tempProjectDir)) {
        fs.rmSync(tempProjectDir, { recursive: true });
      }
    });

    it("getLocalConfigPath returns correct path", () => {
      const configPath = getLocalConfigPath(tempProjectDir);
      expect(configPath).toBe(path.join(tempProjectDir, ".agentrace", "config.json"));
    });

    it("saveLocalConfig creates config file", () => {
      saveLocalConfig(tempProjectDir, testConfig);

      const configPath = getLocalConfigPath(tempProjectDir);
      expect(fs.existsSync(configPath)).toBe(true);

      const content = JSON.parse(fs.readFileSync(configPath, "utf-8"));
      expect(content).toEqual(testConfig);
    });

    it("loadLocalConfig returns config when exists", () => {
      saveLocalConfig(tempProjectDir, testConfig);

      const loaded = loadLocalConfig(tempProjectDir);
      expect(loaded).toEqual(testConfig);
    });

    it("loadLocalConfig returns null when config does not exist", () => {
      const loaded = loadLocalConfig(tempProjectDir);
      expect(loaded).toBeNull();
    });

    it("deleteLocalConfig removes config directory", () => {
      saveLocalConfig(tempProjectDir, testConfig);
      const configDir = path.join(tempProjectDir, ".agentrace");
      expect(fs.existsSync(configDir)).toBe(true);

      const result = deleteLocalConfig(tempProjectDir);
      expect(result).toBe(true);
      expect(fs.existsSync(configDir)).toBe(false);
    });

    it("deleteLocalConfig returns false when no config exists", () => {
      const result = deleteLocalConfig(tempProjectDir);
      expect(result).toBe(false);
    });
  });

  describe("loadConfigWithFallback", () => {
    let tempProjectDir: string;

    const localConfig: AgentraceConfig = {
      server_url: "http://local:8080",
      api_key: "test_local_key",
    };

    beforeEach(() => {
      tempProjectDir = fs.mkdtempSync(path.join(os.tmpdir(), "agentrace-project-"));
    });

    afterEach(() => {
      if (fs.existsSync(tempProjectDir)) {
        fs.rmSync(tempProjectDir, { recursive: true });
      }
    });

    it("returns local config when it exists", () => {
      saveLocalConfig(tempProjectDir, localConfig);

      const loaded = loadConfigWithFallback(tempProjectDir);
      expect(loaded).toEqual(localConfig);
    });

    it("prefers local config over global config", () => {
      // Save local config
      saveLocalConfig(tempProjectDir, localConfig);

      // loadConfigWithFallback should return local config
      const loaded = loadConfigWithFallback(tempProjectDir);
      expect(loaded).toEqual(localConfig);
    });
  });
});
