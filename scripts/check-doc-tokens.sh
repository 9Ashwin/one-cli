#!/usr/bin/env bash
# Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT
#
# check-doc-tokens.sh
#
# Scans docs/examples for token-like values that look realistic but are not
# using the required placeholder format (*_EXAMPLE_TOKEN or similar).
#
# Docs MUST use clearly fake placeholders, e.g.:
#   your_token_here   <api_key>   XXXXEXAMPLE
#
# If this check fails, replace the realistic-looking value with a placeholder
# so gitleaks CI won't flag it as a real secret.

set -euo pipefail

SCAN_DIR="${1:-docs}"
ERRORS=0

# Generic high-entropy token shapes (long base64/hex runs). Real platform
# token prefixes get added here as one-cli integrates new platforms.
TOKEN_BODY='[A-Za-z0-9_-]{32,}'
REALISTIC_TOKEN_RE="\"${TOKEN_BODY}\"|\`${TOKEN_BODY}\`"
PLACEHOLDER_RE='(EXAMPLE|_TOKEN|XXXX|xxxx|<|>|your_|_here|placeholder)'

if [ ! -d "$SCAN_DIR" ]; then
  echo "ℹ️  $SCAN_DIR does not exist yet — nothing to scan."
  exit 0
fi

while IFS= read -r -d '' file; do
  matches=$(grep -nEo "$REALISTIC_TOKEN_RE" "$file" 2>/dev/null | grep -vE "$PLACEHOLDER_RE" || true)
  if [[ -n "$matches" ]]; then
    echo ""
    echo "❌  $file"
    echo "    Contains realistic-looking token values that may trigger gitleaks:"
    while IFS= read -r line; do
      echo "      $line"
    done <<< "$matches"
    echo "    → Replace with a placeholder, e.g. your_token_here or <api_key>"
    ERRORS=$((ERRORS + 1))
  fi
done < <(find "$SCAN_DIR" -type f -name '*.md' -print0)

if [[ $ERRORS -gt 0 ]]; then
  echo ""
  echo "Found $ERRORS file(s) with realistic-looking tokens. Fix before pushing."
  exit 1
fi

echo "✅ No realistic-looking tokens found in $SCAN_DIR."
