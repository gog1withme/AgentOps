# What percentage of README features are production-complete vs heuristic?

[← Developer Q&A](../README.md)

`roadmap` · `maturity` · `roadmap` · `v1.0.1` · `heuristics`

**Current answer — v1.0.1** · 2026-05-26

Honest v1.0.1 breakdown of README and release-note claims:

| Tier | Approx. share | Examples |
|------|---------------|----------|
| **Production-solid** | ~60–65% | CLI (`init`, `dev`, `budget`, `restore`, `trace`, `replay`, `diff`, `blame`, `prompt`, `mcp`, `alerts`, `upgrade`, `context summary`), live SSE feed, LLM proxy + configurable pricing, privacy scrubber, budgets, prompt hashing + diff, trace UI, DuckDB/SQLite store, MCP p95 (rolling window) |
| **Working but heuristic** | ~20–25% | AI blame (process-name attribution), token efficiency score (n-gram overlap), dangerous-shell alerts (substring blocklist), cost estimates (user-configurable but still estimated) |
| **MVP / partial UI** | ~10% | Context Inspector (duplicate/noisy detection shipped; no memory drift or per-segment token weight yet) |
| **Roadmap only** | remainder | OpenTelemetry export, Parquet storage, approval gates, multi-agent conflict detection, smart insights (AI-powered), evals, alert webhooks, cloud/team features, RAG tracing, compliance exports |

**Bottom line:** v1.0.1 improves trust on metrics and cost while adding first actionable context analysis — still not yet a Datadog-grade attribution or security engine.

See [ROADMAP.md](../../ROADMAP.md) for what ships next vs what is planned.

---

**Superseded answer — v1.0.0** · 2026-05-25

~~Honest v1.0.0 breakdown of README and release-note claims:~~

~~| Tier | Approx. share | Examples |~~
~~|------|---------------|----------|~~
~~| **Production-solid** | ~55–60% | CLI (`init`, `dev`, `budget`, `restore`, `trace`, `replay`, `blame`, `prompt`, `mcp`, `alerts`), live SSE feed, LLM proxy + token/cost tracking, privacy scrubber, budgets (alert/pause/kill), prompt hashing + diff, trace UI, DuckDB/SQLite store |~~
~~| **Working but heuristic** | ~25–30% | AI blame (process-name attribution), token efficiency score (n-gram overlap), dangerous-shell alerts (substring blocklist), MCP "p95" (EMA of last latency, not true p95), cost estimates (hardcoded model rates) |~~
~~| **MVP / partial UI** | ~10–15% | Context Inspector (lists LLM calls; no duplicate-context detection, per-file token weight, or memory drift) |~~
~~| **Roadmap only** | remainder | OpenTelemetry export, Parquet storage, approval gates, multi-agent conflict detection, smart insights, evals, cloud/team features, RAG tracing, compliance exports |~~

~~**Bottom line:** v1.0 is a real local flight recorder with working restore, budgets, and proxy — not yet a Datadog-grade attribution or security engine.~~

~~See [ROADMAP.md](../../ROADMAP.md) for what ships next vs what is planned.~~
