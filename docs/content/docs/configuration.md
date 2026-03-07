---
title: "Configuration"
weight: 4
---

# Configuration

A JSON config is read from the path given by `--config` (defaults to `.gomarklint.json`) if the file exists.

## Example

```json
{
  "include": ["docs", "README.md"],
  "ignore": ["node_modules", "vendor"],
  "outputFormat": "text",
  "minHeadingLevel": 2,
  "enableLinkCheck": false,
  "enableHeadingLevelCheck": true,
  "enableDuplicateHeadingCheck": true,
  "skipLinkPatterns": [
    "^https://localhost(:[0-9]+)?/",
    "example\\.com"
  ]
}
```

## Field reference

| Field                        | Type     | Default | Description                                           |
| ---------------------------- | -------- | ------- | ----------------------------------------------------- |
| `include`                    | string[] | `["."]` | Paths to lint when no CLI paths are provided.         |
| `ignore`                     | string[] | `[]`    | Path patterns to exclude.                             |
| `outputFormat`               | string   | `text`  | `text` or `json`.                                     |
| `minHeadingLevel`            | int      | `2`     | Minimum allowed heading level.                        |
| `enableHeadingLevelCheck`    | bool     | `true`  | Enable heading level validation.                      |
| `enableDuplicateHeadingCheck`| bool     | `true`  | Enable duplicate heading detection.                   |
| `enableLinkCheck`            | bool     | `false` | Enable external link checking.                        |
| `enableNoSetextHeadingsCheck`| bool     | `true`  | Disallow setext-style headings.                       |
| `skipLinkPatterns`           | string[] | `[]`    | Regex patterns for URLs to skip during link checking. |

## Notes

- CLI flags take precedence over config values.
- If no CLI paths are provided, `include` becomes the target set.
