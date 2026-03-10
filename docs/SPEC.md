# SPEC（初期リリース: 完結版）

## 目的
- Gitリポジトリ内で、TODO/IDEA/MEMO を軽量に管理する
- 目標→日→具体タスクのような **無限ネスト（親子ツリー）** を表現する
- 子同士・別枝同士の関係（依存/関連）を **リンク（edges）** として表現する
- 端末で完結し、差分が読みやすく、マージ衝突が起きにくい保存形式にする

## コア概念
### タスク（Task）
- 1つの作業/アイデア/メモの単位
- `kind` と `status` を持つ
- 任意で `due_on`（`YYYY-MM-DD`）を持てる
- `parent` を持てる（親は最大1つ、rootは親なし）
- 本文（メモ）は任意

### 親子（Tree）
- 内包/分解を表す
- 各タスクは親を0または1つ持つ（ツリーになる）
- 深さは無制限（無限入れ子）

### リンク（Edges）
- 内包ではない関係（依存/関連）
- outbound edges を `.shelf/edges/<src_id>.toml` に保存
- inbound edges は全edge走査で逆引きして表示

## 保存ディレクトリ
- ルートに `.shelf/` を置く
- `.shelf/config.toml` に user setting を置く
- `.shelf/tasks/` にタスク本体
- `.shelf/edges/` にリンク

## Kind と Status
- `kind`: タスクの種類（例: todo/idea/memo…）
- `status`: 進捗（`open`, `in_progress`, `blocked`, `done`, `cancelled`）
- 両者は独立であり、`kind=idea` でも `status=done` は許可（実装は制限しない）

## リンクの向き（重要な不変条件）
- 使用可能な link type は config の `link_types.names` で定義する
- そのうち readiness / cycle check / ancestor restriction に使う blocking relation は `link_types.blocking` で定義する
- デフォルトでは `depends_on` が blocking relation で、`A depends_on B` は **Aを行うにはBが先** を意味する
- 表示は常に `A --<type>--> B` とする

## 非機能要件（初期リリースで守る）
- **差分が安定**: ソート順とフォーマットを固定
- **原子的更新**: 一時ファイルに書いて `rename` で置換
- **壊れても復旧可能**: 形式が壊れたら、どのファイルが壊れているか明確にエラー表示
- **大きくなっても使える**: 対話選択に検索・ページングがある

## 仕様外（やらない）
- 常駐デーモン/サーバ
- DB（SQLite等）
- Web UI / GUI
- 複数親（DAGツリー）は採用しない（親は1つ）

補足:
- 通常操作はCLI前提
- ただし status を横断して扱うための専用 `board` TUI は許可する

## 公開CLI

現在の公開コマンドは次のみ:

- `cockpit`
- `calendar`
- `tree`
- `board`
- `review`
- `now`
- `link`
- `unlink`
- `links`
- `ls`
- `next`
- `init`
- `completion`

通常操作は `Cockpit` に集約する。

- `shelf` を TTY で実行すると `Cockpit` が開く
- `calendar/tree/board/review/now` は `Cockpit` の起動プリセット
- `ls` と `next` は read-only query 用
- `link` / `unlink` / `links` は script から link を扱う command

日付入力で使う token:

- `today`
- `tomorrow`
- `+Nd`
- `-Nd`
- `next-week`
- `this-week`
- `mon..sun`
- `next-mon..next-sun`
- `in N days`

## 用語
- root: 親を持たないタスク
- subtree: あるタスク配下のツリー
- edge: linksの1本

詳細は `docs/STORAGE.md` と `docs/COMMANDS.md` を参照。
