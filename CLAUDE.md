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
- **Pre-push hook**: run `make install-hooks` once after cloning to automatically run lint and unit tests before each push

## GitHub (issues, PRs, comments, commit messages)

Write in **English**. Japanese is for direct conversation only.

### Rule implementation issues

When creating an issue for a new lint rule:

1. Write the issue body with: Overview, Problem, Proposed rule key, Detection logic (with examples table), Fix, References.
2. Add a follow-up **comment** titled `## Implementation Plan` containing:
   - `### Files to create` — new files list
   - `### Files to modify` — modified files list
   - `### Checklist` — implementation, tests, config registration, linter integration, e2e, `make test-all`

See [#106](https://github.com/shinagawa-web/gomarklint/issues/106#issuecomment-4205193874) as the canonical example.

## Adding a new rule

When adding a new lint rule:

1. Ensure `generateComplexMarkdown` in `cmd/root_bench_test.go` contains content that exercises the new rule's main scan path without producing violations.
2. If the rule is disabled by default, explicitly enable it in `benchmarkConfig()` with violation-free option values.
3. Do **not** add a rule-level `_bench_test.go` under `internal/rule/` — the CI benchmark comparison runs only on `cmd/root_bench_test.go`.
4. Verify `TestBenchmarkContentIsViolationFree` still passes after your changes.

## Project context

- **#76**: master tracking issue for rule expansion (Priority 1 → 2 → 3)
- **#123**: migration guide from markdownlint/remark-lint/textlint — start after #76 Priority 2 rules land
