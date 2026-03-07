#!/usr/bin/env bash
set -euo pipefail

SHELF_ROOT="${SHELF_ROOT:-$(pwd)}"
SHELF_BIN="${SHELF_BIN:-shelf}"
SHELF_BACKUP_DIR="${SHELF_BACKUP_DIR:-${SHELF_ROOT}/.shelf/backups}"
SHELF_BACKUP_KEEP="${SHELF_BACKUP_KEEP:-30}"

mkdir -p "${SHELF_BACKUP_DIR}"
timestamp="$(date +%Y%m%d-%H%M%S)"
outfile="${SHELF_BACKUP_DIR}/shelf-${timestamp}.json"

"${SHELF_BIN}" export --root "${SHELF_ROOT}" --out "${outfile}"
echo "Backup created: ${outfile}"

mapfile -t backups < <(ls -1t "${SHELF_BACKUP_DIR}"/shelf-*.json 2>/dev/null || true)
if [[ ${#backups[@]} -gt ${SHELF_BACKUP_KEEP} ]]; then
  for ((i=SHELF_BACKUP_KEEP; i<${#backups[@]}; i++)); do
    rm -f "${backups[$i]}"
  done
fi
echo "Retention: keep=${SHELF_BACKUP_KEEP}"
