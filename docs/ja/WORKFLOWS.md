# WORKFLOWS（日本語版）

現状の `shelf` をどう使うのが自然かをまとめた実用ガイドです。

- CLI 仕様: [`COMMANDS.md`](COMMANDS.md)
- コマンドごとの使い分け: [`COMMAND_GUIDE.md`](COMMAND_GUIDE.md)
- interactive の詳細: [`INTERACTIVE.md`](INTERACTIVE.md)

## 主な入口

現在のツールは、1つの主 workspace と、その起動用コマンド群という構成です。

- `shelf`
  - TTY なら Cockpit を直接開く
  - 非TTYなら help を表示
- `shelf cockpit`
  - 主入口の interactive workspace を明示的に開く
- `shelf calendar`
  - Cockpit の `calendar` mode で開く
- `shelf tree`
  - TTY では Cockpit の `tree` mode で開く
  - `--plain` で text tree
- `shelf board`
  - Cockpit の `board` mode で開く
- `shelf review`
  - TTY では Cockpit の `review` mode で開く
- `shelf now`
  - TTY では Cockpit の `now` mode で開く
  - `today` は alias のまま

基本方針:

- 普段は `shelf` または `shelf cockpit`
- 用途が狭い入口や script 用には個別コマンド

## おすすめの日次フロー

### 1. とにかく積む

まだ整理したくないものは `capture` を使います。

```bash
shelf capture "Call vendor"
shelf capture "Investigate parser regression" --tag backend --due tomorrow
```

結果:

- `kind=inbox`
- `status=open`

### 2. inbox を捌く

`triage` で inbox を処理します。

```bash
shelf triage
shelf triage --auto done
```

向いている作業:

- kind/status の整理
- 不要項目の archive
- inbox から正式 task への昇格

### 3. Cockpit で作業する

主 workspace を開きます。

```bash
shelf
# または
shelf cockpit
```

mode 切り替え:

- `C`: calendar
- `T`: tree
- `B`: board
- `R`: review
- `N`: now

### 4. 1件を深く見る

詳細確認や直接編集が必要なとき:

```bash
shelf show <id>
shelf edit <id>
shelf set <id> --status blocked --append-body "Waiting on API answer."
```

### 5. 構造を整える

親子構造を直したいとき:

```bash
shelf tree
shelf mv <id> --parent root
```

または Cockpit の `tree` mode で:

- `h` / `l`: 開閉
- `m`: 現在 task / mark 済み task を移動
- `v` / `V`: 単体選択 / 範囲選択

## `capture` と `add` の使い分け

`capture` を使う場面:

- 速さ優先
- まだ kind や parent を決めたくない
- inbox として集めたい

`add` を使う場面:

- もう正式 task として作りたい
- kind/status/parent を決めて登録したい
- review しながら interactive に作りたい

例:

```bash
shelf capture "Check flaky CI failure"
shelf add --title "Refactor parser" --kind todo --status in_progress --parent root
```

## 現在の `add` interactive flow

TTY で `shelf add` を `--title` なしで実行すると、現在の flow は次です。

1. `Kind` 選択
2. `Status` 選択
3. Review 画面で編集
   - `Title`
   - `Kind`
   - `Status`
   - `Tags`
   - `Due`
   - `Repeat`
   - `Parent`
4. 作成またはキャンセル

重要:

- `Title` は review 画面で編集する
- `Title` は必須
- `Ctrl+S` でそのまま作成できる
- `Ctrl+Enter` も端末が対応していれば使える
- parent 選択は tree 形式
- 候補が多いときは自動スクロールする

## Cockpit の mode

### `calendar`

向いている場面:

- 日付基準で考えたい
- 期限ベースで計画したい
- focused day に直接 task を追加したい

主な操作:

- `t`: 今日へ移動
- `n` / `p`: focused day の task 切り替え
- `a`: focused day に task 追加
- `o/i/b/d/c`: status 変更

### `tree`

向いている場面:

- 親子構造の分解が重要
- subtree の整理をしたい

主な操作:

- `h` / `l`: 開閉
- `m`: 現在 / mark 済み task を移動
- `v`: 単体 mark
- `V`: 範囲 mark
- `u`: mark 全解除

### `board`

向いている場面:

- status 基準で考えたい
- 複数 task の status をまとめて変えたい

主な操作:

- `v` / `V`: mark
- `u`: mark 全解除
- `o/i/b/d/c`: 選択または mark 済み task の status 更新

### `review`

向いている場面:

- 日次確認をコンパクトにやりたい

section:

- `Inbox`
- `Overdue`
- `Today`
- `Blocked`
- `Ready`

### `now`

向いている場面:

- 今日の実行に集中したい

主表示:

- `Focused Day`
- `Overdue`
- `Today`

## Cockpit を使わない方がよい場面

次のようなときは plain command の方が向いています。

- script で使う
- JSON が欲しい
- workspace ではなく、単発の結果が欲しい

例:

```bash
shelf ls --status open --json
shelf review --plain
shelf tree --plain
shelf calendar --json --months 3
```

## よく使う保守コマンド

### status

```bash
shelf start <id>
shelf block <id>
shelf done <id>
shelf cancel <id>
shelf reopen <id>
```

### metadata と本文

```bash
shelf set <id> --tag backend
shelf set <id> --due next-week
shelf set <id> --append-body "Need benchmark results."
```

### スケジュール調整

```bash
shelf snooze <id> --by 2d
shelf snooze <id> --to tomorrow
```

### 関係づけ

```bash
shelf link --from <a> --to <b> --type depends_on
shelf link --from <a> --to <b> --type related
shelf deps <id> --transitive
shelf links <id> --transitive
```

## 現在の用語

- 進捗用語は `status` に統一
- 既定 status:
  - `open`
  - `in_progress`
  - `blocked`
  - `done`
  - `cancelled`
- link type:
  - `depends_on`
  - `related`

`depends_on` の向き:

- `A depends_on B` は「A をやるには B が先」

## 迷ったらここから

まずはこれで十分です。

```bash
shelf
```

そのあと:

1. `R` で review
2. `N` で now
3. `T` で構造確認
4. `B` で status 整理
5. `C` で日付確認
