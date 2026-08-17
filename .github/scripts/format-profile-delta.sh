#!/bin/bash
# Append CPU profile delta sections for benchmarks with time/op regression (⚠️ or ❌)

set -e

COMPARISON_FILE="$1"
PROFILES_DIR="${2:-profiles}"

# Find regressing benchmarks: time/op rows with explicit ⚠️ or ❌ marker
REGRESSING=$(grep 'time/op' "$COMPARISON_FILE" | grep -E '⚠️|❌' | awk -F'|' '{
  name = $2
  gsub(/^[[:space:]]+|[[:space:]]+$/, "", name)
  gsub(/-[0-9]+$/, "", name)
  print "Benchmark" name
}')

[[ -z "$REGRESSING" ]] && exit 0

echo ""
echo "---"

for bench in $REGRESSING; do
  old_prof="${PROFILES_DIR}/old-${bench}.prof"
  new_prof="${PROFILES_DIR}/new-${bench}.prof"
  new_bin="${PROFILES_DIR}/new-${bench}.test"

  [[ ! -f "$old_prof" || ! -f "$new_prof" ]] && continue

  short="${bench#Benchmark}"
  echo ""
  echo "### CPU Profile Delta — \`${short}\`"
  echo ""

  # Build pprof args — include binary only if it exists
  PPROF_ARGS_CUM=("-top" "-cum" "-base=${old_prof}")
  PPROF_ARGS_FLAT=("-top" "-base=${old_prof}")
  if [[ -f "$new_bin" ]]; then
    PPROF_ARGS_CUM+=("$new_bin")
    PPROF_ARGS_FLAT+=("$new_bin")
  fi
  PPROF_ARGS_CUM+=("$new_prof")
  PPROF_ARGS_FLAT+=("$new_prof")

  echo "**Which rule got slower** — gomarklint functions, sorted by cum delta:"
  echo ""
  echo "| Function | Δ cum |"
  echo "|---|---:|"
  go tool pprof "${PPROF_ARGS_CUM[@]}" 2>/dev/null | \
    awk 'NR>5 && NF>=6 && /(gomarklint|shinagawa-web)/ && $4 ~ /^[0-9]/ {
      name = ($NF == "(inline)") ? $(NF-1) : $NF
      printf "| %s | %s |\n", name, $4
    }' | head -10
  echo ""

  echo "**What operation got more expensive** — all functions, sorted by flat delta:"
  echo ""
  echo "| Function | Δ flat |"
  echo "|---|---:|"
  go tool pprof "${PPROF_ARGS_FLAT[@]}" 2>/dev/null | \
    awk 'NR>5 && NF>=6 && $1 ~ /^[0-9]/ {
      name = ($NF == "(inline)") ? $(NF-1) : $NF
      printf "| %s | %s |\n", name, $1
    }' | head -10
  echo ""
done
