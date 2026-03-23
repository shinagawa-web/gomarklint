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
- New rule files must have 100% coverage (verify with `go tool cover -func`)

## Conventions

- Rule function signature: `func CheckXxx(filename string, lines []string, offset int) []LintError`
- Do not set `Severity` in rule functions — it is assigned by `linter.withSeverity()`
- Use `strings.TrimSpace()` when matching fence/heading markers
- Test offset: pass `offset > 0` in at least one test case to verify line numbers shift correctly
- Table-driven tests; keep `Severity` field out of expected `LintError` (it's empty at rule level)
