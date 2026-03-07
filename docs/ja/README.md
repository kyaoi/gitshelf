# gitshelf 日本語ガイド

`gitshelf` は、1つの TUI ワークスペース `Cockpit` を中心とした、Git-friendly なタスクマネージャーです。

* CLI コマンド: `shelf`
* メイン入口: `shelf` または `shelf cockpit`
* 保存ルート: `.shelf/`
* タスク: `.shelf/tasks/<id>.md`
* リンク: `.shelf/edges/<src_id>.toml`

## ドキュメント

* CLI 仕様: [`docs/COMMANDS.md`](docs/COMMANDS.md)
* コマンドガイド: [`docs/COMMAND_GUIDE.md`](docs/COMMAND_GUIDE.md)
* ワークフローガイド: [`docs/WORKFLOWS.md`](docs/WORKFLOWS.md)
* 対話動作: [`docs/INTERACTIVE.md`](docs/INTERACTIVE.md)
* 保存形式: [`docs/STORAGE.md`](docs/STORAGE.md)
* 日本語ユーザードキュメント: [`docs/ja/README.md`](docs/ja/README.md)

## Install

### 推奨: `go install` で直接インストール

```bash
go install github.com/kyaoi/gitshelf/cmd/shelf@latest
```

### ローカル開発: clone してビルド

```bash
git clone https://github.com/kyaoi/gitshelf.git
cd gitshelf
go install ./cmd/shelf
```

## Shell Completion

利用しているシェル向けの completion を生成できます。

```bash
shelf completion zsh
shelf completion bash
shelf completion fish
shelf completion powershell
```

例:

### zsh

```bash
mkdir -p "${HOME}/.zsh/completions"
shelf completion zsh > "${HOME}/.zsh/completions/_shelf"
echo 'fpath=("${HOME}/.zsh/completions" $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
```

### bash

```bash
mkdir -p "${HOME}/.local/share/bash-completion/completions"
shelf completion bash > "${HOME}/.local/share/bash-completion/completions/shelf"
```

### fish

```bash
mkdir -p "${HOME}/.config/fish/completions"
shelf completion fish > "${HOME}/.config/fish/completions/shelf.fish"
```

### PowerShell

```powershell
shelf completion powershell | Out-String | Invoke-Expression
```

## Quick Start

```bash
# 初期化
shelf init

# メインワークスペース
shelf
shelf cockpit

# Cockpit ランチャー
shelf calendar
shelf tree
shelf board
shelf review
shelf now

# スクリプト向けクエリ
shelf ls --status open --json
shelf next
```

## Command Surface

現在の公開 CLI に含まれる top-level command は次のものだけです。

* `shelf init`
* `shelf completion`
* `shelf cockpit`
* `shelf calendar`
* `shelf tree`
* `shelf board`
* `shelf review`
* `shelf now`
* `shelf ls`
* `shelf next`

それ以外の操作はすべて Cockpit 内で行う想定です。

## Cockpit-First Usage

`Cockpit` がメインワークスペースです。

* TTY 上で `shelf` を実行すると `Cockpit` が開く
* `shelf cockpit` で明示的に開ける
* `calendar/tree/board/review/now` は同じワークスペースに対するランチャープリセット
* 作成、編集、移動、スヌーズ、リンク、アーカイブ、ステータス変更は TUI 内で行う

推奨される開始地点:

```bash
shelf
```

## Current Data Model

タスクのメタデータ:

* `title`
* `kind`
* `status`
* `tags`
* `due_on`
* `repeat_every`
* `archived_at`
* `parent`
* timestamps

リンクで使うのは次の 2 種類のみです。

* `depends_on`
* `related`

## Quality Checks

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

