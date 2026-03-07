# COMMAND_GUIDE（日本語版）

このファイルは、「どのコマンドを、どんなときに使うべきか」を判断するための実用ガイドです。

- 正式な CLI 仕様: [`COMMANDS.md`](COMMANDS.md)
- ここで扱う内容: 目的、使い分け、典型パターン

## 早見表

| やりたいこと | コマンド |
|---|---|
| `.shelf/` を初期化したい | `shelf init` |
| とにかくすぐ積みたい | `shelf capture` |
| 構造化して task を作りたい | `shelf add` |
| inbox を捌きたい | `shelf triage` |
| 全件をフラットに見たい | `shelf ls` |
| 親子構造で見たい | `shelf tree` |
| 1件を深く見たい | `shelf show` |
| task ファイルを直接編集したい | `shelf edit` |
| メタデータを更新したい | `shelf set` |
| ツリー内で移動したい | `shelf mv` |
| 状態だけ素早く変えたい | `done`, `start`, `block`, `cancel`, `reopen` |
| 今日やることを見たい | `review`, `next`, `today`, `agenda` |
| 日付軸で見たい | `calendar` |
| ステータス列で見たい | `board` |
| タスク同士をつなぎたい | `link`, `unlink`, `links`, `deps` |
| GitHub issue / PR と同期したい | `github`, `sync github` |
| 見積もりや実績を持ちたい | `estimate`, `track` |
| 通知を飛ばしたい | `notify` |
| 再利用テンプレートを使いたい | `template` |
| 保存済みフィルタを使いたい | `view`, `preset` |
| 表示理由やブロック理由を知りたい | `explain` |
| データの壊れを検査したい | `doctor` |
| 全体をバックアップしたい | `export`, `import` |
| 変更を戻したい | `undo`, `redo`, `history` |

## 初期化

### `shelf init`

意味:
`.shelf/` の基本構造とデフォルト設定を作ります。

使う場面:
- 新しいプロジェクトで gitshelf を使い始めるとき
- グローバル fallback shelf を作るとき

例:

```bash
shelf init
shelf init --root /path/to/project
shelf init --global
```

補足:
- 冪等です
- `--force` で config をデフォルトへ戻せます

### `shelf completion`

意味:
shell 補完スクリプトを出力します。

使う場面:
- `zsh` や `bash` で補完を使いたいとき

## 作成と取り込み

### `shelf capture`

意味:
迷わずすぐ記録するための inbox 入口です。

使う場面:
- まだ kind や parent を決めたくない
- 会話中や作業中に思いつきを退避したい

挙動:
- 常に `kind=inbox`
- 常に `status=open`

### `shelf add`

意味:
明示的に task を作ります。

使う場面:
- すでにタイトルや kind、親子関係が決まっている
- 最初から整理した状態で登録したい

例:

```bash
shelf add --title "Weekly Goal"
shelf add --title "Refactor parser" --kind todo --status in_progress
shelf add --title "Monday Plan" --parent root
```

### `shelf template`

意味:
サブツリーを保存し、あとで再展開します。

使う場面:
- 毎週の計画
- 定型的な調査タスク
- 反復する作業分解

主な subcommand:
- `template save`
- `template list`
- `template show`
- `template apply`
- `template delete`

## inbox 処理と日次レビュー

### `shelf triage`

意味:
inbox を処理します。

使う場面:
- `capture` で積んだものを、todo / idea / memo に振り分けたい
- 状態をまとめて更新したい

モード:
- interactive: 1件ずつ見ながら処理
- auto: 一括で同じアクションを適用

### `shelf review`

意味:
日次確認のための集約ビューです。

使う場面:
- inbox、期限切れ、今日、blocked、ready を一度に見たい

### `shelf next`

意味:
今すぐ着手可能な task のみを出します。

使う場面:
- 「で、次に何をやるのか」だけ知りたいとき

### `shelf today`

意味:
overdue と today に集中した確認です。

使う場面:
- 今日の作業計画を立てるとき

特別な使い方:
- `--carry-over`: overdue の active task を今日へ繰り上げる

### `shelf agenda`

意味:
期限ベースで bucket 分けした一覧です。

使う場面:
- `Overdue`, `Today`, `Tomorrow`, `Upcoming`, `Later`, `No due` をまとめて確認したいとき

### `shelf calendar`

意味:
due date カレンダービューです。

使う場面:
- 日付を中心に負荷や予定を見たいとき

補足:
- `calendar` は通常 TUI で開きます。
- 非TTYで使うときは `--json` を使います。
- 明示 range 未指定時は config `[commands.calendar]` が `days/months/years` のどれを使うか決めます。
- `--days` は明示的な day range です。
- `--months` で 1か月単位、3か月単位のレンジをまとめて開けます。
- `--years` で年単位レンジも開けます。
- フォーカス中の日付には task 一覧と本文プレビューが出ます。
- `a` で focused day に task を追加できます。kind/status は config default を使います。
- TUI 内から editor 起動や snooze ができます。
- `o/i/b/d/c` で選択 task の status も直接更新できます。
- 現在の filter から外れる status に変えても、context を失わないよう reload まではその場に残します。
- 現在の表示レンジを超えて移動すると、calendar が自動で過去/未来側へスライドします。

### `shelf board`

意味:
status 列で横断的に見る TUI です。

