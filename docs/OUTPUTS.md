# OUTPUTS

Machine-readable output contract for `shelf`.

This document covers the public shapes for `--json`, `--format jsonl`, and
tabular `--format tsv|csv` outputs.

Task and edge queries use one canonical contract. They are not schema-versioned.

## Task Record

Used by:

- `shelf ls`
- `shelf next`
- `shelf show --json` `task`

Fields:

- `id`
- `title`
- `path`
- `file`
- `kind`
- `status`
- `tags`
- `due_on`
- `repeat_every`
- `archived_at`
- `parent_id`
- `parent_path`
- `created_at`
- `updated_at`
- `body`

Tabular task fields:

- `id`
- `title`
- `path`
- `kind`
- `status`
- `tags`
- `due_on`
- `repeat_every`
- `archived_at`
- `parent_id`
- `parent_path`
- `file`
- `created_at`
- `updated_at`
- `body`

`shelf show --format tsv|csv` also supports:

- `outbound_count`
- `inbound_count`

## Grouped Task Record

Used by:

- `shelf ls --group-by ...`

Adds:

- `group`

JSON and JSONL keep the normal task fields and add `group`.
TSV and CSV expose the same `group` field.

## Edge Ref

Nested endpoint ref used inside edge records.

Fields:

- `id`
- `title`
- `path`
- `file`

## Edge Record

Used by:

- `shelf links --json` `edges`
- `shelf show --json` `edges`
- `shelf links --format jsonl`

Fields:

- `direction`
- `type`
- `source`
- `target`

`source` and `target` are `edgeRef` objects.

Tabular edge fields:

- `direction`
- `type`
- `source_id`
- `source_title`
- `source_path`
- `source_file`
- `target_id`
- `target_title`
- `target_path`
- `target_file`

## Link Summary Record

Used by:

- `shelf links --summary`

Fields:

- `direction`
- `type`
- `count`

## Count Output

Used by:

- `shelf ls --count`
- `shelf next --count`

Text output:

- a single integer line

JSON output:

```json
{ "count": 3 }
```

## Command Shapes

### `shelf ls --json`

- `[]taskRecord`

### `shelf ls --format jsonl`

- one `taskRecord` per line
- with `--group-by`, one grouped task record per line

### `shelf next --json`

- `[]taskRecord`

### `shelf next --format jsonl`

- one `taskRecord` per line

### `shelf show --json`

Contains:

- `task`: `taskRecord`
- `edges`: `[]edgeRecord`

### `shelf links --json`

Contains:

- `task`: inspected `edgeRef`
- `edges`: `[]edgeRecord`

### `shelf links --json --summary`

Contains:

- `task`: inspected `edgeRef`
- `summary`: `[]linkSummaryRecord`

### `shelf config copy-preset list --json`

- `[]copyPresetRecord`

### `shelf config copy-preset get --json`

- `copyPresetRecord`

`copyPresetRecord` fields:

- `name`
- `scope`
- `subtree_style`
- `template`
- `join_with`

## Stability Notes

- Machine-readable output uses one canonical contract.
- Task records use `parent_id`, not legacy parent aliases.
- Edge records use `source` and `target`, not task/other aliases.
