# AgentOps — Build Context

> Universal telemetry + observability layer for AI coding agents.
> "Datadog + Linear for AI developer workflows."

> **Shipped status:** See [docs/ROADMAP.md](docs/ROADMAP.md) and [docs/releases/](docs/releases/) for what is live today (v1.0.1). This file is the internal build spec — checkboxes below track aspirational scope, not release status.

> **Public roadmap:** [docs/ROADMAP.md](docs/ROADMAP.md) — user-facing feature direction.

---

## Product North Star

```bash
npx agentops init
# or
curl -fsSL https://agentops.dev/install | sh
```

Zero SDK required. Passive hooks. Beautiful local UI. Works with existing agents.

---

## MVP Phases

### v0.1 — Core Observability (Ship First)

**Goal:** Single command install that immediately delivers value — and earns trust from day one.

**CLI**
- [ ] `agentops init` — auto-detect Claude Code, Cursor, Copilot, Aider, shells, MCP configs, OpenAI SDKs
- [ ] `agentops dev` — start local collector + web dashboard + live streaming metrics

**Data Collection**
- [ ] OpenAI-compatible proxy (intercepts LLM calls without SDK changes)
- [ ] Shell command interception
- [ ] Filesystem watcher (file diffs, edits, AI attribution per change)
- [ ] Claude Code + Cursor process detection

**Privacy & Secrets Filter (ship with v0.1 — non-negotiable)**
- [ ] Pre-storage scrubbing pipeline: strip secrets before any event touches DuckDB
- [ ] Pattern matching for: `sk-*`, `AKIA*` (AWS), Bearer tokens, JWTs, `.env` values, PII patterns
- [ ] Redact prompt/output content beyond file paths when `AGENTOPS_REDACT_CONTENT=1`
- [ ] Scrubber runs as a middleware step in the collector before the store write
- [ ] `agentops init` outputs: `✓ Privacy filter: active (12 pattern rules)`
- [ ] User can extend patterns via `~/.agentops/scrub_patterns.txt`

**Agent Budget + Guardrails (ship with v0.1)**
- [ ] `agentops budget set --cost 2.00 --tools 50 --action pause` — define limits per session
- [ ] Budget daemon checks limits every 5s and on each new LLM/tool event
- [ ] Actions: `pause` (sends SIGSTOP to agent process), `alert` (dashboard + terminal), `kill` (SIGTERM)
- [ ] `agentops budget status` — show current spend vs limits
- [ ] Limits stored in `~/.agentops/config.json` under `"budget": {}`
- [ ] Dashboard: persistent banner when within 80% of any limit

**Storage**
- [ ] Local event store with DuckDB + Parquet
- [ ] Event schema (OpenTelemetry-compatible)
- [ ] AI diff attribution table (tracks every AI-caused file change with source + session)

**Dashboard** (`localhost:4318`)
- [ ] Live activity feed (Claude edited 4 files, Cursor generated migration, etc.)
- [ ] Token usage tracking
- [ ] Cost tracking per agent/model/session
- [ ] Agent trace replay timeline
- [ ] Budget status banner + spend meter

**Infra**
- [ ] CLI daemon (local collector process)
- [ ] Next.js + Tailwind + shadcn/ui dashboard
- [ ] OpenTelemetry-compatible event transport

---

### v0.2 — Intelligence Layer

**Goal:** Surface actionable insights automatically. Make developers faster and safer.

**CLI**
- [ ] `agentops trace` — list and inspect traces
- [ ] `agentops replay <run_id>` — full timeline replay
- [ ] `agentops diff <run_a> <run_b>` — compare runs
- [ ] `agentops alerts` — view and configure alerts
- [ ] `agentops eval run <tests/>` — run evals
- [ ] `agentops doctor` — diagnose integration health
- [ ] `agentops upgrade` — self-update
- [ ] `agentops restore <run_id> [--at <timestamp>]` — restore repo to any point in a session
- [ ] `agentops blame <file>` — show AI attribution log for a file (who changed what, when)
- [ ] `agentops prompt list` — list captured system prompt versions
- [ ] `agentops prompt diff <hash_a> <hash_b>` — diff two prompt versions

**Prompt Versioning**
- [ ] Capture full system prompt per session, hash with SHA-256, store in `prompts` table
- [ ] Link each `EventLLMCall` to its `prompt_hash`
- [ ] `agentops prompt list` shows: hash, first seen, sessions using it, avg cost, avg tokens
- [ ] `agentops prompt diff` renders a unified diff of two prompt versions
- [ ] Foundation for evals: "run these test cases against prompt v3 vs v7"

