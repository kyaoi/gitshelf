# COMMANDS（CLI仕様）

## 共通
- ワークディレクトリの上位に `.shelf/` が無ければエラー（ただし `init` は例外）
- 出力は人間向け（必要に応じて `--json` を追加してもよいが必須ではない）
- 失敗時は非ゼロ終了コード

## shelf init
- `.shelf/` と必要ファイルを作る
- 既に存在する場合は安全にスキップ（上書きしない）

### 作成物
- `.shelf/config.toml`（デフォルト）
- `.shelf/tasks/`
- `.shelf/edges/`

## shelf add
- タスクを追加する
- TTYの場合: 対話（Title → Kind → Parent）
- 非TTYの場合: `--title` 必須（対話不可）

### flags（例）
- `--title <str>`
- `--kind <kind>`
- `--state <state>`（通常は default）
- `--parent <id|root>`
- `--body <str>`（任意）

### 出力（例）
- `Created: [<short>] <title>`

## shelf ls
- フラット一覧
- フィルタ可能

### flags（例）
- `--kind <kind>`
- `--state <state>`
- `--parent <id|root>`
- `--limit <n>`（デフォルト50）
- `--search <query>`（title/bodyの部分一致）

## shelf tree
- 親子をツリー表示

### flags
- `--from <id|root>`
- `--max-depth <n>`（省略時は無制限）
- `--state <state>`（doneを隠す等）

## shelf show <id>
- タスク詳細（front matter + body）
- 併せて outbound/inbound links のサマリも表示して良い

## shelf set <id>
- `kind` / `state` / `title` / `body` の更新

### flags
- `--title <str>`
- `--kind <kind>`
- `--state <state>`
- `--body <str>`（置換）
- `--append-body <str>`（追記）

## shelf mv <id>
- 親の付け替え

### flags
- `--parent <id|root>`

## shelf link
- リンク追加（outbound）
- TTY: source → dest → type の対話（j/k + /search）
- 非TTY: `--from --to --type` 必須

### flags
- `--from <id>`
- `--to <id>`
- `--type <link_type>`

### 出力（例）
- `Linked: [src] --depends_on--> [dst]`

## shelf unlink
- リンク削除（outbound）
- `--from --to --type` 必須

## shelf links <id>
- 指定タスクの links を表示
  - outbound: `.shelf/edges/<id>.toml`
  - inbound: 全edge走査で逆引き（`to == id`）

表示は `type` ごとにグルーピングしても良い。

## shelf doctor
- `.shelf/` の整合性チェック（不変条件検証）
- 壊れているファイルを列挙し、修正ヒントを出す
