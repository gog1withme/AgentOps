# How does passive interception work without SDK integration?

[← Developer Q&A](../README.md)

`architecture` · `proxy` · `hooks` · `cursor` · `claude` · `copilot` · `mcp`

**Current answer — v1.0.0** · 2026-05-25

AgentOps v1.0 uses four OS-level integration points — no SDK required:

| Layer | Mechanism | What it captures |
|-------|-----------|------------------|
| LLM calls | `OPENAI_BASE_URL` / `ANTHROPIC_BASE_URL` → local reverse proxy at `/proxy` | Tokens, cost, model, system prompt hash |
| Shell | `PROMPT_COMMAND` (bash/zsh) or PowerShell profile → POST `/api/ingest/shell` | Commands run in hooked terminal sessions |
| File edits | `fsnotify` watcher on the project directory | Write/create events with diffs |
| MCP | Patch `mcp.json` URLs to route through `/mcp/{server}` | Latency, errors, call volume (URL-based servers only) |

**Agent detection** uses lightweight heuristics: User-Agent on LLM requests (`cli/collector/proxy.go`) and running process names for file attribution (`cli/internal/platform/attribution.go`).

**Brittleness tradeoff:** We deliberately avoid deep editor integration. The proxy parses OpenAI and Anthropic JSON request/response shapes — if upstream APIs change, we update the parser, not your agent. Copilot is detected at init but may not route through the proxy unless it respects those env vars. MCP stdio/command servers are not proxied. Integrated terminals inside Cursor may not hit shell hooks unless they inherit a hooked shell profile.

This is the zero-friction onboarding bet: works immediately for Cursor and Claude Code when you run `agentops env` and restart the editor, at the cost of not having kernel-level provenance.
