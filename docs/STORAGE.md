# STORAGE（保存形式と不変条件）

## ディレクトリ
```text
.shelf/
  config.toml
  tasks/
    <id>.md
  edges/
    <src_id>.toml
```

## ID
- ULID推奨（時系列ソートに強い）
- 表示用短縮ID: `id[:8]`（または `id[:10]`）

## タスクファイル（Markdown + TOML front matter）
- 先頭を `+++` で囲むTOML front matter
- 本文は任意（空でも可）

### 例
```md
+++
id = "01JABCDEF0123456789XYZ"
title = "月曜日にやること"
kind = "todo"
state = "open"
parent = "01JWEEKGOAL000000000000" # rootなら省略/空
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

（任意のメモ）
```

### 必須/任意
- 必須: `id`, `title`, `kind`, `state`, `created_at`, `updated_at`
- 任意: `parent`, 本文

## edges（outbound links）
- ファイル: `.shelf/edges/<src_id>.toml`
- `[[edge]]` の配列

### 例
```toml
[[edge]]
to = "01JNOTE...."
type = "depends_on"

[[edge]]
to = "01JQUANT...."
type = "related"
```

## 不変条件（MUST）
### タスク
- `id` はファイル名の `<id>` と一致
- `title` は空でない
- `kind` は `config.toml` の `kinds` に含まれる（含まれない場合はエラー）
- `state` は `config.toml` の `states` に含まれる（含まれない場合はエラー）
- `parent` がある場合:
  - 対象タスクが存在する
  - 自分自身ではない
  - 循環が発生しない（親を辿って自分に戻らない）

### edges
- `type` は `config.toml` の `link_types` に含まれる（含まれない場合はエラー）
- `to` のタスクが存在する（存在しない場合はエラー）
- 同じ `(type,to)` の重複は禁止（idempotent）
- `depends_on` の意味は固定（SPEC参照）

## 安定ソート（SHOULD）
- `ls` のデフォルト: `created_at desc`（新しい順）
- `tree`: siblingは `created_at asc` か `title asc` のいずれかで固定（どちらでも良いが固定する）
- `edges` ファイル内の `[[edge]]`: `to asc`, `type asc` で固定

## 原子的更新（MUST）
- 変更は `*.tmp` に書き出して `rename`（同一FS上）
- 失敗時は既存ファイルを壊さない

## エラーメッセージ（MUST）
- パスと原因を含める（例: `.shelf/tasks/<id>.md: invalid kind: ...`）
