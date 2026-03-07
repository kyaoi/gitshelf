# IMPLEMENTATION_PLAN（初期リリース: タスク分割）

> ここでは“完結版”を一気に作るため、依存順にタスクを並べます。  
> 各タスクは1コミット推奨（AGENTS.md参照）。

## GS-01: プロジェクト雛形
- Go module 初期化
- `cmd/shelf` エントリ追加（cobra等は任意、標準flagでも可）
- `internal/` の骨格

受け入れ条件:
- `shelf --help` が動く

## GS-02: init 実装
- `.shelf/` 作成（冪等）
- デフォルト `config.toml` 作成（既存なら触らない）

受け入れ条件:
- `shelf init` 2回実行しても壊れない

## GS-03: データモデル & パーサ
- task front matter (TOML) の読み書き
- edges TOML の読み書き
- config TOML の読み書き

受け入れ条件:
- 読んで書き戻しても内容が保たれる（安定フォーマット）

## GS-04: ID生成（ULID）
- 新規作成時にULID生成
- 短縮ID表示ヘルパ

受け入れ条件:
- `add` がIDを生成して保存できる準備が整う

## GS-05: add（非対話）
- `--title` 必須で作成
- kind/status/parent 検証
- 原子的更新

受け入れ条件:
- `shelf add --title ...` で `.shelf/tasks/<id>.md` ができる

## GS-06: add（対話）
- TTY判定
- Title入力
- Kind選択（j/k）
- Parent選択（検索/0=root）

受け入れ条件:
- 対話でタスクが作れる

## GS-07: ls / show / tree
- `ls` フィルタと安定ソート
- `show` 表示
- `tree` 構築（parent参照、循環検出）

受け入れ条件:
- 週→曜日→タスクが表示できる

## GS-08: set / mv
- set: title/kind/status/body更新
- mv: parent付け替え（循環拒否）

受け入れ条件:
- `mv` でツリー構造を変更できる

## GS-09: link / unlink / links
- edges分離で追加/削除
- `depends_on` の向き表示を必須化
- inbound逆引き

受け入れ条件:
- 子同士の依存が表現・確認できる

## GS-10: doctor（整合性チェック）
- task/edge/configの検証
- 壊れた箇所の列挙

受け入れ条件:
- 意図的に壊したデータを検出できる

## GS-11: ドキュメント仕上げ
- README に基本例
- `docs/` の内容と実装の一致チェック

受け入れ条件:
- 仕様通りに使える
