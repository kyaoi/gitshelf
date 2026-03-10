# COMMANDS（日本語版）

現在の公開 CLI surface の仕様です。

## 共通

- `--root <dir>` で `.shelf/` を含む root を明示できます
- `--root` と global `default_root` では `~/DailyTodo` のような `~` 略記を使えます
- 指定パスが `<root>/.shelf` を指す場合は自動で `<root>` に正規化します
- `.shelf/tasks` のような `.shelf/` 配下のパスは root として拒否します
- `--root` 省略時は上方向探索します
- ローカル `.shelf/` が無ければ global `default_root` に fallback します
- `init` と `completion` は既存 `.shelf/` 不要です
- `--show-id`, `-i` で text 出力や selector に ID を表示します
- `--git-on-exit <none|commit|commit_push>` で Cockpit 終了後の git 動作を上書きできます
- `--git-message <text>` で `--git-on-exit` が使う commit message を上書きできます
- task data の保存先は `.shelf/config.toml` の `storage_root` で制御します

## `shelf`

- TTY: `Cockpit` を `calendar` mode で開く
- 非TTY: help を表示

## `shelf init`

現在の shelf を初期化または整理します。

作成・維持するもの:

- `.shelf/config.toml`
- `<storage_root>/tasks/`
- `<storage_root>/edges/`

フラグ:

- `--force`: `config.toml` をデフォルトで再生成
- `--global`: global default root と global config を初期化

## `shelf completion`

shell completion を生成します。

subcommand:

- `completion bash`
- `completion zsh`
- `completion fish`
- `completion powershell`

## `shelf cockpit`

主入口の TUI workspace です。

alias:

- `cp`

TTY 必須。

フラグ:

- `--mode <calendar|tree|board|review|now>`
- `--start <YYYY-MM-DD|today|tomorrow>`
- `--days <n>`
- `--months <n>`
- `--years <n>`
- `--limit <n>`
- `--kind <kind>`（複数可）
- `--status <status>`（複数可）
- `--tag <tag>`（複数可）
- `--not-kind <kind>`（複数可）
- `--not-status <status>`（複数可）
- `--not-tag <tag>`（複数可）

## `shelf config`

user-facing config を永続化します。

### `shelf config show`

現在の root で有効な config を表示します。

フラグ:

- `--json`
- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` (`--format tsv|csv` 専用)
- `--header`
- `--no-header`

### `shelf config copy-preset list`

保存済み advanced copy preset の一覧を表示します。

フラグ:

- `--json`
- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` (`--format tsv|csv` 専用)
- `--header`
- `--no-header`

### `shelf config copy-preset get`

保存済み advanced copy preset を1件表示します。

使い方:

- `shelf config copy-preset get <name>`

フラグ:

- `--json`

### `shelf config copy-preset rm`

保存済み advanced copy preset を1件削除します。

使い方:

- `shelf config copy-preset rm <name>`

### `shelf config copy-preset set`

Cockpit の `M` で使う advanced copy preset を追加または更新します。

フラグ:

- `--name <preset-name>`
- `--scope <task|subtree>`
- `--subtree-style <indented|tree>` 省略可。`{{subtree}}` の描画方式
- `--template <text>`
- `--join-with <text>` 省略時は `commands.cockpit.copy_separator`

使える placeholder:

- `{{title}}`
- `{{path}}`
- `{{body}}`
- `{{subtree}}`

## launcher 群

次のコマンドは `shelf cockpit --mode ...` の薄い wrapper です。

### `shelf calendar`

alias:

- `cal`

### `shelf tree`

alias:

- `tr`

### `shelf board`

alias:

- `kb`

### `shelf review`

alias:

- `rv`

### `shelf now`

alias:

- `nw`

launcher で使えるフラグは共通です。

- `--start`
- `--days`
- `--months`
- `--years`
- `--limit`
- `--kind`
- `--status`
- `--tag`
- `--not-kind`
- `--not-status`
- `--not-tag`

すべて TTY 必須です。

## `shelf ls`

script と単発確認向けの read-only 一覧です。

フラグ:

