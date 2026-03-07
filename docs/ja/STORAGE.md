# STORAGE（保存形式と不変条件 日本語版）

## ディレクトリ構造

```text
.shelf/
  config.toml
  .write.lock
  tasks/
    <id>.md
  edges/
    <src_id>.toml
  templates/
    <name>.json
  history/
    index.json
    actions.log
    snapshots/
```

## Config File（`.shelf/config.toml`）

主要キー:

- `kinds`
- `statuses`
- `tags`
- `link_types`
- `default_kind`
- `default_status`
- `[commands.calendar]`
  - `default_range_unit`
  - `default_days`
  - `default_months`
  - `default_years`

保存済み view:

```toml
[views."active"]
tags = ["backend"]
not_statuses = ["done", "cancelled"]
```

保存済み output preset:

```toml
[output_presets."ls_focus"]
command = "ls"
view = "active"
format = "detail"
limit = 20
```

## ID

- 新規 task では ULID を使う
- 表示用 short ID は先頭 8 文字

## Task File（`.shelf/tasks/<id>.md`）

- Markdown body + TOML front matter（`+++ ... +++`）
- front matter は構造化 metadata
- body は自由記述ノート

主要キー:

- 必須:
  - `id`
  - `title`
  - `kind`
  - `status`
  - `created_at`
  - `updated_at`
- 任意:
  - `tags`
  - `github_urls`
  - `estimate_minutes`
  - `spent_minutes`
  - `timer_started_at`
  - `due_on`
  - `repeat_every`
  - `archived_at`
  - `parent`
  - body text

`due_on` 入力では以下を受け付け、保存時には `YYYY-MM-DD` へ正規化します。

- `today`
- `tomorrow`
- `+Nd`
- `-Nd`
- `next-week`
- `this-week`
- `mon..sun`
- `next-mon..next-sun`
- `in N days`

body の典型用途:

- 詳細説明
- 補足
- 実行メモ
- 進捗ログ
- アイデア下書き
- 参考情報

### front matter の安定順序

1. `id`
2. `title`
3. `kind`
4. `status`
5. `tags`
6. `github_urls`
7. `estimate_minutes`
8. `spent_minutes`
9. `timer_started_at`
10. `due_on`
11. `repeat_every`
12. `archived_at`
13. `parent`
14. `created_at`
15. `updated_at`

時刻は RFC3339 を使います。

## Edge File（`.shelf/edges/<src_id>.toml`）

`[[edge]]` 配列で次を持ちます。

- `to`
- `type`

出力順は安定化されます。

1. `to` 昇順
2. `type` 昇順

重複 `(to, type)` は write 時に除去します。

## Template File（`.shelf/templates/<name>.json`）

- subtree の再利用スナップショット
- template 内では local key と `parent_key` を持つ
- `template apply` で parent 参照を新しい task ID に書き換える

## 不変条件

### Task

- ファイル名の ID と front matter `id` は一致する
- `title` は空不可
- `kind` は config `kinds` に存在する
- `status` は config `statuses` に存在する
- `commands.calendar.default_range_unit` は `days` / `months` / `years`
- `commands.calendar.default_days` は 1 以上
- `commands.calendar.default_months` は 1 以上
- `commands.calendar.default_years` は 1 以上
- 各 `tag` は config `tags` に存在する
- 各 `github_urls` は canonical な GitHub issue / PR URL
- `due_on` は存在時 `YYYY-MM-DD`
- `estimate_minutes` / `spent_minutes` は `>= 0`
- `timer_started_at` は存在時 RFC3339
- `repeat_every` は存在時 `<N>d|<N>w|<N>m|<N>y`
- `archived_at` は存在時 RFC3339
- `parent` がある場合:
  - 親 task が存在する
  - 自分自身を指さない
  - 親循環を作らない

### Edge

- source task が存在すること
- `type` は config `link_types` に存在すること
- destination task が存在すること
- `(type, to)` の重複は無効

`add` / `set` で新規 tag を入れた場合、config `tags` へ自動追記します。

## Link Direction

`A depends_on B` は常に以下の意味です。

`A --depends_on--> B`

## Atomic Writes

すべての write は temp file -> rename で行い、途中破損を避けます。

## Undo Snapshot

mutating command は `.shelf/history/snapshots` に snapshot を作ります。

- `index.json`: undo / redo stack
- `actions.log`: action record

## Write Lock

mutating command 実行中は `.shelf/.write.lock` を使って排他します。

## Error Messages

可能な限り file path と原因を含めます。

例:

`.shelf/tasks/<id>.md: unknown kind: ...`
