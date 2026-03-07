---
title: "Rules"
weight: 2
---

# Rules

`gomarklint` currently runs the following checks (ordered as executed):

| Rule key                       | What it detects                                         | Notes / Options                                                                                        |
| ------------------------------ | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `final-blank-line`             | Missing final blank line at EOF                         | Toggle: `--enable-final-blank-line-check` (default **on**)                                             |
| `unclosed-code-block`          | Unclosed fenced code blocks (`` ``` ``)                 | Always on                                                                                              |
| `empty-alt-text`               | Image syntax with an empty alt text                     | Always on                                                                                              |
| `heading-level`                | Invalid heading level progression (e.g., H2 → H4 skip) | Toggle: `--enable-heading-level-check` (default **on**) / `--min-heading` (default **2**)              |
| `duplicate-heading`            | Duplicate headings within one file                      | Toggle: `--enable-duplicate-heading-check` (default **on**)                                            |
| `no-multiple-blank-lines`      | Multiple consecutive blank lines                        | Toggle: `--enable-no-multiple-blank-lines-check` (default **on**)                                      |
| `external-link`                | External links that fail validation                     | Toggle: `--enable-link-check` (default **off**). Skips URLs that match `--skip-link-patterns` (regex). |
| `no-setext-headings`           | Setext heading used instead of ATX style                | Toggle: `--enable-no-setext-headings-check` (default **on**)                                           |

## Execution details

- Files/dirs are expanded with ignore patterns from config.
- Per-file issues are sorted by line asc before printing.
- Line count is computed as `\n` count + 1 for reporting.
