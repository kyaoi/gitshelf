# INTERACTIVE（対話 UI 仕様 日本語版）

interactive mode は stdin/stdout が TTY のときだけ有効です。
非 TTY では、必要な引数やフラグを明示する必要があります。

## 対応コマンド

- `shelf add`（`--title` 省略時）
- `shelf link`（`--from/--to/--type` 省略時）
- `shelf unlink`（`--from/--to/--type` 省略時）
- `shelf show`（`<id>` 省略時）
- `shelf explain`（`<id>` 省略時）
- `shelf edit`（`<id>` 省略時）
- `shelf set`（`<id>` 省略時）
- `shelf snooze`（`<id>` または `--by/--to` 省略時）
- `shelf mv`（`<id>` や `--parent` 省略時）
- `shelf done`（`<id>` 省略時、`status!=done` を優先表示）
- `shelf links`（`<id>` 省略時）
- `shelf triage`（`--auto` なし）
- `shelf board`（TTY 専用、Daily Cockpit の `board` mode）
- `shelf calendar`（`--json` なしでは Daily Cockpit の `calendar` mode）
- `shelf tree`（`--plain` / `--json` なしでは Daily Cockpit の `tree` mode）
- `shelf cockpit`（TTY 専用、統一 Daily Cockpit 入口）

## キーバインド

- `j` / `k`: 下 / 上へ移動
- `Enter`: 決定
- `/`: 検索モード
- `?`: ヘルプ表示切り替え
- `Esc`:
  - 検索モード中: 検索をクリアして抜ける
  - 通常時: キャンセル
- `q`: 通常時のキャンセル
- `Ctrl+C`: キャンセル
- `ArrowUp` / `ArrowDown` も利用可能

## 検索

- `/` でインクリメンタル検索に入ります。
- label / search text に対して一致判定します。
- task selector では title と short/full ID が対象です。

## 表示規則

task 候補行の基本形:

- デフォルト: `{tree-prefix}{title}`
- `--show-id` 時: `[{short}] {tree-prefix}{title}`

表示ルール:
- デフォルトでは ID 非表示
- `--show-id` / `-i` で short ID を表示
- task selector は本文プレビューを表示
- 本文が空なら `(empty body)` を表示
- enum selector は本文プレビューなし
- 選択行、prompt、help line、preview header は TTY で色付け
- `NO_COLOR=1` で色無効、`CLICOLOR_FORCE=1` で色強制

## `add` の対話フロー

1. `Title` 入力
2. `Kind` 選択
3. `Status` 選択
4. Review 画面で `Title` / `Kind` / `Status` / `Tags` / `Due` / `Repeat` / `Parent` を確認・編集
5. `Create task` または `Cancel`

tags selector では以下ができます。

- config にある tag の on/off
- 新規 freeform tag の追加
- 選択済み tag の全解除

Parent 候補は tree 形式で表示します。例:

```text
(root)
週目標
├─ 月曜日
│  └─ 英単語100個
```

## `link` の対話フロー

1. source task 選択
2. destination task 選択
3. link type 選択

type 選択画面では、次の注意文を出します。

`A depends_on B = B must be done before A`

## `unlink` の対話フロー

1. source task を選択
2. その task の outbound edge から削除対象を選択

## `show` / `explain` / `edit` / `set` / `done` / `links` / `snooze`

1. 対象 task を tree 形式で選択
2. `set` だけは、更新フラグ未指定時に menu で編集対象を選ぶ
3. `snooze` は `--by` / `--to` 未指定時に `Today` / `Tomorrow` / `By +3 days` などのプリセット、または custom 入力を選ぶ
4. `set` は適用前に change preview を表示

`set` の interactive 編集対象:

- `Title`
- `Kind`
- `Status`
- `Tags`
- `Due`
- `Repeat`
- `Parent`
- `Body replace`
- `Body append`

## `mv` の対話フロー

1. 対象 task を選択（`<id>` 省略時）
2. 新しい parent を選択（`--parent` 省略時）

## `triage` の対話フロー

1. `--kind` + `--status` で対象読み込み（既定 `inbox/open`）
2. 各 task に対して以下から選択
   - `Edit fields`
   - `Set status ...`
   - `Archive task`
   - `Skip`
   - `Quit triage`

## `board` TUI

- `shelf board` は shared Daily Cockpit を `board` mode で開きます
- status 列は config の順序に従います
- `C/T/B/R/N` で他 mode へ切り替えられます

## `calendar` TUI

- `shelf cockpit` / `shelf calendar` / `shelf tree` / `shelf board` で共通利用されます
- `shelf review` / `shelf now` も TTY ではこの TUI を使います（`--plain` / `--json` 指定時を除く）
- レイアウトは `main + right sidebar` です
- `C/T/B/R/N`: mode 切り替え
- `Tab` / `Shift+Tab`: ペイン切り替え
- `h` / `l`: calendar mode では1日移動、右 sidebar に focus があるときは sidebar calendar を1日移動、それ以外では review/now tab 切り替えまたは board 列移動
- `j` / `k`: calendar mode では1週移動、右 sidebar に focus があるときは sidebar calendar を1週移動、それ以外では tree/board/review/now の行移動
- `[` / `]`: 1か月ずつ移動
- `g` / `G`: レンジ先頭 / 末尾、または section 先頭行 / 末尾行へ移動
- calendar mode では month grid を大きく表示し、focused day task list は inspector の上に出ます
- calendar 以外の mode では、右 sidebar に compact calendar が inspector の上に出ます。`Tab` で focus すると日付を直接動かせます
- `n` / `p`: calendar mode では focused day task 切り替え、review/now では tab 切り替え、board では列移動
- `now` mode では `Focused Day` / `Overdue` / `Today` を main pane に同時表示します
- header と mode tabs は上部に固定されます
- `PgUp` / `PgDn` または `Ctrl+U` / `Ctrl+D`: body をスクロール
- `Home` / `End`: body の先頭 / 末尾へ移動
- `1..6`: 見えている section へ直接ジャンプ
- `a`: focused day 用の inline add composer を開く
- `o` / `i` / `b` / `d` / `c`: 選択 task の status を `open` / `in_progress` / `blocked` / `done` / `cancelled` に変更
- `Enter`: compact / detailed inspector の切り替え
- `e`: 選択 task を editor で開く
- `z`: 選択 task の snooze プリセットを開く
- `r`: 再読み込み
- `q` / `Esc` / `Ctrl+C`: 終了
- 現在の表示レンジを越えて移動すると、自動で calendar の表示範囲がずれる
