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
- `shelf board`（TTY 専用 TUI）
- `shelf calendar --days <n>`（`n > 7` のとき TTY 専用 TUI）

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
3. `snooze` は `--by` / `--to` 未指定時に `By days` / `To date token` を選び、値を入力する
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

- status 列は config の順序に従う
- 下部に選択 task の body preview を表示
- `o` / `s` / `b` / `d` / `c` で status をその場で更新

## `calendar` TUI

- `shelf calendar --days <n>` で `n > 7` のときに使います
- 選択中の日付を中心に月グリッドを表示します
- `h` / `l`: 1日ずつ移動
- `j` / `k`: 1週ずつ移動
- `g` / `G`: レンジの先頭 / 末尾へ移動
- `q` / `Esc` / `Ctrl+C`: 終了
