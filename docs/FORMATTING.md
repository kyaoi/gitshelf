# FORMATTING（安定フォーマット規約）

## task front matter
- TOML front matter のキー順序を固定する（推奨）
  1) id
  2) title
  3) kind
  4) status
  5) parent（あれば）
  6) created_at
  7) updated_at
- 日時は RFC3339（例: 2026-03-05T12:34:56+09:00）

## edges
- `[[edge]]` は `to asc`, `type asc` でソートして書く
- ファイル末尾に改行

## CLI出力
- IDは短縮表示を基本（例: `[01JABCDE]`）
- depends_onは必ず矢印で表示する
