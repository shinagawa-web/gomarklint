#!/bin/sh
set -uo pipefail

CONFIG_FILE=".gomarklint.json"

if [ -f "$CONFIG_FILE" ]; then
  echo "Found $CONFIG_FILE"
else
  echo "$CONFIG_FILE not found. This action requires a config file."
  exit 1
fi

# Run gomarklint and capture output + exit code
OUTPUT=$(gomarklint "$@" 2>&1) || true
EXIT_CODE=${PIPESTATUS:-$?}

# Re-run to get the real exit code since capturing swallows it
gomarklint "$@" > /dev/null 2>&1
EXIT_CODE=$?

echo "$OUTPUT"

# Post PR comment if enabled
if [ "${INPUT_COMMENT_ON_PR:-false}" = "true" ]; then
  if [ -z "${INPUT_GITHUB_TOKEN:-}" ]; then
    echo "Warning: github-token is required for comment-on-pr. Skipping comment."
  elif [ -z "${GITHUB_EVENT_NAME:-}" ] || [ "$GITHUB_EVENT_NAME" != "pull_request" ]; then
    echo "Not a pull_request event. Skipping comment."
  else
    MARKER="<!-- gomarklint-result -->"
    REPO="${GITHUB_REPOSITORY}"

    # Extract PR number from the event payload
    PR_NUMBER=$(jq -r '.pull_request.number' "$GITHUB_EVENT_PATH")

    if [ -z "$PR_NUMBER" ] || [ "$PR_NUMBER" = "null" ]; then
      echo "Could not determine PR number. Skipping comment."
    else
      # Build comment body
      if [ $EXIT_CODE -eq 0 ]; then
        BODY=$(printf '%s\n### gomarklint result\n\nNo issues found.\n' "$MARKER")
      else
        BODY=$(printf '%s\n### gomarklint result\n\n```\n%s\n```\n' "$MARKER" "$OUTPUT")
      fi

      # Search for existing comment with marker
      EXISTING_COMMENT_ID=$(curl -s \
        -H "Authorization: token ${INPUT_GITHUB_TOKEN}" \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/${REPO}/issues/${PR_NUMBER}/comments?per_page=100" \
        | jq -r ".[] | select(.body | startswith(\"${MARKER}\")) | .id" \
        | head -1)

      if [ -n "$EXISTING_COMMENT_ID" ] && [ "$EXISTING_COMMENT_ID" != "null" ]; then
        # Update existing comment
        curl -s -X PATCH \
          -H "Authorization: token ${INPUT_GITHUB_TOKEN}" \
          -H "Accept: application/vnd.github.v3+json" \
          "https://api.github.com/repos/${REPO}/issues/comments/${EXISTING_COMMENT_ID}" \
          -d "$(jq -n --arg body "$BODY" '{body: $body}')" > /dev/null
        echo "Updated existing PR comment."
      else
        # Create new comment
        curl -s -X POST \
          -H "Authorization: token ${INPUT_GITHUB_TOKEN}" \
          -H "Accept: application/vnd.github.v3+json" \
          "https://api.github.com/repos/${REPO}/issues/${PR_NUMBER}/comments" \
          -d "$(jq -n --arg body "$BODY" '{body: $body}')" > /dev/null
        echo "Posted new PR comment."
      fi
    fi
  fi
fi

exit $EXIT_CODE
