# gitshelf 初期リリース仕様パック（CLI / Go）

このZipは、`gitshelf`（コマンド: `shelf`）を **“最初のリリースで完結”** させるための仕様・運用・AI Agents向けプロンプトをまとめたものです。

- 言語: Go
- UI: **TUIは作らない**（ただしターミナル上の対話選択は行う）
- 保存: **Git管理前提**（差分が読みやすく、衝突しにくいフォーマット）
- モデル: **親子=ツリー** / **関連=リンク（別管理）** / **KindとStateを分離**

## まず読む順番
1. `docs/SPEC.md`（全体仕様）
2. `docs/COMMANDS.md`（CLI仕様）
3. `docs/STORAGE.md`（保存形式と不変条件）
4. `docs/INTERACTIVE.md`（j/k選択・検索などの対話仕様）
5. `docs/IMPLEMENTATION_PLAN.md`（実装タスク分割）
6. `AGENTS.md` と `prompts/`（AI Agents運用）

## ゴール（短い定義）
- `.shelf/` 配下にタスクとリンクを保存する
- タスクは **無限に入れ子**（`parent` によるツリー）
- 子同士（兄弟/別枝）の関係は **edges** で表現
- `shelf add / ls / tree / show / set / mv / link / unlink / links / init` が揃っており、日常運用に不足がない

> 注意: このパックは「仕様と運用の確定版」です。コード生成はAI Agents側で行う前提です。
