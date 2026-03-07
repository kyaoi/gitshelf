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

### 方法1: リポジトリを clone して `./bin/shelf` を作る

```bash
git clone https://github.com/kyaoi/gitshelf.git
cd gitshelf
mkdir -p bin
go build -o ./bin/shelf ./cmd/shelf
```

リポジトリ内からそのまま実行できます。

```bash
./bin/shelf
```

どこからでも `shelf` として呼びたい場合:

```bash
export PATH="$PWD/bin:$PATH"
shelf
```

### 方法2: `go install` で直接入れる

```bash
go install github.com/kyaoi/gitshelf/cmd/shelf@latest
```

インストール先は `$(go env GOPATH)/bin` または `$(go env GOBIN)` です。

## completion

shell ごとの completion を生成できます。

```bash
./bin/shelf completion zsh
./bin/shelf completion bash
./bin/shelf completion fish
./bin/shelf completion powershell
```

例:

### zsh

```bash
mkdir -p "${HOME}/.zsh/completions"
./bin/shelf completion zsh > "${HOME}/.zsh/completions/_shelf"
echo 'fpath=("${HOME}/.zsh/completions" $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
```

### bash

```bash
mkdir -p "${HOME}/.local/share/bash-completion/completions"
./bin/shelf completion bash > "${HOME}/.local/share/bash-completion/completions/shelf"
```

### fish

```bash
mkdir -p "${HOME}/.config/fish/completions"
./bin/shelf completion fish > "${HOME}/.config/fish/completions/shelf.fish"
```

### PowerShell

```powershell
./bin/shelf completion powershell | Out-String | Invoke-Expression
```

## クイックスタート

```bash
./bin/shelf init

# 主入口
./bin/shelf
./bin/shelf cockpit

# launcher
./bin/shelf calendar
./bin/shelf tree
./bin/shelf board
./bin/shelf review
./bin/shelf now

# script / 確認
./bin/shelf ls --status open --json
./bin/shelf next
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
- `shelf now`
- `shelf ls`
- `shelf next`

それ以外の操作は Cockpit 内で完結させる前提です。

## 基本方針

- 普段は `shelf`
- 明示したいときは `shelf cockpit`
- `calendar/tree/board/review/now` は Cockpit の起動プリセット
- 作成・編集・移動・期限変更・リンク・archive・status 変更は Cockpit 内で行う
