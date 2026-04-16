# CLAUDE.md

## Build & Test

Use `make` targets — do not run raw `go build` / `go test` commands.

| Task | Command |
|---|---|
| Build | `make build` |
| Unit tests | `make test` |
| E2E tests | `make test-e2e` |
| All tests | `make test-all` |
| Benchmarks | `make bench` |
| Static analysis | `make static-lint` |

## Git

- **Commits**: no `Co-Authored-By` trailer
- **Updating a feature branch**: `git fetch origin main && git rebase origin/main` — never merge main into a feature branch

## GitHub (issues, PRs, comments, commit messages)

Write in **English**. Japanese is for direct conversation only.

### Rule implementation issues

When creating an issue for a new lint rule:

1. Write the issue body with: Overview, Problem, Proposed rule key, Detection logic (with examples table), Fix, References.
2. Add a follow-up **comment** titled `## Implementation Plan` containing:
   - `### Files to create` — new files list
   - `### Files to modify` — modified files list
   - `### Checklist` — implementation, tests, config registration, linter integration, e2e, `go test ./...`

See [#106](https://github.com/shinagawa-web/gomarklint/issues/106#issuecomment-4205193874) as the canonical example.

## Project context

- **#76**: master tracking issue for rule expansion (Priority 1 → 2 → 3)
- **#123**: migration guide from markdownlint/remark-lint/textlint — start after #76 Priority 2 rules land
