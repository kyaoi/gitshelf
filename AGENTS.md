# AI Agents 運用ガイド（gitshelf）

このリポジトリをAI Agentsで実装するためのルール集です。

## ブランチ運用
- 原則 `develop` ブランチで進める（ブランチを切らない）
- 1タスク=できれば1コミット
- コミットメッセージに **タスクID** を含める（例: `GS-03: implement shelf add interactive`）

## 品質ゲート（最低限）
- `go test ./...`
- `go test -race ./...`（環境が許すなら）
- `go vet ./...`
- `gofmt -w`（自動整形）

## 仕様の優先順位
1. `docs/SPEC.md`
2. `docs/STORAGE.md`（不変条件、ソート、原子的更新）
3. `docs/COMMANDS.md`（CLI I/O）
4. `docs/INTERACTIVE.md`（対話UI）

仕様に無い挙動は「実装都合で勝手に増やさない」。必要なら `docs/DECISIONS.md` を更新して決定を記録する。

## 生成AI向けの作業開始プロンプト方針
- **短く**、タスクに直接関係することだけを書く
- 参照すべきファイルを明示して迷子にならないようにする
- 変更対象のパスを具体化する

`prompts/agent_start.md` と `prompts/tasks/*.md` を使う。
