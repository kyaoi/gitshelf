# SPEC（日本語版）

## 目的

- Git リポジトリ内で TODO / IDEA / MEMO を軽量に管理する
- 目標 → 日 → 具体タスクのような無限ネストの親子ツリーを表現する
- 子同士・別枝同士の関係を link（edge）として表現する
- 端末だけで完結し、差分が読みやすく、マージ衝突を抑えやすい保存形式にする

## コア概念

### Task

- 作業、アイデア、メモの単位
- `kind` と `status` を持つ
- 任意で `due_on` を持てる
- 任意で `parent` を持てる
- 本文は自由記述

### Tree

- 親子による内包 / 分解を表す
- 各 task の親は 0 または 1
- 深さ制限なし

### Edge

- 親子ではない関係を表す
- outbound edge を `.shelf/edges/<src_id>.toml` に保存
- inbound edge は edge file 全体を走査して逆引きする

## 保存ディレクトリ

- `.shelf/config.toml`
- `.shelf/tasks/`
- `.shelf/edges/`
- `.shelf/templates/`
- `.shelf/history/`

## Kind と Status

- `kind`: task の種類
- `status`: 進捗状態
- 両者は独立
- 例: `kind=idea` かつ `status=done` も許可

## link の向き

- 使える link type は `depends_on`, `related` のみ
- `A depends_on B` は「A をやるには B が先」
- 表示は必ず `A --depends_on--> B`
- `related` は緩い関連づけ

## 非機能要件

- 差分が安定していること
- 原子的更新であること
- 破損時に壊れたファイルが分かること
- task 数が増えても interactive 検索で操作できること

## 初期リリースでやらないこと

- 常駐デーモン / サーバ
- DB（SQLite など）
- Web UI / GUI
- 複数親

補足:
- 通常操作は CLI 前提
- ただし status を横断して扱う `board` TUI は許可

## 運用補助コマンド

- `agenda`: 期限ベースの確認
- `today`: overdue + today の集中確認
- `today --carry-over`: overdue active task を今日に繰り上げ
- `snooze`: `due_on` を相対または絶対日付で調整
- `archive` / `unarchive`: 一覧から退避 / 復帰
- `done` / `reopen`: status ショートカット
- `--view`: 保存済み filter の適用
- `preset`: 出力プリセット管理
- `view`: 保存済み filter 管理
- `explain`: readiness / view 一致理由の説明
- `deps`: depends_on の前提・逆依存確認
- `deps --graph`: ASCII tree による可視化
- `undo` / `redo`: ロールバック / 再適用
- `history`: 更新履歴の確認
- mutating command は `.shelf/.write.lock` を使って排他
- `doctor --strict`: 期限未設定などの warning を追加検出
- `export` / `import`: `.shelf` 全体の backup / restore

## 用語

- root: 親を持たない task
- subtree: ある task 配下のツリー
- edge: link の 1 本

詳細は [`STORAGE.md`](STORAGE.md) と [`COMMANDS.md`](COMMANDS.md) を参照してください。
