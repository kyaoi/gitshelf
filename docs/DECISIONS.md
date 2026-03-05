# DECISIONS

## 保存形式（確定）
- タスク: `.shelf/tasks/<id>.md`（フラット）
- 親子: タスクfront matterの `parent` で表現（無限ネスト可）
- リンク: `.shelf/edges/<src_id>.toml`（outbound edges）
- Kind/State: 分離する

## UI（確定）
- TUIは作らない
- ただし `shelf add` や `shelf link` はターミナルで対話選択を提供（j/k、検索、ページング）

## ID（確定）
- ULID（推奨）
- 表示用は短縮（先頭8〜10文字）
