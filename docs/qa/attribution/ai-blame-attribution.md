# How do you attribute AI vs human edits reliably for AI Blame?

[← Developer Q&A](../README.md)

`attribution` · `blame` · `file-edits` · `git` · `attribution`

**Current answer — v1.0.0** · 2026-05-25

Short answer: **we don't claim reliable AI vs human attribution in v1.0.** AI Blame is event-sourced session telemetry, not git-blame-accurate provenance.

**What happens today:**

1. The filesystem watcher (`cli/collector/filesystem.go`) fires on every write/create in the project directory.
2. Each event is stored in the `attribution` table with timestamp, line counts, and a diff.
3. The agent label comes from `AttributeFileAgent()` — it scans running processes for names containing `cursor`, `claude`, `copilot`, or `aider`. If none match, the label is `unknown`.

**Limitations:**

- There is **no human vs AI classifier**. If you edit a file while Cursor is open, the edit may be tagged `cursor`.
- **Rebases, commits, merges, and cherry-picks are invisible.** Attribution reflects what AgentOps observed during a session, not current line ownership.
- **Post-rebase blame** shows historical session events; it does not reconcile with git history.

**What it's good for:** "During this agent session, what touched `src/auth.ts` and when?" — a flight recorder, not forensic git blame.

Roadmap items (see `docs/ROADMAP.md`): temporal correlation (file edit within N seconds of an LLM/tool event), multi-agent conflict detection, and `.agentignore` rules.