- `--kind <kind>`（複数可）
- `--status <status>`（複数可）
- `--tag <tag>`（複数可）
- `--not-kind <kind>`（複数可）
- `--not-status <status>`（複数可）
- `--not-tag <tag>`（複数可）
- `--ready`
- `--blocked-by-deps`
- `--due-before <date>`
- `--due-after <date>`
- `--overdue`
- `--no-due`
- `--parent <id|root>`
- `--search <text>`
- `--limit <n>`
- `--include-archived`
- `--only-archived`
- `--format compact|detail|kanban|tree|tsv|csv|jsonl`
- `--preset <now|review|board>`
- `--fields <name,...>` (`--format tsv|csv` 専用)
- `--header`
- `--no-header`
- `--sort <id|title|path|kind|status|due_on|created_at|updated_at>`
- `--reverse`
- `--group-by <status|kind|parent>`
- `--count`
- `--json`

未知の kind/status/tag は即エラーです。

`--preset` は対応する Cockpit view に近い read-only default を適用します。
明示した flag がある場合はそちらが優先されます。

`--format tsv` と `--format csv` は同じ task field を使います。
TSV は既定で header なし、CSV は既定で header ありです。
`--header` と `--no-header` で上書きできます。

task field:

- `ls`: `id`, `title`, `path`, `kind`, `status`, `due_on`, `repeat_every`, `archived_at`, `parent`, `parent_path`, `tags`, `file`
- `next`: `id`, `title`, `path`, `kind`, `status`, `due_on`, `repeat_every`, `parent`, `parent_path`, `tags`, `file`

`--fields` で列の順序変更や絞り込みができます。
`--format jsonl` は `--json` と同じ task object を1行1件で出力します。
`--group-by` を付けると tabular/JSON 出力には `group` field が追加され、text 出力は group ごとの見出し付きになります。
`--group-by` は `--format kanban`, `--format tree`, `--count` と併用できません。
`--count` は条件に一致する task 総数だけを返します。`--json` を付けると `{ "count": N }` です。

## `shelf next`

着手可能 task の shortlist を返す read-only command です。

フラグ:

- `--limit <n>`
- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` (`--format tsv|csv` 専用)
- `--header`
- `--no-header`
- `--sort <id|title|path|kind|status|due_on|created_at|updated_at>`
- `--reverse`
- `--count`
- `--json`

`--count` は ready task の総数だけを返します。`--json` を付けると `{ "count": N }` です。

## `shelf show`

1つの task を inspector 風の詳細表示で返します。

使い方:

- `shelf show <task-id>`

フラグ:

- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` (`--format tsv|csv` 専用)
- `--header`
- `--no-header`
- `--json`

`--format tsv` と `--format csv` は選択 task を1行で出力します。
使える field:

- `id`, `title`, `path`, `kind`, `status`, `tags`, `due_on`, `repeat_every`, `archived_at`
- `parent`, `parent_path`, `file`, `created_at`, `updated_at`, `body`
- `outbound_count`, `inbound_count`

`--format jsonl` は task object を1行で出力します。

## `shelf link`

outbound link を作成します。

フラグ:

- `--from <id>`
- `--to <id>`
- `--type <link-type>`

`--type` を省略した場合は config の blocking relation を使います。

## `shelf unlink`

outbound link を削除します。

フラグ:

- `--from <id>`
- `--to <id>`
- `--type <link-type>`

`--type` を省略した場合は config の blocking relation を使います。

## `shelf links`

1つの task の outbound / inbound link を表示します。

使い方:

- `shelf links <task-id>`

フラグ:

- `--format compact|tsv|csv|jsonl`
- `--fields <name,...>` (`--format tsv|csv` 専用)
- `--header`
- `--no-header`
- `--summary`
- `--json`

JSON 出力には正規化された `task` と `edges` が含まれ、互換維持のために `outbound` / `inbound` の path/file 情報も残しています。
`--format tsv` と `--format csv` は edge ごとに1行です。
使える field:

- `direction`, `type`
- `task_id`, `task_title`, `task_path`, `task_file`
- `other_id`, `other_title`, `other_path`, `other_file`

`--format jsonl` は edge object を1行1件で出力します。
`--summary` を付けると `direction/type/count` の集計行に切り替わります。

text 出力は tree/path ラベルを使い、同名 task を見分けやすくしています。
ID は `--show-id` を付けたときだけ表示します。

## 補足

日常編集の中心は Cockpit です。
link 管理は standalone command でも行えます。
