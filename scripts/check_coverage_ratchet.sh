#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
baseline_file="${repo_root}/scripts/coverage_baseline.txt"
tmp_output="$(mktemp)"
trap 'rm -f "${tmp_output}"' EXIT

if [[ ! -f "${baseline_file}" ]]; then
  echo "coverage baseline file not found: ${baseline_file}" >&2
  exit 1
fi

go test ./... -cover | tee "${tmp_output}"

declare -A actual
while read -r pkg pct; do
  actual["${pkg}"]="${pct}"
done < <(
  awk '
    /^ok[[:space:]]/ && /coverage:/ {
      pkg = $2
      for (i = 1; i <= NF; i++) {
        if ($i == "coverage:") {
          pct = $(i + 1)
          gsub("%", "", pct)
          print pkg, pct
          break
        }
      }
    }
  ' "${tmp_output}"
)

status=0
while read -r pkg min; do
  [[ -z "${pkg}" ]] && continue
  [[ "${pkg}" == \#* ]] && continue

  got="${actual[${pkg}]:-}"
  if [[ -z "${got}" ]]; then
    echo "missing coverage result for ${pkg}" >&2
    status=1
    continue
  fi

  if awk -v got="${got}" -v min="${min}" 'BEGIN { exit !((got + 0.0) < (min + 0.0)) }'; then
    echo "coverage regression: ${pkg} ${got}% < ${min}%" >&2
    status=1
  fi
done < "${baseline_file}"

exit "${status}"
