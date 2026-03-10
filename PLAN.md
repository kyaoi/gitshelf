# Development Plan

## Source Tasks

Reference parent task: `01KK45TT3QW82TTXGT80CTCKHH`

Open or partially completed child tasks discovered under `~/DailyTodo/.shelf/tasks/`:

1. `01KK45YB8M7ZHPPJ0GDVQAY1QB` `.shelf配下に作られないようにする`
2. `01KK88T87QWB6DYWAACB5CF28C` `タイトルをコピーできるようにする`
3. `01KKB4VMHVSC98PRD7TTMP2Q53` `入力を矢印キーで戻れるようにする（通常の文字入力と同じように）`
4. `01KKB8XBGDQMMBZSXP350V60DX` `JAのDocsのリンクがおかしいので修正する`

## Current State Summary

- The project is a Go CLI/TUI task manager centered on `shelf cockpit`.
- Storage already supports `kind`, `tags`, `parent`, and `edges`.
- TUI currently supports:
  - add with context-derived defaults
  - status changes
  - archive toggle
  - move in tree mode
  - kind editing
  - tag editing
  - link / unlink actions
  - title copy to the clipboard for the selected task or marked tasks
  - opening the task file or edge file in an external editor
- Remaining gaps:
  - task and edge storage paths are still hard-coded under `.shelf/`
  - post-exit git automation still assumes all writable data lives under `.shelf/`
  - copy currently covers titles only, not path / body / subtree payloads
  - text entry currently supports append and backspace but not cursor-style left/right editing
  - `docs/ja/README.md` contains broken relative links

## Task Analysis

These notes are based on the current task files under `~/DailyTodo/.shelf/tasks/` plus the repository state:

1. Do not create inside `.shelf`
   - The actual task body is broader than root normalization.
   - Goal: keep `.shelf/config.toml`, but allow task data to live outside `.shelf/` with `tasks` and `edges` kept together at the same configured path.
   - Default remains the current `.shelf/` layout.
   - Current implementation only covers root normalization / rejection of invalid `.shelf`-internal roots; it does not yet support configurable task and edge storage paths.

2. Copy title
   - The task body marks the first milestone as already done: copy selected task titles only, with multi-select support.
   - Current repo state matches that first milestone.
   - Remaining follow-ups from the task body:
     - copy task path / file path
     - split commands between parent-only copy and subtree / tree-structured copy
     - optionally copy task body, ideally with a configurable format

3. Arrow-key editing in text input
   - The task file currently has no body text beyond the title.
   - Observed gap in the repo: add title input, tag input, link query input, and the generic text prompt all append text and handle backspace, but do not support cursor-style left/right movement or mid-string editing.
   - Treat this as an open ergonomics task and avoid claiming a narrower final scope until the task body is clarified.

4. Broken links in Japanese docs
   - The task file currently has no body text beyond the title.
   - Observed issue in the repo: broken relative links are currently concentrated in `docs/ja/README.md`.
   - Other cross-links checked under `docs/ja/*.md` resolve correctly.
   - Treat this as a docs fix task, with the currently confirmed scope limited to the Japanese README links.

## Proposed Delivery Order

1. Docs integrity:
   - fix broken relative links in `docs/ja/README.md`
   - keep the Japanese docs tree internally consistent
2. Text input ergonomics:
   - add cursor-style editing for TUI text-entry surfaces
   - document the final key behavior once implemented
3. Storage layout redesign:
   - keep `.shelf/config.toml`
   - make the shared storage location for `tasks` and `edges` configurable
   - update git-on-exit, undo/snapshot handling, and storage docs to follow the new layout
4. Copy workflow follow-ups:
   - add task path / file path copy
   - decide how parent-only copy and subtree / tree copy are exposed
   - add optional body copy formatting if it is still desired

## Commit Strategy

- Keep one commit per confirmed task or tightly coupled subtask.
- Update docs in the same commit as the behavior change.
- Run:
  - `gofmt -w`
  - `go test ./...`
  - `go test -race ./...` when practical
  - `go vet ./...`
