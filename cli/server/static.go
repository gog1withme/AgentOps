package server

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func dashboardStatic(root string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/proxy") || strings.HasPrefix(path, "/mcp") {
			return c.Next()
		}

		rel := strings.TrimPrefix(path, "/")
		rel = strings.ReplaceAll(rel, "..", "")

		candidates := []string{}
		if rel == "" {
			candidates = append(candidates, "index.html")
		} else {
			candidates = append(candidates,
				filepath.Join(rel, "index.html"),
				rel,
			)
		}

		for _, candidate := range candidates {
			full := filepath.Join(root, filepath.FromSlash(candidate))
			if info, err := os.Stat(full); err == nil && !info.IsDir() {
				return sendStaticFile(c, full)
			}
		}

		return sendStaticFile(c, filepath.Join(root, "index.html"))
	}
}

func sendStaticFile(c *fiber.Ctx, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	switch filepath.Ext(path) {
	case ".html":
		c.Type("html")
	case ".js":
		c.Type("javascript")
	case ".css":
		c.Type("css")
	case ".json":
		c.Type("json")
	}
	return c.Send(data)
}
