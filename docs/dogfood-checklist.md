# Dogfood Checklist

Manual validation for AgentOps integration before public release. Run on your daily driver (Windows first), then macOS or Linux.

**Automated gate (v1.0.1):** `go test ./cli/...` — passed on Windows 2026-05-26.

**Setup before each session**

1. `agentops init` (once)
2. Source env: run output of `agentops env` (e.g. `. ~/.agentops/env.ps1`)
3. Restart Cursor so it picks up `OPENAI_BASE_URL` / `ANTHROPIC_BASE_URL`
4. `agentops dev` in the project directory

---

| # | Test | Pass criteria | Pass? |
|---|------|---------------|-------|
| 1 | `agentops init` | Completes in < 3s, tools detected, privacy rule count shown | ☐ |
| 2 | `agentops dev` + dashboard | Dashboard loads; SSE stable (no crash on refresh) | ☐ |
| 3 | LLM proxy | One Cursor chat → `llm_call` in live feed with tokens/cost | ☐ |
| 4 | Scrubber | `go test ./cli/scrubber/...` passes | ☐ |
| 5 | Shell hook | Run `echo test` in new shell → `shell_command` event in feed | ☐ |
| 6 | FS watcher | Save a file → `file_edit` event + attribution row | ☐ |
| 7 | Budget 80% | Lower limit (`agentops budget set`), trigger → warning banner | ☐ |
| 8 | Budget 100% | Hit limit → configured `alert`/`pause` action fires | ☐ |
| 9 | Snapshots | `agentops restore --dry-run` lists snapshots for session | ☐ |
| 10 | Prompt hash | After LLM call → `agentops prompt list` is non-empty | ☐ |
| 11 | MCP | Trigger MCP tool → `/mcp` page shows server stats | ☐ |
| 12 | Efficiency | LLM row in feed shows efficiency badge/score | ☐ |
| 13 | Restore UI | Trace detail → Restore modal calls API successfully | ☐ |

---

## Quick commands

```powershell
# Windows
.\build.ps1 build
.\bin\agentops.exe init
.\bin\agentops.exe env          # copy/run the printed command
.\bin\agentops.exe dev
.\bin\agentops.exe doctor --verbose
go test ./cli/...
```

```bash
# macOS / Linux
make build
./bin/agentops init
eval "$(./bin/agentops env 2>/dev/null | grep -E '^(source|\.)' | head -1)"  # or run printed command
./bin/agentops dev
./bin/agentops doctor --verbose
go test ./cli/...
```

## Notes

- **Cursor must be restarted** after sourcing env vars; child processes inherit env at launch.
- MCP proxy only applies to **URL-based** servers in `mcp.json` (not stdio/command servers).
- Windows uses **SQLite**; macOS/Linux use **DuckDB** for the event store.
- Log failures as GitHub issues or inline fixes; re-run this checklist over 3–5 real coding sessions.

## Exit gate (public v0.1)

- All 13 items pass on Windows
- macOS or Linux verified for DuckDB store + Unix hooks
- CI green without `|| true` on E2E
- No daemon crashes during a 30-minute session with live feed open
- `agentops doctor --verbose` reports integration checks green after `init` + `dev`
- Review `docs/qa/source/` — update any Q&A answers that changed behavior in this release (`make qa`)
