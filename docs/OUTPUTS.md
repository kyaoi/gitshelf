# OUTPUTS

Machine-readable output contract for `shelf`.

This document covers the current public shapes for `--json`, `--format jsonl`,
and tabular `--format tsv|csv` outputs.

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

- top-level task fields for compatibility
- `task`: normalized task record
- `edges`: normalized edge records
- `outbound`: compatibility link payloads
- `inbound`: compatibility link payloads

### `shelf links --json`

Contains:

- `task`: inspected task ref
- `edges`: normalized edge records
- `outbound`: compatibility link payloads
- `inbound`: compatibility link payloads

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

- `parent_id` and `parent` currently contain the same task ID.
- `source_*` / `target_*` are canonical edge aliases.
- `task_*` / `other_*` remain available for compatibility with older scripts.
