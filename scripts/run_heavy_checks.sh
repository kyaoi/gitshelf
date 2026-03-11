#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

echo "[heavy] go test ./..."
go test ./...

echo "[heavy] bash scripts/check_coverage_ratchet.sh"
bash scripts/check_coverage_ratchet.sh

echo "[heavy] go test -race ./..."
go test -race ./...

echo "[heavy] go vet ./..."
go vet ./...
