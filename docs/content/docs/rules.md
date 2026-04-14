---
title: "Rules"
weight: 2
---

# Rules

`gomarklint` currently runs the following checks (ordered as executed):

| Rule key                       | What it detects                                                         | Notes / Options                                                                                       |
| ------------------------------ | ----------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `final-blank-line`             | Missing final blank line at EOF                                         | Default **on**                                                                                        |
| `unclosed-code-block`          | Unclosed fenced code blocks (`` ``` ``)                                 | Default **on**                                                                                        |
| `empty-alt-text`               | Image syntax with an empty alt text                                     | Default **on**                                                                                        |
| `heading-level`                | Invalid heading level progression (e.g., H2 → H4 skip)                 | Default **on**. Option: `minLevel` (default `2`)                                                      |
| `fenced-code-language`         | Fenced code blocks without a language identifier                        | Default **on**                                                                                        |
| `duplicate-heading`            | Duplicate headings within one file                                      | Default **on**                                                                                        |
| `no-multiple-blank-lines`      | Multiple consecutive blank lines                                        | Default **on**                                                                                        |
| `no-setext-headings`           | Setext heading used instead of ATX style                                | Default **on**                                                                                        |
| `single-h1`                    | More than one H1 heading in a file                                      | Default **on**                                                                                        |
| `blanks-around-headings`       | Headings not surrounded by blank lines                                  | Default **on**                                                                                        |
| `no-bare-urls`                 | HTTP/HTTPS URLs written as bare text instead of proper Markdown links   | Default **on**                                                                                        |
| `no-empty-links`               | Links or images with an empty destination (`[]()`, `[](#)`, `[](<>)`)  | Default **on**                                                                                        |
| `no-emphasis-as-heading`       | Bold/italic text used as a heading substitute instead of ATX headings   | Default **on**. Punctuation-ending spans (`. , ; : ! ? 。 、 ； ： ！ ？`) are excluded              |
| `external-link`                | External links that fail HTTP validation                                | Default **off**. Options: `timeoutSeconds` (default `5`), `skipPatterns` (regex list)                 |

## Execution details

- Files/dirs are expanded with ignore patterns from config.
- Per-file issues are sorted by line asc before printing.
- Line count is computed as `\n` count + 1 for reporting.
