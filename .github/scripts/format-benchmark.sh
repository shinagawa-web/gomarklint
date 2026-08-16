#!/bin/bash
# Format benchmark comparison results as a markdown table

set -e

INPUT_FILE="$1"
OUTPUT_FILE="$2"
ENV_INFO="${3:-}"

if [[ ! -f "$INPUT_FILE" ]]; then
  echo "Error: Input file '$INPUT_FILE' does not exist" >&2
  exit 1
fi

if [[ ! -s "$INPUT_FILE" ]]; then
  echo "Warning: Input file '$INPUT_FILE' is empty. No benchmark comparison available." >&2
  echo "No benchmark comparison data available." > "$OUTPUT_FILE"
  exit 0
fi

awk -v env_info="$ENV_INFO" '
  BEGIN {
    time_old=""; time_new=""; time_delta=""; time_status=""
    asize_old=""; asize_new=""; asize_delta=""; asize_status=""
    alloc_old=""; alloc_new=""; alloc_delta=""; alloc_status=""
    in_cmd_pkg=0
    current_metric=""
  }

  /^pkg:/ {
    in_cmd_pkg = ($0 ~ /\/cmd$/) ? 1 : 0
    next
  }

  /│[[:space:]]*sec\/op[[:space:]]*│/    { current_metric = "time";   next }
  /│[[:space:]]*B\/op[[:space:]]*│/      { current_metric = "alloc_size"; next }
  /│[[:space:]]*allocs\/op[[:space:]]*│/ { current_metric = "allocs"; next }

  /^geomean/ && in_cmd_pkg {
    gsub(/±[[:space:]]*∞[[:space:]]*[¹²³⁴⁵⁶⁷⁸⁹⁰]*/, "")
    gsub(/\([^)]*\)[[:space:]]*[¹²³⁴⁵⁶⁷⁸⁹⁰]*/, "")

    old_val = $2
    new_val = $3
    delta   = $NF

    status = "✅"
    if (delta ~ /^\+[0-9.]+%$/) {
      num = delta
      sub(/^\+/, "", num); sub(/%$/, "", num)
      if (num + 0 >= 50)      status = "❌"
      else if (num + 0 >= 10) status = "⚠️"
    }

    if (current_metric == "time") {
      time_old=old_val; time_new=new_val; time_delta=delta; time_status=status
    } else if (current_metric == "alloc_size") {
      asize_old=old_val; asize_new=new_val; asize_delta=delta; asize_status=status
    } else if (current_metric == "allocs") {
      alloc_old=old_val; alloc_new=new_val; alloc_delta=delta; alloc_status=status
    }
  }

  END {
    if (time_old == "" && asize_old == "" && alloc_old == "") {
      print "No benchmark comparison data available."
      exit
    }

    print "| Metric | main | PR | Change |"
    print "|------------|-----:|---:|-------:|"
    if (time_old  != "") printf "| Exec time   | %s | %s | %s %s |\n", time_old,  time_new,  time_delta,  time_status
    if (asize_old != "") printf "| Alloc size  | %s | %s | %s %s |\n", asize_old, asize_new, asize_delta, asize_status
    if (alloc_old != "") printf "| Alloc count | %s | %s | %s %s |\n", alloc_old, alloc_new, alloc_delta, alloc_status

    if (env_info != "") {
      print ""
      print "> " env_info
    }
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"
