# COMMAND GUIDE（日本語版）

どのコマンドを使うべきかの短いガイドです。

## 普段は `shelf`

TTY で普通に作業したいなら、まずこれです。

```bash
shelf
```

これで主 workspace の `Cockpit` が開きます。

## 明示したいときは `shelf cockpit`

```bash
shelf cockpit
shelf cockpit --mode tree
shelf cockpit --mode board --months 3
```

## 開始 view が決まっているなら launcher

- `shelf calendar`
- `shelf tree`
- `shelf board`
- `shelf review`
- `shelf now`

これらは同じ TUI の開始プリセットです。

## 日常編集は Cockpit の中で行う

- `a`: 選択 task の子として追加
- `A`: root に追加
- `K`: kind 編集
- `#`: tag 編集
- `o/i/b/d/c`: status 変更
- `L` / `U`: link の追加 / 削除
- `y`: title コピー

完全なキー一覧は [`INTERACTIVE.md`](INTERACTIVE.md) を参照してください。

## text / JSON が欲しいなら `shelf ls`

```bash
shelf ls --status open
shelf ls --kind todo --json
shelf ls --ready --json
```

## 着手候補だけ欲しいなら `shelf next`

```bash
shelf next
shelf next --json
```

## script から link を触るなら standalone command

```bash
shelf link --from 01AAA --to 01BBB --type depends_on
shelf unlink --from 01AAA --to 01BBB --type depends_on
shelf links 01AAA
```

## 初期化や整理は `shelf init`

```bash
shelf init
shelf init --force
shelf init --global
```

## shell 設定だけなら `shelf completion`

```bash
shelf completion zsh
shelf completion bash
```
