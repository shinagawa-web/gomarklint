# demo fixture

`sample.md` **intentionally** contains lint violations (a broken anchor link,
an image with empty alt text, and a fenced code block without a language).
They are what the recorded demo GIF shows being caught — do **not** "fix" them.

The GIF is regenerated with `make demo` (see `demo.tape` and
`scripts/record-demo.sh`) and committed to `docs/static/demo.gif`.

This directory is outside the self-lint scope (`README.md`, `README.ja.md`,
`docs/content` — see `.gomarklint.ci.json`), so the violations don't fail CI.
