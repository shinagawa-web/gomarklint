---
title: "CLI"
weight: 3
---

# CLI

```sh
gomarklint [files or directories] [flags]
```

If no paths are given, the tool will use `include` from `.gomarklint.json` if present, otherwise error out with:

> "please provide a markdown file or directory (or set 'include' in .gomarklint.json)"

## Flags

| Flag                                     | Type             | Default            | Description                                                                            |
| ---------------------------------------- | ---------------- | ------------------ | -------------------------------------------------------------------------------------- |
| `--config`                               | string           | `.gomarklint.json` | Path to config file. Loaded if the file exists.                                        |
| `--min-heading`                          | int              | `2`                | Minimum heading level considered by the heading-level rule.                            |
| `--enable-link-check`                    | bool             | `false`            | Enable external link checking.                                                         |
| `--enable-heading-level-check`           | bool             | `true`             | Enable heading level validation.                                                       |
| `--enable-duplicate-heading-check`       | bool             | `true`             | Enable duplicate heading detection.                                                    |
| `--enable-no-multiple-blank-lines-check` | bool             | `true`             | Enable check that disallows multiple consecutive blank lines.                          |
| `--enable-no-setext-headings-check`      | bool             | `true`             | Enable check that disallows Setext-style headings.                                     |
| `--enable-final-blank-line-check`        | bool             | `true`             | Enable check that enforces a final blank line at end of file.                          |
| `--skip-link-patterns`                   | string[] (regex) | `[]`               | Regex patterns; matching URLs are skipped by link check. Can be passed multiple times. |
| `--output`                               | `text` \| `json` | `text`             | Output format. Any other value is rejected.                                            |

## Notes

- Flags override config values when explicitly provided.
- Only existing files or directories are accepted; matching paths are then filtered by ignore (from config).
- Exit behavior: the command returns a non-nil error (non-zero exit) if issues are found, zero otherwise.
