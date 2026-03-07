# DECISIONS

## 保存形式（確定）
- タスク: `.shelf/tasks/<id>.md`（フラット）
- 親子: タスクfront matterの `parent` で表現（無限ネスト可）
- リンク: `.shelf/edges/<src_id>.toml`（outbound edges）
- Kind/Status: 分離する

## UI（確定）
- 基本は通常CLI
- ただし `shelf add` や `shelf link` はターミナルで対話選択を提供（j/k、検索、ページング）
- 追加拡張として `shelf board` は TUI を許可する

## ID（確定）
- ULID（推奨）
- 表示用は短縮（先頭8〜10文字）
- デフォルト表示は ID 非表示（`--show-id` / `-i` で表示）

## Lock（確定）
- mutating command は `.shelf/.write.lock` で排他
- lock 取得に失敗した場合はタイムアウトエラーを返す
