# IMPLEMENTATION_PLAN（日本語版）

この文書は初期リリース時の実装計画の記録です。
現在の実装状態そのものではなく、仕様をどの順番で組み立てたかを示します。

## GS-01: プロジェクト雛形

- Go module 初期化
- `cmd/shelf` 追加
- `internal/` 骨格

## GS-02: init

- `.shelf/` 作成
- デフォルト `config.toml` 作成

## GS-03: データモデルとパーサ

- task front matter の読み書き
- edge TOML の読み書き
- config TOML の読み書き

## GS-04: ID 生成

- ULID 生成
- short ID helper

## GS-05: add（非対話）

- `--title` 必須
- kind / status / parent 検証
- atomic write

## GS-06: add（対話）

- TTY 判定
- title 入力
- kind 選択
- parent 選択

## GS-07: ls / show / tree

- `ls` filter
- `show`
- `tree`

## GS-08: set / mv

- metadata 更新
- parent 付け替え

## GS-09: link / unlink / links

- edge 追加 / 削除
- inbound 逆引き

## GS-10: doctor

- task / edge / config 検証

## GS-11: ドキュメント仕上げ

- README
- `docs/` と実装の整合
