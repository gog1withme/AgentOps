# AgentOps

Local-first observability for AI coding agents — passive hooks, privacy scrubbing, budgets, and a beautiful dashboard.

[![Release](https://img.shields.io/github/v/release/gog1withme/AgentOps)](https://github.com/gog1withme/AgentOps/releases)

## What's new

**[v1.0.0 (2026-05-25)](docs/releases/v1.0.0.md)** — First public release: live agent feed, LLM cost tracking, trace replay with one-click restore, MCP health, AI blame, and privacy scrubbing. [Full release notes →](docs/releases/v1.0.0.md)

![AgentOps live activity feed showing real-time Cursor file edits](docs/assets/v1.0.0/live-activity-feed.png)

## Install

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/gog1withme/AgentOps/main/scripts/install.sh | sh
export PATH="$HOME/.agentops/bin:$PATH"
agentops init
agentops env          # run the printed command in this shell
agentops dev
```

**macOS (Homebrew)**

```bash
brew tap gog1withme/homebrew-agentops
brew install agentops
agentops init
agentops env
agentops dev
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/gog1withme/AgentOps/main/scripts/install.ps1 | iex
$env:Path = "$env:USERPROFILE\.agentops\bin;$env:Path"
agentops init
agentops env
agentops dev
```

Download binaries manually from [GitHub Releases](https://github.com/gog1withme/AgentOps/releases).

After `agentops init`, run `agentops env` and **restart Cursor** so it picks up `OPENAI_BASE_URL` / `ANTHROPIC_BASE_URL`.

## Quick start (from source)

**Windows (PowerShell)**

```powershell
.\build.ps1 build
.\bin\agentops.exe init
.\bin\agentops.exe env
.\bin\agentops.exe dev
```

**macOS / Linux**

```bash
make build
./bin/agentops init
source ~/.agentops/env.sh
./bin/agentops dev
```

## Build dashboard

```powershell
# Windows
.\build.ps1 dashboard-install
.\build.ps1 dashboard-build
```

```bash
# macOS / Linux
cd dashboard && npm install && npm run build
```

The Go server serves static files from `dashboard/out` when present (or from `~/.agentops/dashboard/out` after install).

## Commands

| Command | Description |
|---------|-------------|
| `agentops version` | Print installed version |
| `agentops init` | Detect agents, install hooks, activate scrubber |
| `agentops env` | Print command to load LLM proxy environment variables |
| `agentops dev` | Start collector + dashboard |
| `agentops budget set/status` | Session spend limits |
| `agentops restore` | Time-travel file restore |
| `agentops trace list` | List sessions |
| `agentops replay <id>` | Timeline replay |
| `agentops blame <file>` | AI edit attribution |
| `agentops prompt list/diff` | Prompt versioning |
| `agentops mcp list` | MCP health |
| `agentops alerts` | Active alerts |
| `agentops doctor` | Health checks |
| `agentops sec scan` | Security scan |

## Architecture

Go CLI/daemon → scrubber → DuckDB (macOS/Linux) or SQLite (Windows) → HTTP API/SSE → Next.js dashboard

See [context.md](context.md) and [development.context](development.context) for full spec.

Dogfood validation checklist: [docs/dogfood-checklist.md](docs/dogfood-checklist.md).

## Environment

- `AGENTOPS_PORT` — dashboard/API port (default 4318); written to `~/.agentops/env.ps1` or `env.sh` on init
- `OPENAI_BASE_URL` / `ANTHROPIC_BASE_URL` — set via `agentops env` so Cursor routes LLM calls through `/proxy`
- `AGENTOPS_DATA_DIR` — override data directory
- `AGENTOPS_REDACT_CONTENT=1` — redact all content fields
- `AGENTOPS_NO_SNAPSHOTS=1` — disable snapshots
- `AGENTOPS_NO_BROWSER=1` — skip opening browser on dev

## License

MIT — see [LICENSE](LICENSE).
