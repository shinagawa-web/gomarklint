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

# Filter and format benchmark results
# Remove unnecessary statistical notes and keep only meaningful metrics
awk '
  # Skip noise lines (statistical notes)
  /^¹/ || /^²/ || /^³/ {next}
  
  # Print system info and headers
  /^goos:|^goarch:|^pkg:|^cpu:/ {print; next}
  
  # Print table headers
  /^name/ || /│.*sec\/op.*│/ || /│.*B\/op.*│/ || /│.*allocs\/op.*│/ {
    # Skip if this is a continuation of previous table header
    if (prev_was_header && /^[[:space:]]*│/) {next}
    print
    prev_was_header = (/^name/ || /│.*sec\/op.*│/ || /│.*B\/op.*│/ || /│.*allocs\/op.*│/)
    next
  }
  
  # Process benchmark result lines
  {
    prev_was_header = 0
    
    # Skip empty lines
    if (NF == 0) {next}
    
    delta = $NF
    status = ""
    
    # Extract percentage from delta (e.g., "+5.03%" -> 5.03)
    if (match(delta, /\+([0-9.]+)%/, arr)) {
      percent = arr[1]
      if (percent >= 50) {
        status = " ❌"
      } else if (percent >= 10) {
        status = " ⚠️"
      } else {
        status = " ✅"
      }
    } else if (match(delta, /-([0-9.]+)%/, arr)) {
      status = " ✅"  # Faster is good
    } else if (delta == "~") {
      status = " ✅"  # No change is good
    }
    
    # Only print lines with actual benchmark data
    if (status != "") {
      print $0 status
    }
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"
