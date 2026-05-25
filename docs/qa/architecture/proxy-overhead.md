# What is the overhead of proxying and token inspection on long autonomous sessions?

[← Developer Q&A](../README.md)

`architecture` · `proxy` · `performance` · `tokens` · `latency`

**Current answer — v1.0.0** · 2026-05-25

The LLM proxy (`cli/collector/proxy.go`) is a **streaming pass-through**:

1. Read full request body from the agent
2. Forward to upstream (OpenAI or Anthropic)
3. Read full response body
4. Return response to the agent
5. Asynchronously ingest metadata + scrubbed content into the event store

**Added latency:** Approximately one local HTTP hop plus JSON parse, privacy scrub, and DB write — not an extra upstream round-trip.

**Memory:** Full request and response bodies are buffered per call (120s client timeout). Large context windows mean proportional RAM per request.

**Backpressure:** The collector event channel holds 512 events; overflow drops events with a warning (`cli/collector/collector.go`).

**Efficiency scoring:** After each LLM call, `ScoreEfficiency()` (`cli/collector/efficiency.go`) scans prompt chunks against output text — O(chunks × output length), cheap at typical sizes but non-zero at 100k+ tokens.

**Published benchmarks:** None yet. Dogfood target (see `docs/dogfood-checklist.md`) is no daemon crashes during a 30-minute session with the live feed open.

For most interactive coding sessions the proxy should feel invisible; very long autonomous runs with huge payloads are the stress case.
