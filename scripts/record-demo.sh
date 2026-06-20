#!/usr/bin/env bash
# Record the README/docs demo GIF with VHS (https://github.com/charmbracelet/vhs).
#
#   ./scripts/record-demo.sh      # or: make demo
#
# Requires: vhs (brew install vhs).
set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v vhs >/dev/null 2>&1; then
  echo "error: vhs is not installed. Install it with: brew install vhs" >&2
  exit 1
fi

echo "==> Building gomarklint..."
go build -o gomarklint .

echo "==> Recording demo -> docs/static/demo.gif"
vhs demo.tape

echo "==> Done. Review docs/static/demo.gif and commit it if it looks good."
