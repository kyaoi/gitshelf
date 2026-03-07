# gitshelf 日本語ガイド

`gitshelf` は、Git リポジトリにそのまま置ける軽量 CLI タスク管理ツールです。

- コマンド名: `shelf`
- 保存先: `.shelf/`
- タスク: `.shelf/tasks/<id>.md`
- リンク: `.shelf/edges/<src_id>.toml`
- 親子構造: task の `parent`
- 追加関係: `depends_on`, `related`

## 日本語ドキュメント一覧

- CLI 仕様: [`COMMANDS.md`](COMMANDS.md)
- 詳細コマンドガイド: [`COMMAND_GUIDE.md`](COMMAND_GUIDE.md)
- 保存形式と不変条件: [`STORAGE.md`](STORAGE.md)
- 対話 UI 仕様: [`INTERACTIVE.md`](INTERACTIVE.md)
- 全体仕様: [`SPEC.md`](SPEC.md)
- 運用例: [`EXAMPLES.md`](EXAMPLES.md)
- フォーマット規約: [`FORMATTING.md`](FORMATTING.md)
- 品質確認の観点: [`TESTING.md`](TESTING.md)
- 仕様と実装タスクの対応: [`TRACEABILITY.md`](TRACEABILITY.md)
- 主要な設計判断: [`DECISIONS.md`](DECISIONS.md)
- 実装計画の履歴: [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md)

補足:
- `docs/default_config.toml` は言語非依存のサンプル設定としてそのまま参照します。

## インストール

```bash
go build -o shelf ./cmd/shelf
```

## クイックスタート

```bash
# 現在のディレクトリで初期化
./shelf init

# すぐにメモを積む
./shelf capture "Call vendor"

# 構造化された task を追加
./shelf add --title "Weekly Goal"
./shelf add --title "Monday Plan" --parent root

# 一覧・ツリー・詳細
./shelf ls
./shelf tree
./shelf show <task-id>

# 更新
./shelf set <task-id> --status done
./shelf mv <task-id> --parent root

# 日次レビュー
./shelf triage
./shelf cockpit
./shelf review
./shelf next
./shelf now

# 関係づけ
./shelf link --from <a> --to <b> --type depends_on
./shelf deps <a> --graph --transitive

# GitHub 連携
./shelf github link <task-id> --url https://github.com/acme/repo/issues/42
./shelf sync github <task-id>

# 整合性チェック
./shelf doctor
```

## 基本概念

### kind

タスクの種類です。例:

- `todo`
- `idea`
- `memo`
- `inbox`

### status

進捗状態です。デフォルトでは次を使います。

- `open`
- `in_progress`
- `blocked`
- `done`
- `cancelled`

### 本文

task ファイルの front matter より下は自由記述のノート欄です。用途は以下です。

- 詳細説明
- 補足
- 実行メモ
- 進捗ログ
- アイデア本文
- 参考情報

### link type

- `depends_on`: `A depends_on B` は「A をやるには B が先」
- `related`: 雑な関連づけ

## どのコマンドを使うべきか

- すぐ積む: `capture`
- きちんと作る: `add`
- inbox を捌く: `triage`
- 今日の全体像を見る: `review`
- 今やるものだけ見る: `next`
- 期限ベースで見る: `now`, `agenda`, `calendar`, `cockpit`
- 階層で見る: `tree`
- 1件を深掘りする: `show`
- 生ファイルを触る: `edit`
- メタデータ更新: `set`
- 親子を変える: `mv`
- 依存や関連を貼る: `link`, `links`, `deps`
- GitHub issue / PR を紐付ける: `github`, `sync github`
- 工数を見る: `estimate`, `track`
- 壊れを確認する: `doctor`

詳しい判断基準は [`COMMAND_GUIDE.md`](COMMAND_GUIDE.md) を参照してください。

## 対話 UI

TTY では、引数を省略したときに対話選択へ入るコマンドがあります。

- `add`
- `show`
- `edit`
- `set`
- `mv`
- `done`
- `links`
- `triage`

基本操作:

- `j` / `k`: 移動
- `Enter`: 決定
- `/`: 検索
- `q`, `Esc`, `Ctrl+C`: キャンセル

詳細は [`INTERACTIVE.md`](INTERACTIVE.md) を参照してください。

## 保存形式

- `.shelf/config.toml`: kinds / statuses / tags / views / presets
- `.shelf/tasks/<id>.md`: task 本体
- `.shelf/edges/<src_id>.toml`: outbound links
- `.shelf/templates/<name>.json`: 再利用テンプレート
- `.shelf/history/`: undo / redo 用 snapshot

詳細は [`STORAGE.md`](STORAGE.md) を参照してください。

## Global Shelf と Fallback

gitshelf は global default root をサポートします。

- config path: `~/.config/gitshelf/config.toml`（Linux 既定、内部的には `os.UserConfigDir()`）

形式:

```toml
default_root = "/abs/path/to/store"
```

解決順:

1. `--root` があればそれを使う
2. なければ cwd から上方向に `.shelf/config.toml` を探す
3. 見つからなければ global config の `default_root` を使う
4. global config も無ければ `shelf init --global` を案内して失敗する

## Task ファイル形式

```md
+++
id = "01..."
title = "Example"
kind = "todo"
status = "open"
tags = ["backend", "urgent"] # optional
due_on = "2026-03-31" # optional
estimate_minutes = 120 # optional
spent_minutes = 30 # optional
timer_started_at = "2026-03-31T11:22:33+09:00" # optional
repeat_every = "1w" # optional
archived_at = "2026-03-31T11:22:33+09:00" # optional
parent = "01..." # optional
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

Body text...
```

task file は次の 2 層に分かれます。

- front matter: 構造化 metadata
- body: 自由記述ノート

`shelf show <id>` は metadata と body の両方を表示します。

## 品質ゲート

変更後は最低でも以下を回すのが前提です。

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

## FAQ

### `depends_on` はどういう意味ですか？

`A depends_on B` は「A をやるには B が先」です。
CLI では必ず次の向きで表示します。

`A --depends_on--> B`

### interactive mode はいつでも使えますか？

いいえ。stdin/stdout が TTY のときだけ有効です。
非 TTY では必要な引数やフラグを明示する必要があります。

## ローカル品質ゲート（mise + lefthook）

```bash
mise install
mise run hooks-install
mise run hooks-pre-commit
mise run hooks-pre-push
```

hook の内容:

- `pre-commit`: staged gofmt check + `go test ./...`
- `pre-push`: `go test ./...`, `go test -race ./...`, `go vet ./...`

## 自動バックアップスクリプト

```bash
SHELF_ROOT=/path/to/repo ./scripts/backup_shelf.sh
```

環境変数:

- `SHELF_BIN`（既定: `shelf`）
- `SHELF_BACKUP_DIR`（既定: `${SHELF_ROOT}/.shelf/backups`）
- `SHELF_BACKUP_KEEP`（既定: `30`）

## English Docs

英語版ドキュメントは 1 つ上の `docs/` にあります。

- English README: [`../../README.md`](../../README.md)
- English command guide: [`../COMMAND_GUIDE.md`](../COMMAND_GUIDE.md)
- English CLI spec: [`../COMMANDS.md`](../COMMANDS.md)
