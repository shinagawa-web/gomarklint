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

# Add status symbols based on delta percentage
awk '
  NR==1 {print; next}  # Print header
  /^name/ {print; next}  # Print column names
  {
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
    
    print $0 status
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"
