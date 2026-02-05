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
  /^goos:|^goarch:|^pkg:|^cpu:/ {
    print
    next
  }
  
  # Print table separator headers for context
  /│.*vs base.*│/ {
    if (!header_printed) {
      print
      header_printed = 1
    }
    next
  }
  
  # Process only geomean lines (summary statistics)
  /^geomean/ {
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
    
    print $0 status
    print ""  # Add blank line for readability
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"
