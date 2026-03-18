---
title: "Migrating to v2"
weight: 9
---

# Migrating to v2

v2 replaces the flat boolean config fields with a unified `rules` map. This enables per-rule severity levels, rule options, and a much cleaner foundation for adding new rules.

---

## What changed

### Config file format

**v1**

```json
{
  "minHeadingLevel": 2,
  "enableLinkCheck": false,
  "linkCheckTimeoutSeconds": 5,
  "skipLinkPatterns": [],
  "enableHeadingLevelCheck": true,
  "enableDuplicateHeadingCheck": true,
  "enableNoMultipleBlankLinesCheck": true,
  "enableNoSetextHeadingsCheck": true,
  "enableFinalBlankLineCheck": true,
  "include": ["README.md", "docs"],
  "ignore": [],
  "output": "text"
}
```

**v2**

```json
{
  "default": true,
  "rules": {
    "heading-level": { "enabled": true, "severity": "error", "minLevel": 2 },
    "duplicate-heading": true,
    "no-multiple-blank-lines": true,
    "no-setext-headings": true,
    "final-blank-line": true,
    "unclosed-code-block": true,
    "empty-alt-text": true,
    "external-link": { "enabled": false, "severity": "error", "timeoutSeconds": 5, "skipPatterns": [] }
  },
  "include": ["README.md", "docs"],
  "ignore": [],
  "output": "text"
}
```

### Rule value shorthand

Each rule entry accepts three forms (bool, string, or object):

| Value | Meaning |
|---|---|
| `true` | enabled, severity = `"error"` |
| `false` | disabled |
| `"error"` | enabled, severity = `"error"` |
| `"warning"` | enabled, severity = `"warning"` |
| `"off"` | disabled |
| `{ "enabled": true, "severity": "warning", ...options }` | full object form |

In the object form, `enabled` can be omitted — it defaults to `true`. These are equivalent:

```json
"heading-level": { "enabled": true, "severity": "warning", "minLevel": 2 }
"heading-level": { "severity": "warning", "minLevel": 2 }
```

### `default` key

Controls what happens to rules **not listed** in the `rules` map:

- `"default": true` — all unlisted rules are **enabled** (opt-out mode, the default)
- `"default": false` — all unlisted rules are **disabled** (opt-in mode)

### CLI flags

Per-rule CLI flags have been removed. Rule configuration belongs in the config file.

| Removed | Replacement |
|---|---|
| `--enable-heading-level-check` | Set `"heading-level": false` in config |
| `--enable-duplicate-heading-check` | Set `"duplicate-heading": false` in config |
| `--enable-no-multiple-blank-lines-check` | Set `"no-multiple-blank-lines": false` in config |
| `--enable-no-setext-headings-check` | Set `"no-setext-headings": false` in config |
| `--enable-final-blank-line-check` | Set `"final-blank-line": false` in config |
| `--enable-link-check` | Set `"external-link": true` in config |
| `--min-heading` | Set `"heading-level": { "minLevel": N }` in config |
| `--skip-link-patterns` | Set `"external-link": { "skipPatterns": [...] }` in config |

The following flags remain unchanged:

```
--config <path>     Path to config file (default: .gomarklint.json)
--output <format>   Output format: text or json
--severity <level>  Minimum severity to report: warning or error (new in v2)
```

### Go module path

If you import gomarklint as a library, update your import paths:

```
github.com/shinagawa-web/gomarklint   →   github.com/shinagawa-web/gomarklint/v2
```

---

## Migration steps

### 1. Regenerate the config file

The quickest way to get a valid v2 config is:

```sh
rm .gomarklint.json
gomarklint init
```

Then re-apply your custom settings using the new format.

### 2. Manual conversion reference

| v1 field | v2 equivalent |
|---|---|
| `"enableHeadingLevelCheck": true` | `"heading-level": true` |
| `"enableHeadingLevelCheck": false` | `"heading-level": false` |
| `"minHeadingLevel": 3` | `"heading-level": { "enabled": true, "minLevel": 3 }` |
| `"enableDuplicateHeadingCheck": true` | `"duplicate-heading": true` |
| `"enableNoMultipleBlankLinesCheck": true` | `"no-multiple-blank-lines": true` |
| `"enableNoSetextHeadingsCheck": true` | `"no-setext-headings": true` |
| `"enableFinalBlankLineCheck": true` | `"final-blank-line": true` |
| `"enableLinkCheck": true` | `"external-link": true` |
| `"linkCheckTimeoutSeconds": 10` | `"external-link": { "enabled": true, "timeoutSeconds": 10 }` |
| `"skipLinkPatterns": [...]` | `"external-link": { "enabled": true, "skipPatterns": [...] }` |

### 3. Validate

```sh
gomarklint --config .gomarklint.json README.md
```

If the config format is wrong you will see:

```
[gomarklint error]: failed to parse config file: ...
```

---

## New features in v2

### Severity levels

Downgrade a rule to a warning instead of an error:

```json
{
  "rules": {
    "no-setext-headings": "warning",
    "heading-level": { "enabled": true, "severity": "warning", "minLevel": 2 }
  }
}
```

**Exit code behavior:**

| Severity of violations | Exit code |
|------------------------|-----------|
| `error` violations present | `1` (failure) |
| `warning` violations only | `0` (success) |
| No violations | `0` (success) |

Warnings are always displayed and counted in the output, but they will **not** fail CI.

Use `--severity error` to suppress warnings from output entirely (only `error` violations are shown and counted).

### Opt-in mode

To enable only the rules you explicitly list:

```json
{
  "default": false,
  "rules": {
    "final-blank-line": true,
    "duplicate-heading": true
  }
}
```

All other rules are off unless listed.
