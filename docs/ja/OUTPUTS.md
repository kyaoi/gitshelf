# OUTPUTS

`shelf` の machine-readable output contract です。

このドキュメントは `--json`, `--format jsonl`, `--format tsv|csv` の
public shape を対象にします。

task / edge query は 1 つの canonical contract を使います。
schema version 分岐はありません。

## Task Record

使われる箇所:

- `shelf ls`
- `shelf next`
- `shelf show --json` の `task`

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

tabular task field:

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

使われる箇所:

- `shelf ls --group-by ...`

追加 field:

- `group`

JSON と JSONL は通常の task field に `group` を追加します。
TSV と CSV でも同じ `group` field を使います。

## Edge Ref

edge record に入る endpoint ref です。

field:

- `id`
- `title`
- `path`
- `file`

## Edge Record

使われる箇所:

- `shelf links --json` の `edges`
- `shelf show --json` の `edges`
- `shelf links --format jsonl`

field:

- `direction`
- `type`
- `source`
- `target`

`source` と `target` は `edgeRef` object です。

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

使われる箇所:

- `shelf links --summary`

field:

- `direction`
- `type`
- `count`

## Count Output

使われる箇所:

- `shelf ls --count`
- `shelf next --count`

text 出力:

- 整数 1 行

JSON 出力:

```json
{ "count": 3 }
```

## Command Shapes

### `shelf ls --json`

- `[]taskRecord`

### `shelf ls --format jsonl`

- 1 行 1 `taskRecord`
- `--group-by` 付きでは 1 行 1 grouped task record

### `shelf next --json`

- `[]taskRecord`

### `shelf next --format jsonl`

- 1 行 1 `taskRecord`

### `shelf show --json`

内容:

- `task`: `taskRecord`
- `edges`: `[]edgeRecord`

### `shelf links --json`

内容:

- `task`: 対象 task の `edgeRef`
- `edges`: `[]edgeRecord`

### `shelf links --json --summary`

内容:

- `task`: 対象 task の `edgeRef`
- `summary`: `[]linkSummaryRecord`

### `shelf config copy-preset list --json`

- `[]copyPresetRecord`

### `shelf config copy-preset get --json`

- `copyPresetRecord`

`copyPresetRecord` の field:

- `name`
- `scope`
- `subtree_style`
- `template`
- `join_with`

## Stability Notes

- machine-readable output は 1 つの canonical contract を使います。
- task record は legacy parent alias ではなく `parent_id` を使います。
- edge record は task/other alias ではなく `source` / `target` を使います。
