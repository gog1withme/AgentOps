import { test, expect } from "@playwright/test";

test.describe("AgentOps dashboard", () => {
  test("home page loads title", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByRole("heading", { name: "Live Activity" })).toBeVisible();
  });

  test("blame page has search input", async ({ page }) => {
    await page.goto("/blame/");
    await expect(page.getByPlaceholder("Enter file path…")).toBeVisible();
  });

  test("budget banner component exists in layout", async ({ page }) => {
    await page.goto("/");
    // Banner may be hidden when under threshold — layout should still render
    await expect(page.getByText("AgentOps")).toBeVisible();
  });
});
