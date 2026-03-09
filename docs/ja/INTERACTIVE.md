# INTERACTIVE（日本語版）

現在の `shelf` は Cockpit-first です。日常的な interactive 操作は基本的に Cockpit 内で行います。

## 主入口

通常は次のどれかから入ります。

- `shelf`
- `shelf cockpit`
- `shelf calendar`
- `shelf tree`
- `shelf board`
- `shelf review`
- `shelf now`

これらは同じ TUI workspace を別 mode で開くだけです。

## 共通操作

- `C`: calendar
- `T`: tree
- `B`: board
- `R`: review
- `N`: now
- `Ctrl+H` / `Ctrl+L`: 前 / 次の mode
- `Tab` / `Shift+Tab`: pane 切り替え
- `?`: help overlay 切り替え
- `q`: help を閉じる、または終了
- `Esc` / `Ctrl+C`: 終了または一時状態から離脱
- `Ctrl+[` : transient overlay から normal に戻る

## Calendar Mode

週表示は日曜始まり、土曜終わりです。

- `t`: 今日へ移動
- `h` / `l`: 日移動
- `j` / `k`: 週移動
- `[` / `]`: 月移動
- `n` / `p`: focused day の task 切り替え
- `a`: 現在文脈で作成
- `A`: quick capture

## Tree Mode

- `h`: subtree を閉じる、または親へ移動
- `l`: subtree を開く
- `m`: 選択 task / mark 済み task を移動
- `v`: 現在 task の mark toggle
- `V`: 範囲 mark の開始 / 終了
- `u`: mark 全解除

## Board Mode

- `h` / `l`: 列移動
- `j` / `k`: 行移動
- `v`: 現在 task の mark toggle
- `V`: 範囲 mark の開始 / 終了
- `u`: mark 全解除

## Review / Now

- `review`: inbox / overdue / blocked / ready を見る operational view
- `now`: 今日の実行に寄せた compact view

## 共通 task 操作

選択 task、または multi-select 中の mark 済み task に適用されます。

- `o`: `open`
- `i`: `in_progress`
- `b`: `blocked`
- `d`: `done`
- `c`: `cancelled`
- `x`: archive toggle
- `z`: snooze presets
- `e`: task file を editor で開く
- `L`: 選択 task の edge file を開く
- `Enter`: compact / detailed inspector 切り替え
- `r`: reload

## 作成

- `a`: 現在 mode の文脈で作成
  - calendar / review / now: focused day を due default にする
  - tree: selected task を parent default にする
  - board: selected column の status を default にする
- `A`: quick capture（`kind=inbox`, `status=open`）

## スクロール

header は固定です。

body の scroll:

- `PgUp` / `PgDn`
- `Ctrl+U` / `Ctrl+D`
- `Home` / `End`

## Selector

候補が多い selector は自動スクロールします。

- hierarchy が重要な場面では tree 形式ラベルを使います
- 必要な場面では `(root)` を明示的に選べます
- 通常の selector では `q`, `Esc`, `Ctrl+C` でキャンセルできます
