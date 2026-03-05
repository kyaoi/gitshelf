# gitshelf

Gitで TODO / IDEA / MEMO を管理するための **軽量CLI**。

- コマンド: `shelf`
- 保存: `.shelf/`（リポジトリ内に置く）
- タスク: `.shelf/tasks/<id>.md`（フラット、1タスク=1ファイル）
- リンク: `.shelf/edges/<src_id>.toml`（outbound edgesを分離）
- 階層: タスクの `parent` により無限に入れ子
- 関連: edgesにより子同士/別枝同士の関係を表現
- 分類: `kind`（todo/idea/memo/...）
- 進捗: `state`（open/done/...）

詳細は `START_HERE.md` と `docs/` を参照。
