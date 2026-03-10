# OUTPUTS

Machine-readable output contract for `shelf`.

This document covers the public shapes for `--json`, `--format jsonl`, and
tabular `--format tsv|csv` outputs.

Task and edge queries support `--schema <v1|v2>`.

- `v1` is the default compatibility schema.
- `v2` is the canonical opt-in schema.

## Task Record

Used by:

- `shelf ls`
- `shelf next`
- `shelf show --json` `task`

### v1

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
- `parent`
- `parent_title`
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
- `parent`
- `parent_path`
- `file`
- `created_at`
- `updated_at`
- `body`

### v2

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

JSON and JSONL keep the normal task fields for the selected schema and add `group`.
TSV and CSV expose the same `group` field.

## Edge Record

Used by:

- `shelf links --json` `edges`
- `shelf show --json` `edges`
- `shelf links --format jsonl`

### v1

Fields:

- `direction`
- `type`
- `source`
- `target`
- `task`
- `other`

Nested refs contain:

- `id`
- `title`
- `path`
- `file`

`source` and `target` are the canonical edge endpoints.
`task` and `other` are compatibility aliases based on the inspected task context.

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
- `task_id`
- `task_title`
- `task_path`
- `task_file`
- `other_id`
- `other_title`
- `other_path`
- `other_file`

### v2

Fields:

- `direction`
- `type`
- `source`
- `target`

Nested refs contain:

- `id`
- `title`
- `path`
- `file`

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

### `shelf show --json --schema v1`

Contains:

- top-level task fields for compatibility
- `task`: normalized task record
- `edges`: normalized edge records
- `outbound`: compatibility link payloads
- `inbound`: compatibility link payloads

### `shelf show --json --schema v2`

Contains:

- `task`: canonical task record
- `edges`: canonical edge records

### `shelf links --json --schema v1`

Contains:

- `task`: inspected task ref
- `edges`: normalized edge records
- `outbound`: compatibility link payloads
- `inbound`: compatibility link payloads

### `shelf links --json --schema v2`

Contains:

- `task`: inspected task ref
- `edges`: canonical edge records

### `shelf links --json --summary`

Contains:

- `task`
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

## Compatibility Notes

- `v1` remains the default schema.
- `v2` removes task field aliases such as `parent`.
- `v2` removes edge field aliases such as `task`, `other`, `task_*`, and `other_*`.
