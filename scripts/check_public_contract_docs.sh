#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

targets=(
  README.md
  docs
  docs/ja
)

declare -a patterns=(
  "--schema"
  "schema v1"
  "schema v2"
  "\\.outbound\\[\\]"
  "\\.other\\."
  "task_\\*"
  "other_\\*"
  "parent_title"
)

status=0
for pattern in "${patterns[@]}"; do
  if rg -n -e "${pattern}" "${targets[@]}"; then
    echo
    echo "public contract docs guard failed on pattern: ${pattern}" >&2
    status=1
  fi
done

exit "${status}"
