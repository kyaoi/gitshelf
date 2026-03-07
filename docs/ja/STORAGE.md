# STORAGE（日本語版）

Cockpit-first な現行ツールセットの保存形式です。

## ディレクトリ構成

```text
.shelf/
  config.toml
  tasks/
    <id>.md
  edges/
    <src_id>.toml
```

`.shelf/templates/` や `.shelf/history/` は現在の active layout には含みません。存在する場合は `shelf init` で削除されます。

## Config

現在の config に保存するもの:

- `kinds`
- `statuses`
- `tags`
- `link_types`
- `default_kind`
- `default_status`
- `[commands.calendar]`

calendar 設定:

```toml
[commands.calendar]
default_range_unit = "days"
default_days = 7
default_months = 6
default_years = 2
```

`[views.*]` や `[output_presets.*]` のような legacy section は migration 用に読める場合がありますが、現在は書き戻しません。

## Task File

各 task は `.shelf/tasks/<id>.md` に保存されます。

現在の front matter:

- `id`
- `title`
- `kind`
- `status`
- `tags`
- `due_on`
- `repeat_every`
- `archived_at`
- `parent`
- `created_at`
- `updated_at`

本文は自由記述ノートです。

補足:

- `github_urls`, `estimate_minutes`, `spent_minutes`, `timer_started_at` のような legacy field は読めることがありますが、現在の書き込みでは出力しません

## Edge File

各 task は outbound edge file を1つ持てます。

- `.shelf/edges/<src_id>.toml`

形式:

```toml
[[edge]]
to = "01..."
type = "depends_on"
```

対応 link type:

- `depends_on`
- `related`

## 不変条件

- 1 task = 1 markdown file
- task file は full ID で識別する
- edge file は outbound link のみ持つ
- 親子構造は `parent` で表現する
- link graph は edge file で表現する
- 現在の書き込みは active schema に正規化し、legacy field/section は再出力しない