**MCP Server Health Monitoring**
- [ ] Dedicated MCP health panel in dashboard (`/mcp`)
- [ ] Per-server metrics: call count, error rate, avg latency, response size, last seen
- [ ] Flag slow servers (p95 latency > 2s), failing servers (error rate > 10%), large payloads (> 100KB)
- [ ] Timeline view: MCP call volume over time, stacked by server
- [ ] Alert: "This MCP server increased latency 3x compared to last session"
- [ ] `agentops mcp list` — show all detected MCP servers + current health

**AI Diff Attribution**
- [ ] Every file write captured with: source agent, session ID, run ID, timestamp, git diff
- [ ] `attribution` table in DuckDB: `file_path, agent, session_id, timestamp, lines_added, lines_removed, diff`
- [ ] Dashboard `/blame` page: file browser → click file → see full AI edit history
- [ ] `agentops blame <file>` in terminal: renders table of AI edits like `git log -p`
- [ ] Trace detail page shows file changes with attribution inline

**Time-Travel Restore**
- [ ] On `agentops dev` start: snapshot current git status (`git stash` + manifest of tracked files)
- [ ] Every 60s and on each `EventFileEdit` batch: save incremental file snapshots to `~/.agentops/snapshots/<session_id>/`
- [ ] `agentops restore <run_id>` — list available restore points
- [ ] `agentops restore <run_id> --at <timestamp>` — restore working tree to that moment
- [ ] Restore uses git apply on stored diffs; falls back to direct file copy for untracked files
- [ ] Dashboard trace timeline has a "Restore to here" button on every event

**Token Efficiency Scoring**
- [ ] After each LLM call, analyze: which input context segments were referenced in output or subsequent tool calls
- [ ] Compute `efficiency_score` = (referenced_tokens / total_prompt_tokens) × 100
- [ ] Store score on `EventLLMCall` in `Metadata["efficiency_score"]`
- [ ] Dashboard shows per-call score as a color-coded badge (green >70%, amber 40-70%, red <40%)
- [ ] Session-level efficiency trend chart in `/cost` page
- [ ] Smart callout: "You're consistently loading `src/legacy/` (avg 8,400 tokens, 14% referenced)"

**Alerts Engine**
- [ ] Token explosion detection
- [ ] Infinite loop detection
- [ ] Prompt injection flagging
- [ ] Excessive retry detection
- [ ] Cost spike alerts
- [ ] Context overflow warnings
- [ ] Hallucinated shell command detection
- [ ] Budget threshold warnings (80% and 100% of set limits)
- [ ] MCP server degradation alerts
- [ ] Low token efficiency alerts (session avg < 30%)

**Context Inspector**
- [ ] Retrieved files view with token weight
- [ ] Duplicate context detection
- [ ] Noisy context flagging
- [ ] Memory drift tracking
- [ ] Token efficiency score per file

**Security Center**
- [ ] `agentops sec scan`
- [ ] Dangerous shell command flagging
- [ ] Leaked secrets detection (backed by same scrubber patterns from v0.1)
- [ ] Suspicious MCP server detection
- [ ] Unsafe permission warnings
- [ ] Prompt injection attempt log with context of what the model did next

**Smart Insights (AI-powered)**
- [ ] "Claude retried this tool N times" callouts
- [ ] "X% of tokens wasted on duplicate context"
- [ ] "This MCP server increased latency Nx"
- [ ] Prompt quality scoring
- [ ] "Prompt v7 costs 30% less than v3 with similar output length"

**Deep SDK Integrations (optional, Layer 2)**
- [ ] LangChain
- [ ] Vercel AI SDK
- [ ] OpenAI Agents SDK
- [ ] AutoGen / CrewAI

---

### v0.3 — Team & Cloud

**Goal:** Expand from solo dev tool to team platform.

**Collaboration**
- [ ] Cloud trace sync
- [ ] Team dashboards
- [ ] Shareable trace links ("Why did Claude delete production?")
- [ ] Session recording (Loom for AI workflows)
- [ ] Shared prompt library — team can publish and pin prompt versions

**Multi-Agent Conflict Detection**
- [ ] Correlate concurrent writes to the same file from different agent sources
- [ ] Flag: "Claude and Cursor both modified `auth.ts` within 30 seconds"
- [ ] Show conflicting diffs side by side in dashboard
- [ ] Alert: "File `db/schema.sql` has been written by 3 different agents this session"
- [ ] Lock suggestion: "Consider .agentlock to coordinate access to migration files"

