# IMPLEMENTATION_PLAN（日本語版）

現在の Cockpit-first 実装に至るまでの大まかな実装段階です。

## Phase 1: bootstrap

- module / entrypoint
- `init`
- 安定した `.shelf/` layout

## Phase 2: storage と不変条件

- task / edge / config parser
- ULID 生成
- 安定フォーマットと atomic write

## Phase 3: Cockpit workspace

- calendar / tree / board / review / now
- TUI 内 task 編集
- sidebar 同期と popup editor

## Phase 4: read-only / scripting command

- `ls`
- `next`
- `link` / `unlink` / `links`

## Phase 5: docs / verification

- docs と実装の一致
- format / test / race / vet を green に保つ
