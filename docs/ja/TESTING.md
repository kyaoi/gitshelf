# TESTING（日本語版）

## 重要テスト

- `init` の冪等性
- `link` / `unlink` の edge 更新と重複抑止
- `links` の outbound / inbound 逆引き
- Cockpit の mode 切り替え時に selection が可能な限り維持されること
- sidebar の `Calendar` と `Selected Day` が main pane と同期すること
- 継承 due が calendar / tree / review / now で一貫表示されること

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
