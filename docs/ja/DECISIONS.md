# DECISIONS（日本語版）

## 保存形式

- task: `.shelf/tasks/<id>.md`
- 親子: task front matter の `parent`
- link: `.shelf/edges/<src_id>.toml`
- kind と status は独立

## UI

- 日常編集は `Cockpit` に集約する
- `calendar/tree/board/review/now` は `Cockpit` の起動プリセット
- read-only query は `ls` / `next`
- script からの link 操作は `link` / `unlink` / `links`

## ID

- ULID を使う
- 表示用は short ID
- 既定では ID を隠し、`--show-id` で表示する

## Lock

- 変更系操作は `.shelf/.write.lock` で排他する
- lock 取得失敗時は timeout error を返す
