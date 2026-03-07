---
title: "Output"
weight: 5
---

# Output

## Human-readable (`--output text`, default)

```sh
❯ gomarklint testdata/sample_links.md

Errors in testdata/sample_links.md:
  testdata/sample_links.md:1: First heading should be level 2 (found level 1)
  testdata/sample_links.md:4: Link unreachable: https://httpstat.us/404
  testdata/sample_links.md:12: Link unreachable: http://localhost-test:3001
  testdata/sample_links.md:16: duplicate heading: "overview"
  testdata/sample_links.md:18: image with empty alt text


✖ 5 issues found
✓ Checked 1 file(s), 19 line(s) in 757ms
```

- Summary: `✖ N issues found` if issues, `✔ No issues found` if clean.
- Always prints: `Checked <files>, <lines> in <Xms|Ys>`.

## JSON (`--output json`)

```json
{
  "files": 1,
  "lines": 19,
  "errors": 5,
  "elapsed_ms": 790,
  "details": {
    "testdata/sample_links.md": [
      { "File": "testdata/sample_links.md", "Line": 1, "Message": "First heading should be level 2 (found level 1)" },
      { "File": "testdata/sample_links.md", "Line": 4, "Message": "Link unreachable: https://httpstat.us/404" },
      { "File": "testdata/sample_links.md", "Line": 12, "Message": "Link unreachable: http://localhost-test:3001" },
      { "File": "testdata/sample_links.md", "Line": 16, "Message": "duplicate heading: \"overview\"" },
      { "File": "testdata/sample_links.md", "Line": 18, "Message": "image with empty alt text" }
    ]
  }
}
```

- `details` maps file path → list of issues (`file`, `line`, `message`).
- `elapsed_ms` is total wall time for the run.
