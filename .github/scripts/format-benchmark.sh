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
    in_cmd_pkg = 0
    current_metric = ""
    bench_count = 0
  }

  /^pkg:/ {
    in_cmd_pkg = ($0 ~ /\/cmd$/) ? 1 : 0
    next
  }

  /│[[:space:]]*sec\/op[[:space:]]*│/    { current_metric = "time";   next }
  /│[[:space:]]*B\/op[[:space:]]*│/      { current_metric = "memory"; next }
  /│[[:space:]]*allocs\/op[[:space:]]*│/ { current_metric = "allocs"; next }

  /^[A-Z]/ && in_cmd_pkg && current_metric != "" {
    gsub(/±[[:space:]]*[0-9.]+%[[:space:]]*/, "")
    gsub(/±[[:space:]]*∞[[:space:]]*[¹²³⁴⁵⁶⁷⁸⁹⁰]*/, "")
    gsub(/\([^)]*\)[[:space:]]*[¹²³⁴⁵⁶⁷⁸⁹⁰]*/, "")

    if (NF < 4) next  # present in only one file; skip

    name = $1
    old_val = $2
    new_val = $3
    delta = $NF

    status = ""
    if (delta ~ /^\+[0-9.]+%$/) {
      num = delta
      sub(/^\+/, "", num); sub(/%$/, "", num)
      if (num + 0 >= 50)      status = " ❌"
      else if (num + 0 >= 10) status = " ⚠️"
      else                    status = " ✅"
    } else if (delta ~ /^-[0-9.]+%$/) {
      status = " ✅"
    } else if (delta == "~") {
      status = " ✅"
    }

    data[name, current_metric, "old"]    = old_val
    data[name, current_metric, "new"]    = new_val
    data[name, current_metric, "delta"]  = delta
    data[name, current_metric, "status"] = status

    if (!(name in seen)) {
      seen[name] = 1
      bench_names[bench_count++] = name
    }
    next
  }

  END {
    if (bench_count == 0) {
      print "No benchmark comparison data available."
      exit
    }

    n_metrics = split("time memory allocs", metric_keys, " ")
    split("time/op B/op allocs/op", metric_labels, " ")

    for (mi = 1; mi <= n_metrics; mi++) {
      mk = metric_keys[mi]
      ml = metric_labels[mi]

      has_data = 0
      for (i = 0; i < bench_count; i++) {
        if (data[bench_names[i], mk, "old"] != "") { has_data = 1; break }
      }
      if (!has_data) continue

      print ""
      print "**" ml "**"
      print ""
      print "| Benchmark | main | PR | Change |"
      print "|-----------|-----:|---:|-------:|"
      for (i = 0; i < bench_count; i++) {
        name = bench_names[i]
        if (data[name, mk, "old"] == "") continue
        printf "| %s | %s | %s | %s%s |\n", \
          name, data[name, mk, "old"], data[name, mk, "new"], \
          data[name, mk, "delta"], data[name, mk, "status"]
      }
    }

    if (env_info != "") {
      print ""
      print "> " env_info
    }
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"