**Analytics**
- [ ] Workflow analytics — success rate, coding velocity, retry rates, hallucination rates
- [ ] AI Workflow Benchmarks — compare models, prompts, agents, tools
- [ ] Per-user and per-project breakdowns
- [ ] Prompt version performance leaderboard (cost, efficiency, task success rate)

**Integrations**
- [ ] CI/CD integration
- [ ] Org-level analytics
- [ ] `.agentignore` file support — per-repo rules for what AgentOps should not capture

**Monetization (Paid Tier)**
- [ ] Hosted dashboards
- [ ] Enterprise auth (SSO/SAML)
- [ ] Long-term trace retention
- [ ] Compliance exports
- [ ] Advanced prompt governance (approval workflows for prompt changes)

---

## Dashboard Sections (All Phases)

| Section | Phase | Notes |
|---|---|---|
| Live Activity Feed | v0.1 | Real-time SSE stream |
| Cost + Token Analytics | v0.1 | Per model/agent/session |
| Agent Trace Replay | v0.1 | Timeline + restore button |
| Budget Status | v0.1 | Banner + spend meter |
| AI Diff Attribution (`/blame`) | v0.2 | File-level AI edit history |
| MCP Health (`/mcp`) | v0.2 | Per-server metrics |
| Context Inspector | v0.2 | Token weight + efficiency score |
| Prompt Versions (`/prompts`) | v0.2 | Hash list + diff viewer |
| Security Center | v0.2 | Alerts + injection log |
| Multi-Agent Conflicts | v0.3 | Concurrent write detection |
| Workflow Analytics | v0.3 | Team-level metrics |

Design inspiration: Linear, Vercel, Raycast — NOT Grafana.

---

## Tech Stack

| Layer | Choice | Notes |
|---|---|---|
| CLI / Daemon | Go | Faster iteration, huge CLI ecosystem |
| Local Storage | DuckDB + Parquet | Analytics-optimized, local-first |
| Event Transport | OpenTelemetry | Ecosystem compatibility, enterprise-ready |
| Dashboard | Next.js + Tailwind + shadcn/ui | |
| Passive Hooks | Shell interception, FS watcher, OAI proxy, MCP proxy, editor hooks | |
| Secrets Scrubbing | Go middleware (pre-storage) | Pattern file is user-extensible |
| File Snapshots | Git diff + flat file copy fallback | Powers time-travel restore |

---

## Architecture Summary

```
AI Tools (Claude, Cursor, Copilot, Aider, OpenAI SDK)
    │
    ▼
Local Collector / CLI Daemon
(event ingestion, token tracking, tool tracing, context analysis, security checks)
    │
    ▼
Privacy Scrubber  ←── scrub_patterns.txt
(strip secrets, PII, tokens before storage)
    │
    ▼
Local Event Store (DuckDB + Parquet)
(events, attribution, prompts, snapshots, budgets)
    │
    ▼
Web UI — localhost:4318
(activity, traces, replay, cost, MCP health, blame, prompts, security, budget)
```

---

## Open Source Model

**Open source:** CLI, local dashboard, tracing SDK, event schema, core collector, scrubber patterns

**Paid:** hosted dashboards, team collaboration, cloud traces, enterprise auth, compliance, long retention, org analytics, prompt governance

Reference model: GitLab, Sentry, PostHog, Supabase, Grafana.

---

## Key Principles (Do Not Compromise)

1. **Zero-friction onboarding** — one command, works immediately
2. **Privacy by default** — secrets scrubbed before storage, on by default, never opt-in
3. **Passive by default** — no SDK required, no workflow disruption
4. **Agent-agnostic** — never become a framework or wrapper
5. **Safety before features** — budget + guardrails ship in v0.1, not v0.3
6. **Beautiful UX** — infra tools don't have to be ugly
7. **Reliable** — no dropped traces, must feel invisible
8. **OpenTelemetry-compatible** — stay interoperable

---

## Risks to Avoid

- Becoming "yet another AI wrapper framework"
- Requiring SDK lock-in before delivering value
- Shipping observability without privacy filtering (enterprise blocker)
- Building v0.3 features before v0.1 is solid
- Dashboard complexity (Grafana trap)
- Letting agents run unbounded without guardrail primitives
