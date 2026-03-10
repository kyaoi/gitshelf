# OUTPUTS（日本語版）

`shelf` の machine-readable output contract です。

対象:

- `--json`
- `--format jsonl`
- `--format tsv|csv`

task / edge query は `--schema <v1|v2>` をサポートします。

- `v1` は既定の互換 schema です。
- `v2` は opt-in の canonical schema です。

## Task Record

使われる場所:

- `shelf ls`
- `shelf next`
- `shelf show --json` の `task`

### v1

field:

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

tabular field:

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

field:

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

tabular field:

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

`shelf show --format tsv|csv` では次も使えます。

- `outbound_count`
- `inbound_count`

## Grouped Task Record

使われる場所:

- `shelf ls --group-by ...`

追加 field:

- `group`

JSON/JSONL では選択した schema の task field に `group` が追加されます。
TSV/CSV でも同じ `group` field を使います。

## Edge Record

使われる場所:

- `shelf links --json` の `edges`
- `shelf show --json` の `edges`
- `shelf links --format jsonl`

### v1

field:

- `direction`
- `type`
- `source`
- `target`
- `task`
- `other`

ネストされた ref の field:

- `id`
- `title`
- `path`
- `file`

`source` と `target` が canonical な edge endpoint です。
`task` と `other` は inspected task 基準の互換 alias です。

tabular edge field:

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

field:

- `direction`
- `type`
- `source`
- `target`

ネストされた ref の field:

- `id`
- `title`
- `path`
- `file`

tabular edge field:

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

使われる場所:

- `shelf links --summary`

field:

- `direction`
- `type`
- `count`

## Count Output

使われる場所:

- `shelf ls --count`
- `shelf next --count`

text 出力:

- 数値1行

JSON 出力:

```json
{ "count": 3 }
```

## Command Shapes

### `shelf ls --json`

- `[]taskRecord`

### `shelf ls --format jsonl`

- 1行1 `taskRecord`
- `--group-by` 付きでは grouped task record を1行ずつ出力

### `shelf next --json`

- `[]taskRecord`

### `shelf next --format jsonl`

- 1行1 `taskRecord`

### `shelf show --json --schema v1`

内容:

- 互換維持のための top-level task field
- `task`: 正規化された task record
- `edges`: 正規化された edge record
- `outbound`: 互換維持の link payload
- `inbound`: 互換維持の link payload

### `shelf show --json --schema v2`

内容:

- `task`: canonical task record
- `edges`: canonical edge record

### `shelf links --json --schema v1`

内容:

- `task`: inspected task ref
- `edges`: 正規化された edge record
- `outbound`: 互換維持の link payload
- `inbound`: 互換維持の link payload

### `shelf links --json --schema v2`

内容:

- `task`: inspected task ref
- `edges`: canonical edge record

### `shelf links --json --summary`

内容:

- `task`
- `summary`: `[]linkSummaryRecord`

### `shelf config copy-preset list --json`

- `[]copyPresetRecord`

### `shelf config copy-preset get --json`

- `copyPresetRecord`

`copyPresetRecord` field:

- `name`
- `scope`
- `subtree_style`
- `template`
- `join_with`

## Compatibility Notes

- `v1` は既定 schema のままです。
- `v2` では `parent` のような task alias を出しません。
- `v2` では `task`, `other`, `task_*`, `other_*` のような edge alias を出しません。
