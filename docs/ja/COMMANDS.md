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

## `shelf`

- TTY: `Cockpit` を `calendar` mode で開く
- 非TTY: help を表示

## `shelf init`

現在の shelf を初期化または整理します。

作成・維持するもの:

- `.shelf/config.toml`
- `.shelf/tasks/`
- `.shelf/edges/`

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
- `--format compact|detail|kanban`
- `--json`

未知の kind/status/tag は即エラーです。

## `shelf next`

着手可能 task の shortlist を返す read-only command です。

フラグ:

- `--limit <n>`
- `--json`

## `shelf link`

outbound link を作成します。

フラグ:

- `--from <id>`
- `--to <id>`
- `--type <depends_on|related>`

## `shelf unlink`

outbound link を削除します。

フラグ:

- `--from <id>`
- `--to <id>`
- `--type <depends_on|related>`

## `shelf links`

1つの task の outbound / inbound link を表示します。

使い方:

- `shelf links <task-id>`

フラグ:

- `--json`

## 補足

現在の公開 CLI では、add/edit/show/set/mv/snooze/archive/history/import/export/github/view/doctor などの standalone command は公開していません。

日常編集の中心は Cockpit のままですが、link 管理は standalone command でも行えます。
