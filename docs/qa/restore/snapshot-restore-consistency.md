# What guarantees exist around snapshot consistency during restore/replay?

[← Developer Q&A](../README.md)

`restore` · `snapshots` · `restore` · `replay` · `git`

**Current answer — v1.0.0** · 2026-05-25

**Replay** (session timeline) is strongly consistent — events are ordered from the local DuckDB/SQLite store and returned via `agentops replay` or `/api/sessions/:id/replay`.

**Restore** is **best-effort, not transactional.** Implementation in `cli/snapshots/snapshots.go`:

1. **Snapshot capture** — on session init, every 60 seconds, and on each `FILE_EDIT` event:
   - `git diff HEAD` saved as `changes.patch`
   - Untracked files copied into the snapshot directory
   - Manifest written with timestamp, git HEAD, and file list

2. **Restore steps** — `git checkout HEAD -- .` → `git apply` patch → copy untracked files back

**What is NOT guaranteed:**

- Atomic or rollback-safe restore (partial failure is possible; some git steps ignore errors)
- Cleanup of files **created after** the snapshot point
- Correctness after commits, rebases, or merges between snapshot and restore
- Stash integration or handling of staged-but-uncommitted index state separately

**Good for:** undoing a bad multi-file agent refactor in your working tree during a session.

**Not for:** backup, disaster recovery, or guaranteed point-in-time consistency across git operations.

Use `agentops restore --dry-run` to preview before applying.