使う場面:
- いまの進捗を視覚的に整理したい
- 列を見ながら status をその場で更新したい

## 一覧と詳細確認

### `shelf ls`

意味:
最も汎用的なフラット一覧です。

使う場面:
- kind / status / tag で絞り込みたい
- include / exclude を組み合わせたい
- JSON で機械処理したい

例:

```bash
shelf ls --kind todo --status open
shelf ls --not-status done --not-status cancelled
shelf ls --tag backend --ready
```

### `shelf tree`

意味:
`parent` ベースの階層表示です。

使う場面:
- 構造そのものを見たい
- 分解した作業の親子関係を確認したい

### `shelf show`

意味:
1件の task を深く見るためのコマンドです。

表示内容:
- front matter
- 本文
- hierarchy path
- context tree
- readiness
- inbound / outbound links

### `shelf explain`

意味:
その task が「なぜ見えているか」「なぜ blocked なのか」を説明します。

使う場面:
- filter/view に合っている理由を知りたい
- depends_on の影響を確認したい

### `shelf edit`

意味:
task ファイル全体を editor で開きます。

使う場面:
- front matter と本文をまとめて編集したい
- 一括置換や自由編集をしたい

## 更新と移動

### `shelf set`

意味:
構造化された更新コマンドです。

使う場面:
- title, kind, status, tags, due, repeat, parent, body, GitHub links, worklog を更新したい

### `shelf mv`

意味:
親子関係だけを変更します。

使う場面:
- task を別の親の下へ移動したい

### 状態ショートカット

対象:
- `shelf done`
- `shelf start`
- `shelf block`
- `shelf cancel`
- `shelf reopen`

意味:
`set --status ...` の短縮形です。

使う場面:
- ステータス遷移だけを素早く行いたい

### `shelf snooze`

意味:
`due_on` をずらします。

使う場面:
- task 自体は維持しつつ、期限だけ後ろに動かしたい

補足:
- `--by` は現在の期限から相対日数でずらす
- `--to` は新しい期限を直接指定する
- TTY なら未指定時に `Today` / `Tomorrow` / `+3日` / `Next week` などのプリセットを先に選べる
- 同じ selector から custom `by` / `to` 入力にも進める

### `shelf archive` / `shelf unarchive`

意味:
表示対象から退避 / 復帰します。

使う場面:
- 消したくはないが、普段の一覧から外したい

## 関係と依存

### `shelf link`

意味:
task 間に関係を追加します。

link type:
- `depends_on`
- `related`

重要:
`A depends_on B` は「A をやるには B が先」です。

### `shelf unlink`

意味:
既存の関係を外します。

### `shelf links`

意味:
1件の task に対する inbound / outbound link を確認します。

使う場面:
- 依存だけでなく、関連も含めた関係を見たいとき

特別な使い方:
- `--suggest`: `related` 候補を理由付きで提案

### `shelf deps`

意味:
`depends_on` に特化した確認コマンドです。

使う場面:
- 前提 task と逆依存 task を見たい
- 依存グラフを見たい

特別な使い方:
- `--graph`: ASCII graph で表示
- `--transitive`: 再帰的に辿る
- `--suggest`: 前提候補を理由付きで提案

## GitHub、工数、通知

### `shelf github`

意味:
GitHub issue / PR URL を task に紐付けます。

subcommand:
- `github link`
- `github unlink`
- `github show`

### `shelf sync github`

意味:
GitHub の title / state を task 側へ反映します。

更新対象:
- `title`
- `status` (`open` -> `open`, `closed` -> `done`)

### `shelf estimate`

意味:
見積もり時間と消化時間を持たせます。

使う場面:
- 軽い見積もり管理をしたいとき

### `shelf track`

意味:
タイマーを start / stop して `spent_minutes` を積み上げます。

使う場面:
- 実作業時間を簡易に記録したいとき

### `shelf notify`

意味:
期限切れ / 当日期限 task に対してローカルコマンドを実行します。

使う場面:
- デスクトップ通知
- 既存ツールとの軽い連携

## 保存済み view / preset

### `shelf view`

意味:
保存済み filter を管理します。

使う場面:
- `active`, `blocked`, チーム別 view などを名前付きで再利用したい

### `shelf preset`

意味:
コマンド別の出力プリセットを管理します。

使う場面:
- `ls` や `tree` の出力形式・件数・view を固定したい

## 履歴、安全性、入出力

### `shelf undo` / `shelf redo`

意味:
mutating command が作った snapshot を巻き戻し / 再適用します。

### `shelf history`

意味:
更新履歴と snapshot を確認します。

### `shelf doctor`

意味:
`.shelf/` の整合性を検査します。

主な検査対象:
- task metadata
- parent の欠損や循環
- edge の妥当性と重複
- unknown kind / status / tag
- invalid GitHub URL

### `shelf export`

意味:
config / tasks / edges を JSON として書き出します。

### `shelf import`

意味:
export JSON を読み込みます。

モード:
- validate-only
- dry-run
- merge
- replace

## 推奨フロー

### 日次

```bash
shelf capture "interrupt / idea"
shelf triage
shelf review
shelf next
shelf today
```

### 週次

```bash
shelf agenda --days 14
shelf calendar
shelf tree
shelf template apply weekly-plan
shelf doctor --strict
```
