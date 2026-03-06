# TESTING（日本語版）

## 重要テスト

- `init` の冪等性
- `add` で task 作成と front matter 検証
- `mv` で parent 更新と循環拒否
- `link` で edge 作成と重複抑止
- `unlink` で edge 削除
- `links` で inbound 逆引き
- `doctor` で壊れたデータ検出

## 安定差分

- `edges` の順序固定
- `tree` の順序固定

## 競合に強いこと

- task 本文編集と link 操作が別ファイルに分離されること

## 推奨コマンド

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```
