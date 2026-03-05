# INTERACTIVE（対話選択仕様）

対象: `shelf add`, `shelf link`

## 基本操作（MUST）
- `j` / `k` で上下移動
- `Enter` で決定
- `Esc` または `Ctrl+C` でキャンセル（非ゼロ終了）

## 検索（MUST）
- `/` で検索入力モード
- 入力中はインクリメンタルに候補を絞る（title、短縮ID）
- `Enter` で検索確定、`Esc` で検索解除

## ページング（SHOULD）
- 候補が多い場合、表示をページングする
- `Ctrl+f` / `Ctrl+b` でページ送り/戻し（または `PgDn/PgUp`）

## 表示フォーマット（MUST）
- 候補行: `[{short}] {title}  ({kind}/{state})`
- 親がある場合は、右端やサブ表示で `parent` を示しても良い

## add の対話順
1. Title（1行入力）
2. Kind（選択）
3. Parent（任意 / `0` で root）

## link の対話順
1. source（選択）
2. destination（選択）
3. type（選択）

## 依存の向きの注意表示（MUST）
- `depends_on` を選ぶ画面に説明を入れる:
  - `A depends_on B = AをやるにはBが先`
- 作成後に必ず矢印付きで表示:
  - `Linked: [A] --depends_on--> [B]`
