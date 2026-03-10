# STORAGE（日本語版）

Cockpit-first な現行ツールセットの保存形式です。

## ディレクトリ構成

既定レイアウト:

```text
.shelf/
  config.toml
  tasks/
    <id>.md
  edges/
    <src_id>.toml
```

config は常に `.shelf/config.toml` に置かれます。
task データは `storage_root` 配下に置かれ、既定値は `.shelf` です。

root 直下に置く例:

```toml
storage_root = "."
```

```text
.shelf/
  config.toml
tasks/
  <id>.md
edges/
  <src_id>.toml
```

## Config

現在の config に保存するもの:

- `kinds`
- `statuses`
- `tags`
- `storage_root`
- `link_types`
- `default_kind`
- `default_status`
- `[commands.calendar]`
- `[commands.cockpit]`

link type 設定:

```toml
[link_types]
names = ["depends_on", "related"]
blocking = "depends_on"
```

- `names` に許可する link type 名を列挙する
- `blocking` は readiness / cycle check に使う関係名
- 既定の blocking relation は、子 task から祖先 task への link も禁止する

calendar 設定:

```toml
[commands.calendar]
default_range_unit = "days"
default_days = 7
default_months = 6
default_years = 2
```

cockpit 設定:

```toml
[commands.cockpit]
copy_separator = "\n"
post_exit_git_action = "none"
commit_message = "chore: update shelf data"

[[commands.cockpit.copy_presets]]
name = "subtree_path"
scope = "subtree"
subtree_style = "indented"
template = "{{path}}\n{{subtree}}"
join_with = "\n\n"
```

- `copy_separator`
- `copy_presets[].name`
- `copy_presets[].scope`: `task` または `subtree`
- `copy_presets[].subtree_style`: `indented` または `tree`
- `copy_presets[].template`: `{{title}}`, `{{path}}`, `{{body}}`, `{{subtree}}` を使えます
- `copy_presets[].join_with`: 省略時は `copy_separator` を使います

保存先設定:

```toml
storage_root = ".shelf"
```

- `storage_root` は `tasks/` と `edges/` の共通親ディレクトリです
- 相対パスは project root 起点で解決します
- `.` を指定すると `tasks/` と `edges/` を root 直下に置きます

## Task File

各 task は `<storage_root>/tasks/<id>.md` に保存されます。

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

## Edge File

各 task は outbound edge file を1つ持てます。

- `<storage_root>/edges/<src_id>.toml`

形式:

```toml
[[edge]]
to = "01..."
type = "depends_on"
```

対応 link type は `config.toml` の `link_types.names` で決まります。
既定値は `depends_on`, `related` です。

## 不変条件

- 1 task = 1 markdown file
- task file は full ID で識別する
- edge file は outbound link のみ持つ
- 親子構造は `parent` で表現する
- link graph は edge file で表現する
