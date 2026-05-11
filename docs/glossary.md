# Glossary

This glossary defines the ubiquitous language of the gomarklint domain.
All contributors, documentation, and code use these terms consistently.

## Rule

A named, versioned check applied to a single Markdown file.
Each rule has a stable string key (e.g. `heading-level`, `no-bare-urls`) that serves as its public identifier in configuration, output, and documentation.
Rules are implemented as functions inside `internal/rule/` and invoked by `internal/linter/`.

## Diagnostic

A single violation emitted by a Rule.
Represented in Go as [`rule.LintError`](../internal/rule/heading_level.go):

```go
type LintError struct {
    File     string
    Line     int
    Rule     string
    Message  string
    Severity string
}
```

A Diagnostic carries everything needed to locate the problem (file, line), identify its source (rule key), describe it (message), and determine its impact (severity).

## Severity

The impact level of a Diagnostic.
Represented in Go as [`config.RuleSeverity`](../internal/config/config.go):

| Value | Meaning |
|---|---|
| `error` | Violation causes a non-zero exit code |
| `warning` | Violation is reported but does not affect the exit code |
| `off` | Rule is disabled; no Diagnostics are emitted |

Users can rely on this contract: an `error`-severity violation always produces a non-zero exit, making gomarklint safe to use as a CI gate.

## Config

The settings that control rule enablement and options.
Two Go types compose the full configuration:

- [`config.Config`](../internal/config/config.go) — global settings: which files to include/ignore, output format, and a map of per-rule configs.
- [`config.RuleConfig`](../internal/config/config.go) — per-rule settings: enabled flag, severity, and rule-specific options.

Config is loaded from `.gomarklint.json` (or the path given via `--config`) and validated at startup.
