# Development Plan

## Source Tasks

Reference parent task: `01KK45TT3QW82TTXGT80CTCKHH`

Child tasks discovered under `~/DailyTodo/.shelf/tasks/`:

1. `01KK45XPG05ZRJCMXE9929815N` `pathを略記法で賭けるようにする`
2. `01KK45YB8M7ZHPPJ0GDVQAY1QB` `.shelf配下に作られないようにする`
3. `01KK460FAM7VACN2AZ5RNR7MZ8` `編集後にCommit/Pushできるようにしたい`
4. `01KK467WXJZMTH1Y12JAZ0HKZC` `tagの概念の復活`
5. `01KK6ZSZE3HGEZ4VPBJPYD09PN` `Sundayから始まるようにする`
6. `01KK84B845TEP3W306WX0MBNHR` `Kindを指定できるようにするorTUIから編集できるようにする`
7. `01KK88S2QMZBJXBXX3NGRECDRV` `親に期限がついているときにはその子についても表示するようにする`
8. `01KK88T87QWB6DYWAACB5CF28C` `タイトルをコピーできるようにする`
9. `01KK8CKV8NJASBZQC4F3G66Z4W` `edgeファイルの必要性について調べる（もし機能が消えていたらタスクごとの関係性をつけられるようにする）`

## Current Refinement Wave

This follow-up plan captures the next Cockpit/TUI refinement pass requested after the first implementation wave.

### Layout invariants

- Keep panel heights fixed regardless of item count.
- When content exceeds the available height, keep the box size stable and reveal overflow through scrolling/selection-following windows.
- Avoid border corruption by using ANSI-aware width trimming/padding for colored strings.

### Popup behavior

- Use large centered popups for Add / Link / Tag / Kind / Filter / Snooze / Help.
- Keep popup headers pinned and scroll only the selectable body content.
- Show the currently selected task in Link and Tag popups.

### Add workflow

- `Tab` / `Shift+Tab` cycle between title and kind.
- `Enter` confirms creation.
- `Esc` / `Ctrl+[` cancel add mode.
- `q` must be accepted as normal title input.
- `j/k` are the only kind cycling keys inside the add popup.
- `a` creates a child when a task is selected, otherwise it falls back to root creation.
- `A` always creates at root.

### Link and Tag input modes

- Link enters query input mode with `/`.
- While typing a link query, navigation keys are treated as text and must not trigger picker actions.
- Tag enters text input mode from `+ Add new tag`.
- While typing a new tag, navigation keys are treated as text and the in-progress input is visible in the same popup.

### Sidebar / Selected Day

- Rename `Focused Day` to `Selected Day`.
- Calendar mode no longer needs pane focus switching.
- In non-calendar modes, `n/p` moves through `Selected Day` tasks and the main view follows the same task.
- Keep the right sidebar split at `60% calendar area / 38% inspector / 2% gap`.
- Split the calendar area into a mini calendar and `Selected Day`.
- Synchronize main selection with the sidebar calendar and `Selected Day`.

### Review / tabs

- Keep review tabs visible from first render.
- If width is insufficient, use a sliding tab window that always keeps the active tab visible.

## Current State Summary

- The project is a Go CLI/TUI task manager centered on `shelf cockpit`.
- Storage already supports `kind`, `tags`, `parent`, and `edges`.
- TUI currently supports:
  - add with context-derived defaults
  - status changes
  - archive toggle
  - move in tree mode
  - opening the task file or edge file in an external editor
- Missing or incomplete areas relative to the task list:
  - no clipboard/copy action
  - no in-TUI kind editing
  - no in-TUI tag editing
  - calendar week starts on Monday
  - no post-edit git automation
  - no first-class edge management UI beyond opening the edge file in an editor

## Inferred Task Meanings

These are the current working interpretations and need user confirmation before implementation:

1. Path shorthand
   - Confirmed: allow global `default_root` to be written with shorthand such as `~/DailyTodo`.

2. Do not create inside `.shelf`
   - Confirmed direction: avoid nested `.shelf/.shelf` creation and reject invalid roots inside `.shelf` internals.

3. Commit/push after edit
   - Confirmed direction: collect edits during one `shelf` session and commit or push after the session exits.
   - Config should control the default behavior, and flags should be able to override it.

4. Bring tags back
   - Confirmed direction: add practical TUI editing for tags and document it.

5. Start on Sunday
   - Confirmed direction: change the week anchor to Sunday.

6. Kind can be specified or edited in TUI
   - Confirmed direction: both create-time kind selection and in-TUI kind editing.

7. Show children when parent has a due date
   - Confirmed direction: calendar/review/now should surface all descendants recursively when a parent is due, even if the descendant has no own `due_on`.

8. Copy title
   - Confirmed first scope: copy selected task titles in the TUI.
   - Multi-select should copy multiple titles joined by a configurable separator.
   - Deferred follow-up tasks to preserve scope:
     - copy task body
     - copy task path / file path
     - copy subtree / tree selection payloads

9. Edge file necessity / task relationships
   - Confirmed direction: keep edge storage, add both standalone CLI commands and TUI operations, and document them.
   - Additional follow-up accepted during implementation:
     - use tree/path labels so the target task is identifiable
     - hide IDs by default and show them only when `--show-id` is enabled
     - make `link_types` configurable in config
     - make the blocking relation configurable instead of hard-coding `depends_on`
     - reject blocking links from a child task to one of its ancestors

## Proposed Delivery Order

1. Confirm ambiguous scope with the user.
2. Harden shelf root handling:
   - reject `.shelf`-internal roots
   - document the behavior
3. Calendar behavior:
   - switch default week start to Sunday
   - document the behavior
4. Metadata editing:
   - add kind selection/edit flow
   - add tag editing flow
   - update docs for keybindings and behavior
5. Task visibility:
   - define and implement descendant visibility when a parent is due
   - update docs
6. Copy workflow:
   - add clipboard/copy action once the exact payload is confirmed
7. Git workflow:
   - add configurable post-edit commit/push behavior
   - update config/docs
8. Relationship workflow:
   - decide whether docs-only clarification is enough or whether TUI edge editing is required
   - implement and document accordingly
9. Path shorthand:
   - implement once the exact meaning is confirmed because it changes either display or selection semantics

## Commit Strategy

- Keep one commit per confirmed task or tightly coupled subtask.
- Update docs in the same commit as the behavior change.
- Run:
  - `gofmt -w`
  - `go test ./...`
  - `go test -race ./...` when practical
  - `go vet ./...`
