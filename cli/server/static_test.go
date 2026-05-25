package server

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestDashboardStaticTracesIndex(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("root"), 0o644); err != nil {
		t.Fatal(err)
	}
	tracesDir := filepath.Join(root, "traces")
	if err := os.MkdirAll(tracesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tracesDir, "index.html"), []byte("traces-page"), 0o644); err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use(dashboardStatic(root))

	req := httptest.NewRequest("GET", "/traces/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "traces-page" {
		t.Fatalf("expected traces index, got %q", string(body))
	}
}

func TestDashboardStaticSPAFallback(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("spa-root"), 0o644); err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use(dashboardStatic(root))

	req := httptest.NewRequest("GET", "/unknown/route/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "spa-root" {
		t.Fatalf("expected SPA fallback, got %q", string(body))
	}
}

func TestDashboardStaticSkipsMCPPrefix(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("root"), 0o644); err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use(dashboardStatic(root))
	app.Get("/mcp/test", func(c *fiber.Ctx) error {
		return c.SendString("mcp-ok")
	})

	req := httptest.NewRequest("GET", "/mcp/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "mcp-ok" {
		t.Fatalf("expected mcp route, got %q", string(body))
	}
}
