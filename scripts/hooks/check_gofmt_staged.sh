#!/usr/bin/env bash
set -euo pipefail

mapfile -t files < <(git diff --cached --name-only --diff-filter=ACM | grep -E '\.go$' || true)
if [[ ${#files[@]} -eq 0 ]]; then
  exit 0
fi

unformatted="$(gofmt -l "${files[@]}")"
if [[ -n "${unformatted}" ]]; then
  echo "The following staged Go files are not gofmt'ed:"
  echo "${unformatted}"
  echo "Run: gofmt -w <files> && git add <files>"
  exit 1
fi
