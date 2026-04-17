---
title: "Disable Comments"
weight: 3
---

# Disable Comments

You can suppress lint violations for specific lines or blocks by adding HTML comment directives directly in your Markdown file.

## Directives

### Block disable / enable

Disable all rules from the directive line onward until re-enabled:

```markdown
<!-- gomarklint-disable -->
https://example.com
<!-- gomarklint-enable -->
```

Disable specific rules only:

```markdown
<!-- gomarklint-disable no-bare-urls -->
https://example.com
<!-- gomarklint-enable no-bare-urls -->
```

Re-enable a specific rule inside a block that disabled everything:

```markdown
<!-- gomarklint-disable -->
https://example.com
<!-- gomarklint-enable no-bare-urls -->
https://example.com   ← no-bare-urls is checked again here
<!-- gomarklint-enable -->
```

### Single-line disable

Suppress violations on the same line the directive appears:

```markdown
https://example.com <!-- gomarklint-disable-line -->

https://example.com <!-- gomarklint-disable-line no-bare-urls -->
```

### Next-line disable

Suppress violations on the line immediately following the directive:

```markdown
<!-- gomarklint-disable-next-line -->
https://example.com

<!-- gomarklint-disable-next-line no-bare-urls -->
https://example.com
```

## Directive reference

| Directive | Scope | Effect |
|---|---|---|
| `<!-- gomarklint-disable -->` | block start | Disable all rules |
| `<!-- gomarklint-disable rule [rule…] -->` | block start | Disable named rules |
| `<!-- gomarklint-enable -->` | block end | Re-enable all rules |
| `<!-- gomarklint-enable rule [rule…] -->` | block end | Re-enable named rules |
| `<!-- gomarklint-disable-line -->` | current line | Disable all rules |
| `<!-- gomarklint-disable-line rule [rule…] -->` | current line | Disable named rules |
| `<!-- gomarklint-disable-next-line -->` | next line | Disable all rules |
| `<!-- gomarklint-disable-next-line rule [rule…] -->` | next line | Disable named rules |

## Notes

- A directive that disables all rules takes priority over one that disables only named rules when both apply to the same line.
- Files without any `gomarklint-disable` comment are parsed with zero overhead — the directive scanner is skipped entirely.
- Rule names match the keys listed in [Rules](../rules).
