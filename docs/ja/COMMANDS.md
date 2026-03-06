# COMMANDS（CLI 仕様 日本語版）

このファイルは CLI の振る舞いを正確に記述する仕様書です。
「どのコマンドをどう使い分けるか」は [`COMMAND_GUIDE.md`](COMMAND_GUIDE.md) を参照してください。

## 共通

- すべてのコマンドは `--root <dir>` を受け取り、`.shelf/` を含む project root を明示できます。
- `--root` を省略した場合は、現在ディレクトリから上方向に `.shelf/` を探索します。
- ローカルに `.shelf/` が見つからない場合は、global config の `default_root` へ fallback します。
- `.shelf/` が見つからなければ非 0 終了です。
- `init` と `completion` は既存 `.shelf/` がなくても使えます。
- `--show-id`, `-i` で一覧・ツリー・対話ラベルに ID を表示します。
- task selector は本文プレビューを表示します。
- enum selector や通常の一覧コマンドは本文プレビューを表示しません。
- TTY では色付き出力が既定です。
- `NO_COLOR=1` で色を無効化できます。
- `CLICOLOR_FORCE=1` で非 TTY でも色を強制できます。

## `shelf init`

意味:
`.shelf/` の基本構造を初期化します。

作成対象:
- `.shelf/config.toml`
- `.shelf/tasks/`
- `.shelf/edges/`

フラグ:
- `--force`: `config.toml` をデフォルト値で上書き
- `--global`: global config と global shelf を初期化

## `shelf add`

意味:
通常の task 作成コマンドです。

モード:
- 非対話: `--title` 必須
- 対話: TTY で `--title` 省略時、`Title -> Kind -> Status -> Review` の順に進む

主なフラグ:
- `--title <str>`
- `--kind <kind>`
- `--status <status>`
- `--tag <tag>`（複数可）
- `--due <YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|this-week|mon..sun|next-mon..next-sun|in N days>`
- `--repeat-every <N>d|<N>w|<N>m|<N>y`
- `--parent <id|root>`
- `--body <str>`

## `shelf capture`

意味:
inbox に素早く積むためのコマンドです。

挙動:
- 常に `kind=inbox`
- 常に `status=open`
- タイトルは位置引数か `--title` で受け取る

主なフラグ:
- `--title <str>`
- `--tag <tag>`（複数可）
- `--due <date-token>`
- `--body <str>`

## `shelf triage`

意味:
`kind=inbox`, `status=open` を既定対象として処理します。

モード:
- interactive: 1件ずつ編集・status 変更・archive・skip
- auto: `--auto` で一括処理

主なフラグ:
- `--kind <kind>`（既定: `inbox`）
- `--status <status>`（既定: `open`）
- `--limit <n>`（既定: `20`）
- `--auto <done|start|block|cancel|reopen|archive>`

## `shelf template`

意味:
再利用できる task subtree を保存・適用します。

subcommand:
- `template list|ls`
- `template save <name> <id>`
- `template show <name>`
- `template apply <name>`
- `template delete|rm <name>`

## `shelf calendar`

意味:
期限をカレンダービューで確認します。

仕様:
- 既定の開始日: 今週の月曜日
- 既定の status: `open`, `in_progress`, `blocked`
- `--days <= 7` はテキスト表示
- `--days > 7` は月グリッドの TUI に切り替え
- 長期表示 TUI は TTY 必須。非TTYでは `--json` を使う

主なフラグ:
- `--start <YYYY-MM-DD|today|tomorrow>`
- `--days <n>`（既定: `7`）
- `--status <status>`（複数可）
- `--json`

## `shelf board`

意味:
status 列ベースの TUI board です。

前提:
- TTY 必須

操作:
- `h` / `l`: 列移動
- `j` / `k`: 行移動
- `o` / `s` / `b` / `d` / `c`: status 更新
- `r`: reload
- `q`: quit

## `shelf estimate`

意味:
見積もり時間・消化時間を表示または更新します。

主なフラグ:
- `--set <duration>`
- `--spent <duration>`
- `--add-spent <duration>`
- `--clear-estimate`
- `--clear-spent`
- `--json`

## `shelf track`

意味:
簡易タイマーです。

subcommand:
- `track start <id>`
- `track stop <id>`
- `track show [id]`

## `shelf notify`

意味:
期限切れ / 当日期限の active task に対してローカル shell command を実行します。

環境変数:
- `SHELF_TASK_ID`
- `SHELF_TASK_SHORT_ID`
- `SHELF_TASK_TITLE`
- `SHELF_TASK_KIND`
- `SHELF_TASK_STATUS`
- `SHELF_TASK_DUE_ON`

主なフラグ:
- `--command <shell>`
- `--dry-run`

## `shelf github`

意味:
GitHub issue / PR URL を task に紐付けます。

subcommand:
- `github link <id> --url <url>`
- `github unlink <id> --url <url>`
- `github show <id> [--json]`

対応 URL:
- `https://github.com/<owner>/<repo>/issues/<number>`
- `https://github.com/<owner>/<repo>/pull/<number>`

## `shelf sync github`

意味:
GitHub から task metadata を同期します。

挙動:
- 紐付いた最初の GitHub URL を対象に fetch
- task `title` を GitHub title で更新
- GitHub `open` -> task `open`
- GitHub `closed` -> task `done`

環境変数:
- `GITHUB_TOKEN`
- `GITSHELF_GITHUB_API_URL`

## `shelf review`

意味:
日次確認用の集約ビューです。

section:
- `Inbox`
- `Overdue`
- `Today`
- `Blocked`
- `Ready`

主なフラグ:
- `--limit <n>`
- `--json`

## `shelf ls`

