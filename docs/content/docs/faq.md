---
title: "FAQ & Troubleshooting"
weight: 11
---

# FAQ & Troubleshooting

Quick answers to the most common problems. Each entry gives an immediate fix and points to deeper docs where useful.

---

## Configuration

### `unknown field: <field-name>`

You are using a v1 config file with a v2 binary. v2 replaced the flat boolean fields (`enableLinkCheck`, `minHeadingLevel`, etc.) with a unified `rules` map.

Run `gomarklint init` to generate a fresh v2 config, then migrate your settings.

→ [Migrating to v2](../migration-v2/)

---

### `failed to parse config file`

The config file contains invalid JSON or an unsupported rule value. Common causes:

- Trailing comma in JSON
- Rule value is a string other than `"error"`, `"warning"`, or `"off"`
- Numeric value where a string or boolean is expected

→ [Configuration](../configuration/)

---

### `invalid rule value`

Rule severity must be one of: `true`, `false`, `"error"`, `"warning"`, or `"off"`.

→ [Configuration](../configuration/)

---

### `invalid value "x" for rule.style`

A `style` option was set to an unrecognised value. gomarklint validates rule options at startup and exits immediately — it does not silently fall back to a default.

Valid values per rule:

| Rule | Valid `style` values |
|---|---|
| `consistent-code-fence` | `consistent`, `backtick`, `tilde` |
| `consistent-emphasis-style` | `consistent`, `asterisk`, `underscore` |
| `consistent-list-marker` | `consistent`, `dash`, `asterisk`, `plus` |

A common mistake is using a plural form (`"backticks"`, `"tildes"`) or a synonym (`"bullet"`). Check the exact spelling shown above.

→ [Rules](../rules/)

---

### `failed to access config file`

The file at the path passed to `--config` exists but cannot be accessed (for example, a permission error on the parent directory). If the file does not exist, `LoadOrDefault` falls back to the default configuration without an error. Check the file permissions and verify the file is committed if running in CI.

→ [CLI Reference](../cli/)

---

### `unknown flag: --enable-*`

The `--enable-*` flags were removed in v2. Use the `rules` map in your config file instead.

→ [Migrating to v2](../migration-v2/)

---

### Apply different rules to different directories

gomarklint does not automatically pick up config files in subdirectories. Use `ignore` in the root config to exclude the subdirectory, then run gomarklint again with an explicit `--config`:

```json
{
  "ignore": ["docs/"]
}
```

```sh
gomarklint .                                     # root config for everything else
gomarklint docs/ --config docs/.gomarklint.json  # docs-specific config
```

In CI, add both commands as separate steps.

---

## Running the linter

### `unknown shorthand flag: 'v'`

The `-v` shorthand is not yet supported. Use `--version` instead. (Tracking: #152)

---

### No violations reported

Three common causes:

1. **Wrong path** — the glob or directory passed to gomarklint does not match your files.
2. **Rule disabled** — the rule is set to `false` or `"off"` in your config.
3. **Severity filter** — if you are using `--severity`, violations below the threshold are suppressed.

→ [Configuration](../configuration/), [CLI Reference](../cli/)

---

### Lint passes locally but fails in CI

Usually a version mismatch or a config file that is not committed. Check that:

- The same gomarklint version is pinned in CI as you use locally.
- `.gomarklint.json` (or your config file) is committed to the repository.

→ [GitHub Actions Integration](../github-actions/)

---

## External link checking

### Link check is slow or timing out

Increase `timeoutSeconds` in your config, or skip known-slow domains with `skipPatterns`.

→ [Configuration](../configuration/)

---

### Valid links reported as broken (403 errors)

Go's default `User-Agent` (`Go-http-client/1.1`) is blocked by CDN-level bot protection (Cloudflare, Akamai, and others). The server returns 403 — not because the link is broken, but because it refuses automated clients.

**Workarounds:**

- Add the domain to `skipPatterns` to skip it during external link checks.
- If only a few links are affected, verify them manually before treating the report as a real broken link.

→ [Configuration](../configuration/)

---

### Academic or journal links are always flagged (doi.org, Wikipedia, etc.)

These sites aggressively block automated HTTP clients regardless of User-Agent. `skipPatterns` is the most reliable workaround:

```json
{
  "skipPatterns": ["doi\\.org", "wikipedia\\.org"]
}
```

External link checking has inherent limits — the tool flags candidates, but final judgment stays with the author.

→ [Configuration](../configuration/)

---

### Not sure if a link is truly broken

Before opening an issue, verify the link using the same User-Agent gomarklint uses:

```sh
curl -s -o /dev/null -w "%{http_code}\n" \
  -A "Go-http-client/1.1" \
  https://your-url-here
```

| Status | Meaning |
|--------|---------|
| 2xx | Reachable — the report is a false-positive |
| 3xx | Redirect — still reachable |
| 403 | Bot protection blocking automated clients — not a broken link |
| 404 | Genuinely broken link |
| 000 / connection error | Network issue or domain does not exist |

If the curl output shows 2xx or 3xx, the link is fine. Include the curl output when filing an issue about a suspected false-positive.
