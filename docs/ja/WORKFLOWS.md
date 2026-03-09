# WORKFLOWS（日本語版）

現在の Cockpit-first な使い方です。

## 基本ルール

- まずは `shelf`
- 普段の作業は Cockpit に留まる
- text/JSON の単発確認だけ `ls` / `next` を使う

## 主な入口

- `shelf`
  - TTY なら Cockpit を直接開く
- `shelf cockpit`
  - 同じ workspace を明示的に開く
- `shelf calendar`
  - calendar mode で開く
- `shelf tree`
  - tree mode で開く
- `shelf board`
  - board mode で開く
- `shelf review`
  - review mode で開く
- `shelf now`
  - now mode で開く

## おすすめの日次フロー

### 1. workspace を開く

```bash
shelf
```

### 2. mode を切り替える

Cockpit 内:

- `C`: calendar
- `T`: tree
- `B`: board
- `R`: review
- `N`: now
- `Ctrl+H` / `Ctrl+L`: 前 / 次の mode

### 3. まずは calendar で見る

主なキー:

- `t`: 今日へ移動
- `h/l`: 日移動
- `j/k`: 週移動
- `[` / `]`: 月移動
- `n/p`: focused day の task 切り替え
- `a`: 現在文脈で作成
- `A`: quick capture

### 4. その場で更新する

主なキー:

- `o`: open
- `i`: in progress
- `b`: blocked
- `d`: done
- `c`: cancelled
- `x`: archive toggle
- `z`: snooze

### 5. tree で構造を直す

主なキー:

- `h/l`: collapse / expand
- `m`: 選択 task / mark 済み task を移動
- `v`: 単体 mark
- `V`: 範囲 mark の開始 / 終了
- `u`: mark 全解除

### 6. 直接答えが欲しいときだけ `ls` / `next`

```bash
shelf ls --status open --json
shelf ls --kind todo --not-status done --not-status cancelled
shelf next
shelf next --json
```

## mode の使い分け

- `calendar`: 日付基準で計画する
- `tree`: 親子構造や移動を扱う
- `board`: status 基準で整理する
- `review`: inbox / overdue / blocked / ready を俯瞰する
- `now`: 今日の実行に集中する

## もう top-level command ではないもの

現在の設計では、次の workflow は standalone command 前提ではありません。

- add/edit/show/set
- move/snooze/archive
- history/undo/redo
- import/export
- GitHub sync
- saved view / preset

これらは Cockpit 内で完結させる方針です。