意味:
最も汎用的なフラット一覧です。

主なフラグ:
- `--view <name>`
- `--preset <name>`
- `--kind <kind>`（複数可）
- `--status <status>`（複数可）
- `--tag <tag>`（複数可）
- `--not-kind <kind>`（複数可）
- `--not-status <status>`（複数可）
- `--not-tag <tag>`（複数可）
- `--ready`
- `--blocked-by-deps`
- `--due-before <YYYY-MM-DD>`
- `--due-after <YYYY-MM-DD>`
- `--overdue`
- `--no-due`
- `--parent <id|root>`
- `--limit <n>`
- `--search <query>`
- `--json`

未知の `kind` / `status` / `tag` はエラーです。

## `shelf next`

意味:
着手可能な task だけを表示します。

主なフラグ:
- `--view <name>`
- `--preset <name>`
- `--limit <n>`
- `--json`

## `shelf view`

意味:
保存済み filter を管理します。

subcommand:
- `view list`
- `view show <name>`
- `view set <name> ...`
- `view copy <from> <to>`
- `view rename <from> <to>`
- `view merge <name> --from <view> --strategy union|intersection`
- `view delete <name>`

## `shelf preset`

意味:
コマンド別 output preset を管理します。

subcommand:
- `preset list`
- `preset show <name>`
- `preset set <name> --command <command> ...`
- `preset delete <name>`

## `shelf agenda`

意味:
期限ベースで bucket 分けされた一覧です。

主なフラグ:
- `--preset <name>`
- `--view <name>`
- `--days <n>`
- `--kind <kind>`
- `--status <status>`
- `--not-kind <kind>`
- `--not-status <status>`
- `--json`

## `shelf today`

意味:
overdue と today に集中した一覧です。

主なフラグ:
- `--preset <name>`
- `--view <name>`
- `--carry-over`
- `--yes`
- `--kind <kind>`
- `--status <status>`
- `--not-kind <kind>`
- `--not-status <status>`
- `--json`

## `shelf tree`

意味:
親子関係による階層表示です。

主なフラグ:
- `--preset <name>`
- `--view <name>`
- `--from <id|root>`
- `--max-depth <n>`
- `--kind`, `--status`, `--tag`
- `--not-kind`, `--not-status`, `--not-tag`
- `--json`

## `shelf show <id>`

意味:
1件の task 詳細を表示します。

主なフラグ:
- `--no-body`
- `--only-body`
- `--json`

## `shelf explain <id>`

意味:
1件の task の readiness や filter/view 一致理由を説明します。

主なフラグ:
- `--view <name>`
- `--json`

## `shelf edit [id]`

意味:
task ファイル全体を editor で開きます。

editor の解決順:
- `$VISUAL`
- `$EDITOR`
- `vi`

## `shelf set <id>`

意味:
構造化更新を行います。

主なフラグ:
- `--title <str>`
- `--kind <kind>`
- `--status <status>`
- `--tag <tag>`
- `--untag <tag>`
- `--clear-tags`
- `--due <date-token>`
- `--clear-due`
- `--repeat-every <repeat>`
- `--clear-repeat`
- `--parent <id|root>`
- `--body <str>`
- `--append-body <str>`

## `shelf snooze <id>`

意味:
期限を相対または絶対日付で変更します。

挙動:
- `--by` のみ指定: 現在の `due_on` を相対日数でずらす
- `--to` のみ指定: 新しい `due_on` を直接設定する
- どちらも未指定:
  - TTY: task 選択後に `by` / `to` を対話で選ぶ
  - 非TTY: 明示的なエラー

指定:
- `--by <Nd>`
- `--to <YYYY-MM-DD|today|tomorrow>`

## `shelf archive` / `shelf unarchive`

意味:
archived 状態へ退避 / 復帰します。

## `shelf mv <id>`

意味:
task の親を付け替えます。

主なフラグ:
- `--parent <id|root>`

## status ショートカット

- `shelf done <id>`
- `shelf start <id>`
- `shelf block <id>`
- `shelf cancel <id>`
- `shelf reopen <id>`

意味:
`status` 変更の短縮コマンドです。

## `shelf link`

意味:
outbound link を作成します。

指定:
- `--from <id>`
- `--to <id>`
- `--type <depends_on|related>`

## `shelf unlink`

意味:
outbound link を削除します。

指定:
- `--from <id>`
- `--to <id>`
- `--type <depends_on|related>`

## `shelf links <id>`

意味:
inbound / outbound link を確認します。

主なフラグ:
- `--transitive`
- `--suggest`
- `--limit <n>`
- `--json`

## `shelf deps <id>`

意味:
`depends_on` の前提・逆依存を確認します。

主なフラグ:
- `--transitive`
- `--reverse`
- `--graph`
- `--suggest`
- `--limit <n>`
- `--json`

## `shelf export`

意味:
`.shelf` 全体を JSON へ書き出します。

主なフラグ:
- `--out <path>`（`-` は stdout）

## `shelf import`

意味:
JSON export を読み込みます。

主なフラグ:
- `--in <path>`
- `--validate-only`
- `--dry-run`
- `--merge`
- `--replace`

## `shelf undo`

意味:
最後の snapshot へ巻き戻します。

## `shelf redo`

意味:
undo した変更を再適用します。

## `shelf history`

意味:
更新履歴と snapshot を確認します。

subcommand:
- `history`
- `history show <entry|snapshot_id>`

## `shelf doctor`

意味:
`.shelf` の整合性チェックを行います。

主なフラグ:
- `--fix`
- `--strict`
- `--json`

## `shelf completion`

意味:
shell completion script を出力します。

対象:
- `bash`
- `zsh`
- `fish`
- `powershell`
