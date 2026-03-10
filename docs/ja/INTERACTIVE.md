# INTERACTIVE（日本語版）

現在の `shelf` は Cockpit-first です。日常的な interactive 操作は基本的に Cockpit 内で行います。

詳細な keybind はこのファイルに集約します。

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
- `Tab` / `Shift+Tab`: non-calendar mode で pane 切り替え
- `?`: help overlay 切り替え
- `q`: help を閉じる、または終了
- `Esc`: 終了または一時状態から離脱
- `Ctrl+[` : popup や入力 mode を抜けて normal に戻る

transient picker / composer は中央 popup で表示します。
件数の多い一覧は box の高さを変えず、内部スクロールで閲覧します。

## Calendar Mode

週表示は日曜始まり、土曜終わりです。

親 task に due がある場合、子孫 task が自前の `due_on` を持たなくても、その日付に継承表示されます。

- `t`: 今日へ移動
- `h` / `l`: 日移動
- `j` / `k`: 週移動
- `[` / `]`: 月移動
- `n` / `p`: selected day の task 切り替え
- `a`: 選択 task の子として作成。未選択時は root に作成
- `A`: root に作成

calendar main view では pane focus 切り替えは使いません。

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

- `K`: 選択 task の kind を編集
- `#`: 選択 task の tag を編集
- `y`: 選択 title、または mark 済み title 群を config の区切り文字でコピー
- `Y`: 選択 task の subtree、または mark 済み subtree 群をインデント付き title tree としてコピー
- `P`: 選択 task file path、または mark 済み file path 群を絶対パスでコピー
- `O`: 選択 task の本文、または mark 済み task 群の本文をコピー
- `o`: `open`
- `i`: `in_progress`
- `b`: `blocked`
- `d`: `done`
- `c`: `cancelled`
- `x`: archive toggle
- `z`: snooze presets
- `e`: task file を editor で開く
- `L`: 選択 task から link を追加
- `U`: 選択 task の outbound link を1つ削除
- `Enter`: compact / detailed inspector 切り替え
- `r`: reload

link selector は tree 風ラベルを使い、件数が多いときはスクロールします。
ID は `--show-id` を有効にしたときだけ表示します。
- link picker では `h` / `l` で Tree mode と同様に開閉できます
- link type の切り替えは `Tab` / `Shift+Tab` です

## 作成

- `a`: 現在 mode の文脈で、選択 task の子として作成
- 未選択時は `a` でも root 作成になる
- `A`: 現在 mode の文脈で、root に作成
- calendar / review / now では focused day を due default にする
- board では selected column の status を default にする
- add composer は title と kind を同じ box 内で編集する
- `Tab` / `Shift+Tab` で title / kind を循環する
- title フィールド中は `Left` / `Right` でカーソル移動できる
- kind フィールド中は `j` / `k` だけで kind を切り替える
- `Enter` で作成を確定する
- `Esc` / `Ctrl+[` で add をキャンセルする
- title 入力中の `q` は通常文字として入力できる

## Filter

- `f`: popup の filter editor を開く
- `status` / `kind` の include / exclude を編集できる
- 適用した filter は Cockpit の各 mode に共通で効く

## Tag

- `Space`: 現在の tag を toggle
- `Enter` on `Done`: 保存して閉じる
- `Enter` on `+ Add new tag`: 入力 mode に入る
- `Ctrl+S`: tag 編集中のどこからでも保存して閉じる
- 新規 tag 入力中は `Left` / `Right` でカーソル移動でき、文字はその位置に挿入されます

## Non-Calendar Sidebar

- 右ペインは `Calendar / Selected Day / Inspector` の3段です
- 高さ比率は `Calendar 40% / gap 1% / Selected Day 28% / gap 1% / Inspector 30%` です
- main selection は sidebar の日付と `Selected Day` に同期します
- sidebar の calendar を動かすと、その日に visible task があれば main selection も追従します
- non-calendar mode では `Selected Day` 上の `n` / `p` でも main selection が追従します
- focus が calendar にあるときは main pane と同様に枠線を強調表示します

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
- 通常の selector では `q`, `Esc` でキャンセルできます
- Link は `/` で query 入力 mode に入り、入力中は `Left` / `Right` でカーソル移動でき、文字はその位置に挿入されます
- 旧 focused day panel 名は `Selected Day` に統一され、main selection と同期します
- `Selected Day` は sidebar の calendar で日付を変えたときも追従します
