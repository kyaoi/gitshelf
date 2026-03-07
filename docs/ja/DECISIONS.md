# DECISIONS（日本語版）

## 保存形式（確定）

- タスク: `.shelf/tasks/<id>.md`（フラット）
- 親子: task front matter の `parent` で表現（無限ネスト可）
- リンク: `.shelf/edges/<src_id>.toml`（outbound edges）
- Kind / Status は分離

## UI（確定）

- 基本は通常 CLI
- `shelf add` や `shelf link` は対話選択を提供
- `shelf board` は例外的に TUI を許可

## ID（確定）

- ULID を使用
- 表示用は short ID
- 既定表示では ID 非表示（`--show-id` / `-i` で表示）

## Lock（確定）

- mutating command は `.shelf/.write.lock` で排他
- lock 取得失敗時は timeout error を返す
