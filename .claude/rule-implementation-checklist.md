# Rule Implementation Checklist

Follow this checklist whenever adding a new rule (e.g. `fenced-code-language`, `no-bare-urls`, etc.).

## Files to create

- [ ] `internal/rule/<rule_name>.go` — rule function
- [ ] `internal/rule/<rule_name>_test.go` — unit tests
- [ ] `e2e/fixtures/<rule_name>.md` — E2E fixture file

## Files to modify

- [ ] `internal/config/config.go` — add to `Default()` AND `DefaultConfigJSON` (keep in sync)
- [ ] `internal/linter/linter.go` — register in `collectErrors()`
- [ ] `e2e/e2e_test.go` — add E2E test case, update file/violation counts if needed

## Code review checklist

> Points raised during actual reviews. Updated as issues are found.

- E2E tests must cover both the success case (No issues found) and the failure case
- Run `go test ./... -skip TestE2E` and verify all packages remain at 100% coverage
- Avoid deep if nesting — use early `continue` or extract helpers to keep the main loop flat
- After pushing a PR, check Copilot review comments and address all of them before merging

## Conventions

- Rule function signature: `func CheckXxx(filename string, lines []string, offset int) []LintError`
- Do not set `Severity` in rule functions — it is assigned by `linter.withSeverity()`
- Use `strings.TrimSpace()` when matching fence/heading markers
- Test offset: pass `offset > 0` in at least one test case to verify line numbers shift correctly
- Table-driven tests; keep `Severity` field out of expected `LintError` (it's empty at rule level)
- Fenced code block detection: use `openingFenceMarker()` to detect opening fences and `IsClosingFence()` to detect closing fences (both in `fence.go`). These handle backtick/tilde, variable-length runs, and CommonMark-compliant longer closing fences (#95)

## E2E test conventions

- **Rule-specific E2E config**: always set `"default": false` and opt-in only the rules under test. This prevents future rule additions from breaking unrelated E2E tests.
- **Fixture naming**: `<rule_name>_valid.md` / `<rule_name>_violation.md` means valid/invalid **for that specific rule**. A `_valid` fixture may still trigger violations from other rules under the default E2E config (`.gomarklint.json`), and that is expected. Example: `heading_level_one.md`, `single_h1_valid.md`.
- **Directory recursion test** (`TestE2E_MultipleFiles/DirectoryRecursion`): runs all fixtures under the default config. When adding fixtures, update the file count and add assertions for any new violations that appear. It is normal for rule-specific `_valid` fixtures to produce errors here from other rules.
