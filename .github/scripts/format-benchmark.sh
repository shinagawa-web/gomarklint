#!/bin/bash
# Format benchmark comparison results with status symbols

set -e

INPUT_FILE="$1"
OUTPUT_FILE="$2"

# Validate input file exists
if [[ ! -f "$INPUT_FILE" ]]; then
  echo "Error: Input file '$INPUT_FILE' does not exist" >&2
  exit 1
fi

# Check if input file has content
if [[ ! -s "$INPUT_FILE" ]]; then
  echo "Warning: Input file '$INPUT_FILE' is empty. No benchmark comparison available." >&2
  echo "No benchmark comparison data available." > "$OUTPUT_FILE"
  exit 0
fi

# Extract only geomean (summary) results for cleaner output
awk '
  # Print system info
  /^goos:|^goarch:|^cpu:/ {
    print
    next
  }
  
  # Track which package we are in
  /^pkg:/ {
    current_pkg = $0
    # Only process cmd package
    if ($0 ~ /\/cmd$/) {
      in_cmd_pkg = 1
      print
    } else {
      in_cmd_pkg = 0
    }
    next
  }
  
  # Track which metric we are in (sec/op, B/op, allocs/op)
  /│[[:space:]]*sec\/op[[:space:]]*│/ {
    current_metric = "time"
    next
  }
  /│[[:space:]]*B\/op[[:space:]]*│/ {
    current_metric = "memory"
    next
  }
  /│[[:space:]]*allocs\/op[[:space:]]*│/ {
    current_metric = "allocs"
    next
  }
  
  # Process only geomean lines (summary statistics) for cmd package
  /^geomean/ {
    # Skip if not in cmd package
    if (!in_cmd_pkg) {
      next
    }
    
    # Remove statistical annotations like ± ∞ ¹
    gsub(/±[[:space:]]*∞[[:space:]]*[¹²³⁴⁵⁶⁷⁸⁹⁰]*/, "")
    # Remove statistical notes like (p=... n=...) ²
    gsub(/\([^)]*\)[[:space:]]*[¹²³⁴⁵⁶⁷⁸⁹⁰]*/, "")
    
    # Extract delta percentage from last field
    delta = $NF
    status = ""
    
    # Check for positive change (+X.XX%)
    if (delta ~ /^\+[0-9.]+%$/) {
      # Extract the number
      sub(/^\+/, "", delta)
      sub(/%$/, "", delta)
      percent = delta + 0
      
      if (percent >= 50) {
        status = " ❌"
      } else if (percent >= 10) {
        status = " ⚠️"
      } else {
        status = " ✅"
      }
    } 
    # Check for negative change (-X.XX%)
    else if (delta ~ /^-[0-9.]+%$/) {
      status = " ✅"  # Faster is good
    } 
    # No change
    else if (delta == "~") {
      status = " ✅"
    }
    
    # Add metric label
    metric_label = ""
    if (current_metric == "time") {
      metric_label = " [time/op]"
    } else if (current_metric == "memory") {
      metric_label = " [memory/op]"
    } else if (current_metric == "allocs") {
      metric_label = " [allocs/op]"
    }
    
    print $0 status metric_label
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"
