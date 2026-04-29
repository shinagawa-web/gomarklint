---
title: "Migrating from Other Linters"
weight: 12
---

# Migrating from Other Markdown Linters

This guide helps teams switching to gomarklint from markdownlint, remark-lint, or textlint. It covers installation replacement, rule mapping, configuration conversion, and CI migration.

**Why switch?**

- Single static binary — no Node.js or Python runtime required
- Fast concurrent linting via goroutines
- First-class CI integration with a native [GitHub Action](https://github.com/marketplace/actions/gomarklint-markdown-linter)
- Frontmatter-aware parsing (YAML/TOML stripped before linting)
- Live HTTP link validation built in (`external-link` rule)

---

## Installation

Remove your existing linter and install gomarklint:

```sh
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/shinagawa-web/gomarklint/main/install.sh | sh

# Homebrew
brew install shinagawa-web/tap/gomarklint

# npm
npm install -g @shinagawa-web/gomarklint

# go install
go install github.com/shinagawa-web/gomarklint/v2@latest
```

Remove the old linter from your project:

```sh
# if you used markdownlint-cli2:
npm uninstall markdownlint-cli2

# if you used remark-lint:
npm uninstall remark remark-cli remark-lint remark-preset-lint-recommended

# if you used textlint:
npm uninstall textlint textlint-rule-*
```

If you used `npm` only to run the linter, you may be able to remove `package.json` entirely and replace the script entry with the gomarklint binary.

---

## markdownlint

### markdownlint rule mapping

| markdownlint rule | gomarklint rule | Notes |
|---|---|---|
| MD001 `heading-increment` | `heading-level` | Checks heading level progression |
| MD003 `heading-style` | `no-setext-headings` | ATX enforcement only; setext detection is the default check |
| MD004 `ul-style` | `consistent-list-marker` | Options: `consistent` \| `dash` \| `asterisk` \| `plus` |
| MD009 `no-trailing-spaces` | — | Not yet implemented |
| MD010 `no-hard-tabs` | `no-hard-tabs` | — |
| MD012 `no-multiple-blanks` | `no-multiple-blank-lines` | — |
| MD013 `line-length` | `max-line-length` | Default **off**; set `lineLength` option |
| MD022 `blanks-around-headings` | `blanks-around-headings` | — |
| MD024 `no-duplicate-heading` | `duplicate-heading` | — |
| MD025 `single-h1` | `single-h1` | — |
| MD026 `no-trailing-punctuation` | `no-trailing-punctuation` | `punctuation` option configures the character set |
| MD031 `blanks-around-fences` | `blanks-around-fences` | — |
| MD032 `blanks-around-lists` | `blanks-around-lists` | — |
| MD034 `no-bare-urls` | `no-bare-urls` | — |
| MD036 `no-emphasis-as-heading` | `no-emphasis-as-heading` | Punctuation-ending spans are excluded |
| MD040 `fenced-code-language` | `fenced-code-language` | — |
| MD041 `first-line-heading` | — | No direct equivalent — `heading-level` checks level progression but does not require a heading on the first line |
| MD042 `no-empty-links` | `no-empty-links` | Also catches `[](#)` and `[](<>)` |
| MD045 `no-alt-text` | `empty-alt-text` | — |
| MD047 `single-trailing-newline` | `final-blank-line` | — |
| MD048 `code-fence-style` | `consistent-code-fence` | Options: `consistent` \| `backtick` \| `tilde` |
| MD049 `emphasis-style` | `consistent-emphasis-style` | Options: `consistent` \| `asterisk` \| `underscore` |
| MD051 `link-fragments` | `link-fragments` | Configurable slug algorithm; default **off** |
| MD052 `reference-links-images` | — | Not yet implemented |
| MD053 `link-image-style` | — | Not yet implemented |

Rules without a gomarklint equivalent yet: MD005–MD008 (list indent), MD011, MD014, MD018–MD021 (heading spaces), MD023, MD027–MD030, MD033, MD035, MD037–MD039, MD043, MD044, MD046, MD050, MD054–MD056, MD058, MD059.

### markdownlint config conversion

**Before** (`.markdownlint.json` / `.markdownlint-cli2.jsonc`):

```json
{
  "default": true,
  "MD013": { "line_length": 120 },
  "MD024": false,
  "MD026": { "punctuation": ".,;:!?" }
}
```

**After** (`.gomarklint.json`):

```json
{
  "default": true,
  "rules": {
    "max-line-length": { "enabled": true, "lineLength": 120 },
    "duplicate-heading": false,
    "no-trailing-punctuation": { "punctuation": ".,;:!?" }
  }
}
```

Key differences:

- Rule keys are descriptive names (`no-trailing-punctuation`) rather than MD numbers (`MD026`).
- All rule configuration lives under `rules.<rule-key>`.
- `false` disables a rule; `true` enables it with defaults; an object enables it with options.

### markdownlint CI migration

**Before** (GitHub Actions with markdownlint-cli2):

```yaml
- uses: DavidAnson/markdownlint-cli2-action@v18
  with:
    globs: "**/*.md"
```

**After** (gomarklint):

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: 'go.mod'

- uses: shinagawa-web/gomarklint-action@v1
```

---

## remark-lint

### remark-lint rule mapping

| remark-lint rule | gomarklint rule | Notes |
|---|---|---|
| `final-newline` | `final-blank-line` | — |
| `no-consecutive-blank-lines` | `no-multiple-blank-lines` | — |
| `heading-increment` | `heading-level` | — |
| `first-heading-level` | — | No direct equivalent — `heading-level` checks level progression but does not enforce position |
| `no-duplicate-headings` | `duplicate-heading` | — |
| `heading-style` (atx) | `no-setext-headings` | — |
| `no-missing-blank-lines` | `blanks-around-headings`, `blanks-around-lists`, `blanks-around-fences` | remark-lint combines these; gomarklint has separate rules |
| `no-literal-urls` | `no-bare-urls` | — |
| `no-empty-url` | `no-empty-links` | — |
| `no-emphasis-as-heading` | `no-emphasis-as-heading` | — |
| `fenced-code-flag` | `fenced-code-language` | — |
| `no-multiple-toplevel-headings` | `single-h1` | — |
| `maximum-line-length` | `max-line-length` | Default **off** |
| `no-tabs` | `no-hard-tabs` | — |
| `no-heading-punctuation` | `no-trailing-punctuation` | — |
| `fenced-code-marker` | `consistent-code-fence` | — |
| `emphasis-marker` | `consistent-emphasis-style` | — |
| `unordered-list-marker-style` | `consistent-list-marker` | — |
| `no-undefined-references` | — | Not yet implemented (Priority 3) |
| `hard-break-spaces` | — | `no-trailing-spaces` not yet implemented |
| `linebreak-style` | — | `consistent-line-endings` not yet implemented |

### remark-lint config conversion

**Before** (`.remarkrc.mjs`):

```js
import remarkPresetLintRecommended from 'remark-preset-lint-recommended'
import remarkLintMaximumLineLength from 'remark-lint-maximum-line-length'
import remarkLintEmphasisMarker from 'remark-lint-emphasis-marker'

export default {
  plugins: [
    remarkPresetLintRecommended,
    [remarkLintMaximumLineLength, 120],
    [remarkLintEmphasisMarker, '*'],
  ],
}
```

**After** (`.gomarklint.json`):

```json
{
  "default": true,
  "rules": {
    "max-line-length": { "enabled": true, "lineLength": 120 },
    "consistent-emphasis-style": { "style": "asterisk" }
  }
}
```

### remark-lint CI migration

**Before** (GitHub Actions with remark-cli):

```yaml
- name: Install remark
  run: npm install -g remark-cli remark-preset-lint-recommended

- name: Lint Markdown
  run: remark --use remark-preset-lint-recommended .
```

**After**:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: 'go.mod'

- uses: shinagawa-web/gomarklint-action@v1
```

---

## textlint

textlint focuses primarily on prose style (word choice, grammar, writing conventions). gomarklint focuses on Markdown structure. The overlap is narrow — only structural rules have equivalents.

### textlint rule mapping

| textlint rule / plugin | gomarklint rule | Notes |
|---|---|---|
| `@textlint-rule/no-duplicate-heading` | `duplicate-heading` | — |
| `textlint-rule-no-empty-element` | `no-empty-links`, `empty-alt-text` | Partial overlap |
| `textlint-rule-no-todo` | — | No equivalent |
| `textlint-rule-ja-*` | — | Japanese prose rules; no equivalent |
| `textlint-rule-spellcheck-tech-word` | — | No equivalent |

Most textlint rules (terminology, spacing, punctuation style for prose) have no structural equivalent in gomarklint. If you use textlint for writing quality checks, you can run both tools side by side during a transition period.

### Running both tools in parallel

```json
{
  "scripts": {
    "lint:structure": "gomarklint .",
    "lint:prose": "textlint **/*.md",
    "lint": "npm run lint:structure && npm run lint:prose"
  }
}
```

---

## Handling uncovered rules

Rules not yet implemented in gomarklint (from the roadmap):

| Planned rule | markdownlint equivalent | remark-lint equivalent | Priority |
|---|---|---|---|
| `no-trailing-spaces` | MD009 | `hard-break-spaces` | Priority 3 |
| `no-undefined-references` | MD052/MD053 | `no-undefined-references` | Priority 3 |
| `consistent-line-endings` | — | `linebreak-style` | Priority 3 |
| `table-formatting` | MD055/MD056/MD058 | `table-pipes`, `table-cell-padding` | Priority 3 |
| `descriptive-link-text` | MD059 | — | Priority 3 |

To request new rules or track progress, see [issue #76](https://github.com/shinagawa-web/gomarklint/issues/76).

---

## FAQ

**Can I run gomarklint alongside markdownlint during migration?**

Yes. Both tools can run independently on the same files. Use severity `warning` in gomarklint while you fix violations:

```json
{
  "default": true,
  "rules": {
    "no-bare-urls": { "severity": "warning" }
  }
}
```

**How do I gradually adopt gomarklint?**

Start with `default: false` and enable only the rules you want to enforce first. Add more rules incrementally as your team addresses violations:

```json
{
  "default": false,
  "rules": {
    "final-blank-line": true,
    "fenced-code-language": true
  }
}
```

**Does gomarklint support inline disable comments?**

Yes. Use `<!-- gomarklint-disable -->` and `<!-- gomarklint-enable -->` to suppress violations for a block, or `<!-- gomarklint-disable rule-name -->` to disable a specific rule. See the [disable comments]({{< relref "disable-comments.md" >}}) page.

**Does gomarklint support per-directory config files?**

Not automatically. markdownlint-cli2 traverses directories and applies the nearest `.markdownlint.json` it finds — gomarklint does not do this yet. The workaround is to exclude the subdirectory from the root config and run gomarklint a second time with an explicit `--config`:

```json
{
  "ignore": ["docs/"]
}
```

```sh
gomarklint .                                   # root config
gomarklint docs/ --config docs/.gomarklint.json  # docs-specific config
```

**Does gomarklint support shared configs across repositories?**

Not yet. There is no `extends` mechanism to inherit from an npm package or a remote config file. Each repository maintains its own `.gomarklint.json`.
