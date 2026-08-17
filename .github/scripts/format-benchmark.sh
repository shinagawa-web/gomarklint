#!/bin/bash
# Format benchmark comparison results as a markdown table

set -e

INPUT_FILE="$1"
OUTPUT_FILE="$2"
ENV_INFO="${3:-}"
NEW_BENCH_FILE="${4:-}"

# Extract PR benchmark order from new-bench.txt (strip Benchmark prefix, deduplicate)
PR_ORDER=""
if [[ -n "$NEW_BENCH_FILE" && -f "$NEW_BENCH_FILE" ]]; then
  PR_ORDER=$(grep '^Benchmark' "$NEW_BENCH_FILE" | awk '{print $1}' | sed 's/^Benchmark//' | awk '!seen[$0]++' | tr '\n' ',')
fi

if [[ ! -f "$INPUT_FILE" ]]; then
  echo "Error: Input file '$INPUT_FILE' does not exist" >&2
  exit 1
fi

if [[ ! -s "$INPUT_FILE" ]]; then
  echo "Warning: Input file '$INPUT_FILE' is empty. No benchmark comparison available." >&2
  echo "No benchmark comparison data available." > "$OUTPUT_FILE"
  exit 0
fi

awk -v env_info="$ENV_INFO" -v pr_order="$PR_ORDER" '
  function parse_val(v,    num, suffix) {
    num = v + 0
    suffix = v
    gsub(/^[0-9.]+/, "", suffix)
    if (suffix == "k")  return num * 1000
    if (suffix == "M")  return num * 1000000
    if (suffix == "Ki") return num * 1024
    if (suffix == "Mi") return num * 1048576
    if (suffix == "Gi") return num * 1073741824
    if (suffix == "m")  return num * 0.001
    if (suffix == "n")  return num * 1e-9
    return num
  }

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
      old_n = parse_val(old_val)
      new_n = parse_val(new_val)
      if (old_n != 0) {
        pct = (new_n - old_n) / old_n * 100
        delta = (pct >= 0) ? sprintf("+%.2f%%", pct) : sprintf("%.2f%%", pct)
      }
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

    print "| Benchmark | Metric | main | PR | Change |"
    print "|-----------|--------|-----:|---:|-------:|"

    if (pr_order != "") {
      n_ordered = split(pr_order, ordered_names, ",")
      for (i = 1; i <= n_ordered; i++) {
        name = ordered_names[i]
        if (name == "" || !(name in seen)) continue
        for (mi = 1; mi <= n_metrics; mi++) {
          mk = metric_keys[mi]
          if (data[name, mk, "old"] == "") continue
          printf "| %s | %s | %s | %s | %s%s |\n", \
            name, metric_labels[mi], data[name, mk, "old"], data[name, mk, "new"], \
            data[name, mk, "delta"], data[name, mk, "status"]
        }
      }
    } else {
      for (i = 0; i < bench_count; i++) {
        name = bench_names[i]
        for (mi = 1; mi <= n_metrics; mi++) {
          mk = metric_keys[mi]
          if (data[name, mk, "old"] == "") continue
          printf "| %s | %s | %s | %s | %s%s |\n", \
            name, metric_labels[mi], data[name, mk, "old"], data[name, mk, "new"], \
            data[name, mk, "delta"], data[name, mk, "status"]
        }
      }
    }

    if (env_info != "") {
      print ""
      print "> " env_info
    }
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"
