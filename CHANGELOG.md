# Changelog

All notable changes to AgentOps are documented in this file.

## [1.0.0] - 2026-05-25

First public release of AgentOps — local-first observability for AI coding agents.

### Added

- **CLI**: `agentops init`, `dev`, `env`, `version`, `doctor`, `budget`, `restore`, `trace`, `replay`, `blame`, `prompt`, `mcp`, `alerts`, `sec`
- **Data collection**: OpenAI-compatible LLM proxy, shell hooks, filesystem watcher, MCP monitoring
- **Privacy**: Pre-storage scrubber with configurable patterns
- **Budgets**: Session cost/tool/LLM limits with alert, pause, and kill actions
- **Storage**: DuckDB (macOS/Linux) and SQLite (Windows) event store
- **Dashboard**: Live activity feed, token/cost tracking, trace replay, MCP health, budget banner
- **Distribution**: GitHub Releases, install scripts (`scripts/install.sh`, `scripts/install.ps1`), Homebrew tap support

### Install

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/gog1withme/AgentOps/main/scripts/install.sh | sh

# macOS (Homebrew)
brew tap gog1withme/homebrew-agentops
brew install agentops

# Windows (PowerShell)
irm https://raw.githubusercontent.com/gog1withme/AgentOps/main/scripts/install.ps1 | iex
```

[1.0.0]: https://github.com/gog1withme/AgentOps/releases/tag/v1.0.0
