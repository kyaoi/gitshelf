# FORMATTING（日本語版）

## task front matter

- TOML front matter のキー順序は固定する
  1. `id`
  2. `title`
  3. `kind`
  4. `status`
  5. `tags`
  6. `due_on`
  7. `repeat_every`
  8. `archived_at`
  9. `parent`
  10. `created_at`
  11. `updated_at`
- 日時は RFC3339

## edges

- `[[edge]]` は `to asc`, `type asc` で安定ソート
- ファイル末尾には改行を入れる

## CLI 出力

- 既定では ID 非表示
- `--show-id` / `-i` で short ID を表示
- `depends_on` は常に矢印つきで表示する
