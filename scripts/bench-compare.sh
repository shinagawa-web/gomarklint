#!/bin/bash
# Compare benchmarks between the current branch and origin/main.
# Exits non-zero when the formatted benchmark comparison reports a regression
# of ≥+10% (⚠️) or ≥+50% (❌) for any metric it flags.
set -euo pipefail

if [[ -n "$(git status --porcelain)" ]]; then
    echo "Error: working tree has uncommitted or untracked changes. Commit or stash before running bench-compare." >&2
    exit 1
fi

BENCHSTAT="$(command -v benchstat || true)"
if [[ -z "$BENCHSTAT" ]]; then
    _GOBIN="$(go env GOBIN)"
    if [[ -n "$_GOBIN" && -x "$_GOBIN/benchstat" ]]; then
        BENCHSTAT="$_GOBIN/benchstat"
    else
        _FIRST_GOPATH="${GOPATH%%:*}"
        _FIRST_GOPATH="${_FIRST_GOPATH:-$(go env GOPATH)}"
        if [[ -n "$_FIRST_GOPATH" && -x "$_FIRST_GOPATH/bin/benchstat" ]]; then
            BENCHSTAT="$_FIRST_GOPATH/bin/benchstat"
        fi
    fi
fi
if [[ -z "$BENCHSTAT" || ! -x "$BENCHSTAT" ]]; then
    echo "Error: benchstat not found. Install it with:" >&2
    echo "  go install golang.org/x/perf/cmd/benchstat@latest" >&2
    exit 1
fi

PKGS=$(go list ./... | grep -v '/e2e')
ORIGINAL_REF=$(git symbolic-ref --short HEAD 2>/dev/null || git rev-parse HEAD)

NEW_BENCH=$(mktemp "${TMPDIR:-/tmp}/gomarklint-bench.XXXXXX")
OLD_BENCH=$(mktemp "${TMPDIR:-/tmp}/gomarklint-bench.XXXXXX")
RAW_CMP=$(mktemp "${TMPDIR:-/tmp}/gomarklint-bench.XXXXXX")
FMT_CMP=$(mktemp "${TMPDIR:-/tmp}/gomarklint-bench.XXXXXX")
CHECKED_OUT_MAIN=false

cleanup() {
    rm -f "$NEW_BENCH" "$OLD_BENCH" "$RAW_CMP" "$FMT_CMP"
    if $CHECKED_OUT_MAIN; then
        git checkout --quiet "$ORIGINAL_REF" 2>/dev/null || true
    fi
}
trap cleanup EXIT

echo "==> Running benchmarks on $ORIGINAL_REF..."
# shellcheck disable=SC2086
go test -bench=. -benchmem $PKGS -run='^$' > "$NEW_BENCH"

echo "==> Fetching origin/main for baseline..."
if ! git fetch --quiet origin main; then
    echo "Error: failed to fetch origin/main." >&2
    exit 1
fi

echo "==> Checking out origin/main for baseline..."
git checkout --quiet origin/main
CHECKED_OUT_MAIN=true
go mod download 2>/dev/null

echo "==> Running benchmarks on origin/main..."
# shellcheck disable=SC2086
go test -bench=. -benchmem $PKGS -run='^$' > "$OLD_BENCH"

echo "==> Returning to $ORIGINAL_REF..."
git checkout --quiet "$ORIGINAL_REF"
CHECKED_OUT_MAIN=false

echo "==> Comparing benchmarks (baseline: main, candidate: $ORIGINAL_REF)..."
"$BENCHSTAT" "$OLD_BENCH" "$NEW_BENCH" > "$RAW_CMP" 2>&1 || true
bash .github/scripts/format-benchmark.sh "$RAW_CMP" "$FMT_CMP"
cat "$FMT_CMP"

if grep -qE '⚠️|❌' "$FMT_CMP"; then
    echo ""
    echo "FAIL: benchmark regression detected (⚠️ ≥+10% or ❌ ≥+50%)." >&2
    exit 1
fi
echo "==> Benchmark comparison OK."
