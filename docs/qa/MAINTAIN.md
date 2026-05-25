# Maintaining Developer Q&A

The Developer Q&A is generated from JSON source files. **Edit the JSON, not the generated markdown.**

## Source layout

```
docs/qa/
├── MAINTAIN.md          ← this file
├── README.md            ← generated index
├── source/
│   └── *.json           ← one file per question (source of truth)
└── <category>/
    └── <slug>.md        ← generated per-question pages
```

## Adding a new question

1. Create `docs/qa/source/<slug>.json`:

```json
{
  "id": "my-question",
  "slug": "my-question",
  "title": "Your question here?",
  "category": "architecture",
  "tags": ["tag1", "tag2"],
  "answers": [
    {
      "release": "1.0.0",
      "status": "current",
      "date": "2026-05-25",
      "body": "Markdown answer..."
    }
  ]
}
```

2. Run the generator:

```bash
make qa
# or
go run scripts/generate-qa.go
# Windows
.\build.ps1 qa
```

3. Commit both the JSON source and generated markdown.

**Validation rules:**

- `id` and `slug` must match
- `release` must be semver (`1.0.0`, not `v1.0.0`)
- Exactly one answer must have `"status": "current"`
- All other answers must have `"status": "superseded"`

## Updating an answer for a new release

When behavior changes in v1.x.x:

1. Open the relevant `docs/qa/source/<slug>.json`
2. Add a **new** answer object at the top of the `answers` array:

```json
{
  "release": "1.1.0",
  "status": "current",
  "date": "2026-06-01",
  "body": "Updated answer reflecting new behavior..."
}
```

3. Change the previous answer's `"status"` from `"current"` to `"superseded"`
4. Run `make qa` (or `go run scripts/generate-qa.go`)
5. Commit source + generated files

The generator renders the current answer normally and wraps superseded answers in strikethrough (per paragraph, for GitHub compatibility).

## Release checklist

Before tagging a release, ask:

> Did this release change behavior described in `docs/qa/source/`?

If yes, update the affected Q&A entries and note it in the release notes:

```markdown
### Documentation
- Updated Developer Q&A: [ai-blame-attribution](../qa/attribution/ai-blame-attribution.md) (v1.1.0)
```

## CI

CI runs `go run scripts/generate-qa.go --check` to ensure generated markdown matches the JSON source. If CI fails, regenerate locally and commit.

## Intake for new questions

Use [GitHub Issues](https://github.com/gog1withme/AgentOps/issues) for new question requests. Label or title them clearly (e.g. "Q&A: …") so they can be triaged into `docs/qa/source/`.
