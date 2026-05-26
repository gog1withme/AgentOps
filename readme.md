# AgentOps

Local-first observability for AI coding agents — passive hooks, privacy scrubbing, budgets, and a beautiful dashboard.

[![Release](https://img.shields.io/github/v/release/gog1withme/AgentOps)](https://github.com/gog1withme/AgentOps/releases)

## What's new

**[v1.0.1 (2026-05-26)](docs/releases/v1.0.1.md)** — Honest MCP p95 metrics, configurable model pricing, Context Inspector callouts, `agentops upgrade`, richer `agentops diff`. [Release notes →](docs/releases/v1.0.1.md)

**[v1.0.0 (2026-05-25)](docs/releases/v1.0.0.md)** — First public release: live agent feed, LLM cost tracking, trace replay with one-click restore, MCP health, AI blame, and privacy scrubbing. [Full release notes →](docs/releases/v1.0.0.md)

![AgentOps live activity feed showing real-time Cursor file edits](docs/assets/v1.0.0/live-activity-feed.png)

## Why AgentOps?

AI agents edit your code, run shell commands, and call MCP tools — often while you are away. AgentOps is a local flight recorder and control panel: passive hooks, zero SDK, privacy scrubbing before storage, and a dashboard at `localhost:4318`.

## Use cases

### Coding with AI assistants

Works with Cursor, Claude Code, Copilot, and any OpenAI-compatible agent.

| Scenario | What AgentOps does | Try it |
|---|---|---|
| Audit what the agent changed during release prep | Live feed streams every `FILE_EDIT` with timestamp and agent name | `agentops dev` |
| Cap spend on a long refactor session | Budget limits + cost dashboard; banner at 80% of limit | `agentops budget set --cost 2.00 --action alert` |
| Find bloated `@` context wasting tokens | Context Inspector shows efficiency score per LLM call | Open [localhost:4318/context/](http://localhost:4318/context/) |
| Roll back a bad multi-file refactor | Trace replay with **Restore to here** on any event | `agentops restore --at "<timestamp>"` |
| Trace a bug to an AI edit vs human edit | AI Blame shows every agent-caused change with diff | `agentops blame src/auth.ts` |

### Supervising autonomous and background agents

When an agent runs unattended — overnight refactors, background Cursor tasks, or long MCP tool chains — AgentOps gives you guardrails and visibility without changing how the agent works.

| Scenario | How supervision works today |
|---|---|
| Agent runs while you are away | Set cost, tool, and LLM limits; dashboard warns at 80%; `pause` or `kill` the agent process at 100% |
| Watch an unattended agent from a second screen | Live feed SSE streams LLM calls, shell commands, file edits, and MCP calls in real time |
| Agent tries a destructive shell command | Security Center raises a critical `dangerous_shell` alert |
| Agent enters an expensive token spiral | Alerts on 100k+ prompt tokens and low efficiency scores |
| Post-mortem after a background run | Trace replay timeline + `agentops replay <id>` |
| Agent corrupted the repo | One-click **Restore to here** on any trace event |
| Long MCP tool chains degrade | MCP Health flags servers with p95 > 2s or error rate > 10% |

**Supervision setup**

```bash
agentops budget set --cost 5.00 --tools 100 --action pause
agentops dev
# Leave dashboard open at http://localhost:4318 — alerts appear in Security Center
```

Detailed examples with screenshots: [v1.0.0 release notes](docs/releases/v1.0.0.md)

## Roadmap

- **Near-term (1.x)** — Smart insights, configurable alerts, prompt evals, SDK integrations
- **Mid-term (1.x–2.x)** — Autonomous agent guardrails, cloud/CI agent tracing, multi-agent conflict detection
- **Long-term (2.x+)** — OpenTelemetry GenAI export, RAG pipeline tracing, compliance audit trails, team observability

See the full [roadmap](docs/ROADMAP.md) for planned features and industry direction.

**Developer Q&A:** Technical answers tagged by release version — [docs/qa/](docs/qa/). Superseded answers stay visible with strikethrough when behavior changes.

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
| `agentops diff <a> <b>` | Compare two sessions (counts, files, models, efficiency) |
| `agentops blame <file>` | AI edit attribution |
| `agentops prompt list/diff` | Prompt versioning |
| `agentops context summary` | Duplicate and noisy context analysis |
| `agentops mcp list` | MCP health |
| `agentops alerts` | Active alerts |
| `agentops doctor` | Health checks (includes pricing config) |
| `agentops upgrade` | Self-update from GitHub Releases |
| `agentops sec scan` | Security scan |

## Architecture

Go CLI/daemon → scrubber → DuckDB (macOS/Linux) or SQLite (Windows) → HTTP API/SSE → Next.js dashboard

See [context.md](context.md) and [development.context](development.context) for full spec. Public roadmap: [docs/ROADMAP.md](docs/ROADMAP.md).

Dogfood validation checklist: [docs/dogfood-checklist.md](docs/dogfood-checklist.md).

## Environment

- `AGENTOPS_PORT` — dashboard/API port (default 4318); written to `~/.agentops/env.ps1` or `env.sh` on init
- `OPENAI_BASE_URL` / `ANTHROPIC_BASE_URL` — set via `agentops env` so Cursor routes LLM calls through `/proxy`
- `AGENTOPS_DATA_DIR` — override data directory
- `AGENTOPS_PRICING` — override path to model pricing JSON (default `~/.agentops/pricing.json`)
- `AGENTOPS_REDACT_CONTENT=1` — redact all content fields
- `AGENTOPS_NO_SNAPSHOTS=1` — disable snapshots
- `AGENTOPS_NO_BROWSER=1` — skip opening browser on dev

## License

MIT — see [LICENSE](LICENSE).
