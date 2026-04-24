---
title: "CLI"
weight: 4
---

# CLI

```sh
gomarklint [files or directories] [flags]
```

If no paths are given, the tool will use `include` from `.gomarklint.json` if present, otherwise error out with:

> "please provide a markdown file or directory (or set 'include' in .gomarklint.json)"

## Flags

| Flag       | Type             | Default            | Description                                             |
| ---------- | ---------------- | ------------------ | ------------------------------------------------------- |
| `--config` | string           | `.gomarklint.json` | Path to config file.                                    |
| `--output` | `text` \| `json` | `text`             | Output format. Any other value is rejected.             |
| `--severity` | `warning` \| `error` | `warning`    | Minimum severity level to include in output (see below). |

## Severity levels

Each rule can be configured with a severity of `"error"` or `"warning"` in the config file.

| Severity  | Shown in output | Counted as issue | Causes non-zero exit |
|-----------|-----------------|------------------|----------------------|
| `error`   | ✅ always        | ✅               | ✅                   |
| `warning` | ✅ by default    | ✅               | ❌ (exit 0)          |

**Key behavior:**

- Violations tagged `warning` are displayed and counted, but the command exits `0` — they will not fail CI.
- Violations tagged `error` cause the command to exit `1`.
- `--severity error` suppresses warnings from output entirely (useful when you only want CI to see hard failures).
- `--severity warning` (default) shows both warnings and errors.

### Example

```json
{
  "rules": {
    "no-setext-headings": "warning",
    "unclosed-code-block": "error"
  }
}
```

```text
$ gomarklint README.md
  README.md:10: [warning] no-setext-headings: Setext heading found
  README.md:20: [error]   unclosed-code-block: Unclosed code block

⚠ 1 warning, ✖ 1 error — exit 1

$ gomarklint README.md --severity error
  README.md:20: [error] unclosed-code-block: Unclosed code block

✖ 1 issues found — exit 1

$ gomarklint docs/  # only warnings, no errors → exit 0
  docs/style.md:5: [warning] no-setext-headings: Setext heading found

⚠ 1 warning found — exit 0
```

## Notes

- Flags override config values when explicitly provided.
- Only existing files or directories are accepted; matching paths are then filtered by `ignore` (from config).
- Exit behavior: exits `1` if any `error`-severity violations are found; exits `0` for warnings-only or clean runs.
