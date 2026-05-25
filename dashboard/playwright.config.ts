import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 30000,
  use: {
    baseURL: "http://127.0.0.1:4318",
    headless: true,
  },
  webServer: {
    command: "npx --yes serve out -l 4318",
    url: "http://127.0.0.1:4318",
    reuseExistingServer: !process.env.CI,
    cwd: ".",
  },
});
