---
title: "Configuration"
weight: 5
---

# Configuration

A JSON config is read from the path given by `--config` (defaults to `.gomarklint.json`) if the file exists.

## Example

```json
{
  "default": true,
  "rules": {
    "final-blank-line": true,
    "unclosed-code-block": true,
    "empty-alt-text": true,
    "heading-level": { "severity": "error", "minLevel": 2 },
    "duplicate-heading": true,
    "no-multiple-blank-lines": true,
    "no-setext-headings": "warning",
    "external-link": { "severity": "error", "timeoutSeconds": 5, "skipPatterns": ["^https://localhost"] }
  },
  "include": ["docs", "README.md"],
  "ignore": ["node_modules", "vendor"],
  "output": "text"
}
```

## Top-level fields

| Field     | Type     | Default                      | Description                                                       |
| --------- | -------- | ---------------------------- | ----------------------------------------------------------------- |
| `default` | bool     | `true`                       | Whether unlisted rules are enabled by default (opt-out mode). Set to `false` for opt-in mode. |
| `rules`   | object   | all rules enabled as `error` | Per-rule configuration. See [Rule values](#rule-values) below.    |
| `include` | string[] | `["README.md", "testdata"]`  | Paths to lint when no CLI paths are provided.                     |
| `ignore`  | string[] | `[]`                         | Path patterns to exclude.                                         |
| `output`  | string   | `text`                       | `text` or `json`.                                                 |

## `default` field

Controls how rules **not listed** in `rules` are treated.

| Value | Behavior |
|---|---|
| `true` (default) | All unlisted rules are **enabled** — opt-out mode. List only the rules you want to disable or customize. |
| `false` | All unlisted rules are **disabled** — opt-in mode. Only rules explicitly listed in `rules` will run. |

**Opt-out example** — disable one rule, keep everything else:

```json
{ "default": true, "rules": { "no-setext-headings": false } }
```

**Opt-in example** — run only `final-blank-line`:

```json
{ "default": false, "rules": { "final-blank-line": true } }
```

## Rule values

Each entry in `rules` accepts three forms:

| Value | Meaning |
|---|---|
| `true` | enabled, severity = `"error"` |
| `false` | disabled |
| `"error"` | enabled, severity = `"error"` |
| `"warning"` | enabled, severity = `"warning"` |
| `"off"` | disabled |
| `{ "severity": "warning", ...options }` | full object form |

In the object form, `enabled` can be omitted — it defaults to `true`. These are equivalent:

```json
"heading-level": { "enabled": true, "severity": "warning", "minLevel": 2 }
"heading-level": { "severity": "warning", "minLevel": 2 }
```

## Available rules

| Rule | Default | Options |
|---|---|---|
| `final-blank-line` | `error` | — |
| `unclosed-code-block` | `error` | — |
| `empty-alt-text` | `error` | — |
| `fenced-code-language` | `error` | — |
| `heading-level` | `error` | `minLevel` (int, default `2`) |
| `duplicate-heading` | `error` | — |
| `no-multiple-blank-lines` | `error` | — |
| `no-setext-headings` | `error` | — |
| `single-h1` | `error` | — |
| `blanks-around-headings` | `error` | — |
| `no-bare-urls` | `error` | — |
| `no-empty-links` | `error` | — |
| `no-emphasis-as-heading` | `error` | — |
| `blanks-around-lists` | `error` | — |
| `blanks-around-fences` | `error` | — |
| `no-hard-tabs` | `error` | — |
| `no-trailing-punctuation` | `error` | `punctuation` (string, default `".,;:!"`) |
| `max-line-length` | disabled | `lineLength` (int, default `80`) |
| `external-link` | disabled | `timeoutSeconds` (int, default `5`), `skipPatterns` (string[]), `allowedStatuses` (int[]) |

## Notes

- If no CLI paths are provided, `include` becomes the target set.
- `external-link` is disabled by default due to network cost. Enable it explicitly in config.
