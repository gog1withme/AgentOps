# How are dangerous shell detections implemented without high false positives?

[← Developer Q&A](../README.md)

`security` · `shell` · `alerts` · `security` · `false-positives`

**Current answer — v1.0.0** · 2026-05-25

v1.0 uses a **small substring blocklist**, not semantic or ML-based analysis. Rules live in `cli/collector/alerts.go`:

```
rm -rf /
format c:
mkfs
:(){ :|:& };:   (fork bomb)
dd if=/dev/zero
```

When a hooked shell posts a matching command to `/api/ingest/shell`, a `dangerous_shell` critical alert is raised. The same patterns appear in `agentops sec scan`.

**False positives:** Low for exact matches on real commands. Higher if someone echoes or discusses these strings in a hooked terminal (e.g. `echo "rm -rf /"`).

**False negatives:** High — many destructive patterns are not covered: `rm -rf .`, `git push --force`, credential exfil, writes outside the project, or commands run in agent-integrated terminals that bypass shell hooks.

**Alert dedup:** Same session + rule is suppressed for 5 minutes to avoid spam.

**Design intent in v1.0:** Catch obviously catastrophic patterns with near-zero tuning, not provide comprehensive command policy enforcement. Roadmap: approval gates, richer rule sets, and hallucinated-command detection (see `docs/ROADMAP.md`).
