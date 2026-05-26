# Changelog

All notable changes to AgentOps are documented in this file.

## [1.0.1] - 2026-05-26

Patch release focused on metric accuracy, context intelligence, and self-update.

> Full release notes: [docs/releases/v1.0.1.md](docs/releases/v1.0.1.md)

### Added

- **`agentops upgrade`** — Self-update from GitHub Releases with checksum verification (`--check` for dry-run)
- **`agentops context summary`** — Terminal summary of duplicate and noisy context
- **`/api/context/analysis`** — Context Inspector backend: duplicate file detection, noisy context callouts, session summary cards
- **Configurable model pricing** — Bundled defaults + user overrides at `~/.agentops/pricing.json`; surfaced in `agentops doctor`

### Fixed

- **MCP p95 latency** — Now computed from a rolling window of the last 100 calls (previously stored latest-call latency mislabeled as p95)
- **Richer `agentops diff`** — Per-type event counts, file overlap, model breakdown, and efficiency delta between sessions

### Changed

- Context Inspector dashboard shows summary cards and actionable callouts when duplicate or low-efficiency context is detected

## [1.0.0] - 2026-05-25

First public release of AgentOps — local-first observability for AI coding agents.

> Full release notes with use cases and screenshots: [docs/releases/v1.0.0.md](docs/releases/v1.0.0.md)

### Added

- **CLI**: `agentops init`, `dev`, `env`, `version`, `doctor`, `budget`, `restore`, `trace`, `replay`, `blame`, `prompt`, `mcp`, `alerts`, `sec`
- **Data collection**: OpenAI-compatible LLM proxy, shell hooks, filesystem watcher, MCP monitoring
- **Privacy**: Pre-storage scrubber with configurable patterns
- **Budgets**: Session cost/tool/LLM limits with alert, pause, and kill actions
- **Storage**: DuckDB (macOS/Linux) and SQLite (Windows) event store
- **Dashboard**: Live activity feed, traces with replay and restore, cost & efficiency charts, AI blame, MCP health, prompt versions, context inspector, security center, budget banner
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

[1.0.1]: https://github.com/gog1withme/AgentOps/releases/tag/v1.0.1
[1.0.0]: https://github.com/gog1withme/AgentOps/releases/tag/v1.0.0
