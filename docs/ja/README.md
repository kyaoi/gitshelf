# gitshelf 日本語ガイド

`gitshelf` は、1つの TUI workspace `Cockpit` を中心にした Git-friendly task manager です。

- コマンド名: `shelf`
- 主入口: `shelf` または `shelf cockpit`
- 保存先: `.shelf/`
- タスク: `.shelf/tasks/<id>.md`
- リンク: `.shelf/edges/<src_id>.toml`

## 日本語ドキュメント

- CLI 仕様: [`COMMANDS.md`](COMMANDS.md)
- コマンドガイド: [`COMMAND_GUIDE.md`](COMMAND_GUIDE.md)
- ワークフロー: [`WORKFLOWS.md`](WORKFLOWS.md)
- 対話 UI: [`INTERACTIVE.md`](INTERACTIVE.md)
- 保存形式: [`STORAGE.md`](STORAGE.md)

## インストール

```bash
go build -o shelf ./cmd/shelf
```

## クイックスタート

```bash
./shelf init

# 主入口
./shelf
./shelf cockpit

# launcher
./shelf calendar
./shelf tree
./shelf board
./shelf review
./shelf now

# script / 確認
./shelf ls --status open --json
./shelf next

# completion
./shelf completion zsh
```

## 現在の公開コマンド面

現在の top-level command は次だけです。

- `shelf init`
- `shelf completion`
- `shelf cockpit`
- `shelf calendar`
- `shelf tree`
- `shelf board`
- `shelf review`
- `shelf now`（`today` は alias）
- `shelf ls`
- `shelf next`

それ以外の操作は Cockpit 内で完結させる前提です。

## 基本方針

- 普段は `shelf`
- 明示したいときは `shelf cockpit`
- `calendar/tree/board/review/now` は Cockpit の起動プリセット
- 作成・編集・移動・期限変更・リンク・archive・status 変更は Cockpit 内で行う
