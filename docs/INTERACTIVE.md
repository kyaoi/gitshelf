# INTERACTIVE (TTY Selection Spec)

Interactive mode is available only when stdin/stdout are TTY.
If not TTY, users must provide required flags.

## Supported Commands

- `shelf add` (when `--title` is omitted)
- `shelf link` (when `--from/--to/--type` are omitted)
- `shelf unlink` (when `--from/--to/--type` are omitted)
- `shelf show` (when `<id>` is omitted)
- `shelf explain` (when `<id>` is omitted)
- `shelf edit` (when `<id>` is omitted)
- `shelf set` (when `<id>` is omitted)
- `shelf snooze` (when `<id>` is omitted or `--by/--to` is omitted)
- `shelf mv` (when `<id>` and/or `--parent` is omitted)
- `shelf done` (when `<id>` is omitted; `status!=done` tasks are prioritized)
- `shelf links` (when `<id>` is omitted)
- `shelf triage` (without `--auto`)
- `shelf board` (TTY only, Daily Cockpit `board` mode)
- `shelf calendar` (TTY only, Daily Cockpit `calendar` mode unless `--json`)
- `shelf tree` (TTY only, Daily Cockpit `tree` mode unless `--plain` / `--json`)
- `shelf cockpit` (TTY only, explicit Daily Cockpit entry point)

## Key Bindings

- `j` / `k`: move selection down/up
- `Enter`: confirm
- `/`: search mode
- `?`: toggle help overlay
- `Esc`:
  - in search mode: clear search and leave search mode
  - otherwise: cancel selection
- `q`: cancel selection (outside search mode)
- `Ctrl+C`: cancel selection
- Arrow up/down are also supported.

## Search

- `/` enters incremental search.
- Search matches option label/search text (task title and short/full ID).
- Result list updates as user types.

## Display

Task candidate line format:

- default: `{tree-prefix}{title}` (IDs hidden)
- with `--show-id`: `[{short}] {tree-prefix}{title}`

- Default: task selectors hide IDs and prefer hierarchical labels.
- `--show-id` / `-i`: include short IDs in selector labels.
- Task selectors always show selected task body preview (`(empty body)` when empty).
- Enum selectors intentionally render without body preview.
- Selected row, prompt, help line, and preview header are colorized on TTY.
- `NO_COLOR=1` disables colors (`CLICOLOR_FORCE=1` overrides non-TTY detection).

## add Interactive Flow

1. Input `Title` (required)
2. Select `Kind`
3. Select `Status`
4. Review screen (`Title`/`Kind`/`Status`/`Tags`/`Due`/`Repeat`/`Parent`)
5. Select `Create task` or `Cancel`

Tags selector supports:

- toggle existing config tags
- add a new freeform tag
- clear selected tags

Parent candidates are rendered as tree labels (without IDs by default), for example:

`(root)`
`週目標`
`├─ 月曜日`
`│  └─ 英単語100個`

## link Interactive Flow

1. Select source task
2. Select destination task
3. Select link type

The type selection screen includes this warning:

`A depends_on B = B must be done before A`

## unlink Interactive Flow

1. Select source task
2. Select existing outbound edge to remove

## show / explain / edit / set / done / links / snooze Interactive Flow

1. Select target task by ID/title
   - Uses hierarchical tree-style labels without IDs by default
2. (`set` only, when no update flags are passed) choose fields in a menu and edit interactively (`Title`/`Kind`/`Status`/`Tags`/`Due`/`Repeat`/`Parent`/`Body replace`/`Body append`)
3. (`snooze` only, when `--by` and `--to` are both omitted) choose a preset like `Today` / `Tomorrow` / `By +3 days`, or choose `Custom by days` / `Custom date token`
4. `set` shows change preview before apply

## mv Interactive Flow

1. Select target task by ID/title (when `<id>` omitted)
2. Select new parent (`0: [root]`) when `--parent` omitted

## triage Interactive Flow

1. Load triage targets by `--kind` + `--status` (default `inbox/open`)
2. For each task choose one action:
   - `Edit fields` (same editor as `set` interactive)
   - `Set status ...`
   - `Archive task`
   - `Skip` / `Quit triage`

## board TUI

- `shelf board` now opens the shared Daily Cockpit in `board` mode
- status columns follow config order
- use `C/T/B/R/N` to switch modes without leaving the shell

## calendar TUI

- used by `shelf cockpit`, `shelf calendar`, `shelf tree`, `shelf board`, and also by `shelf review` / `shelf now` on TTY unless `--plain` or `--json` is specified
- layout is `main + right sidebar`
- `Ctrl+H` / `Ctrl+L`: switch to the previous / next mode
- `C/T/B/R/N`: switch modes
- `t`: jump the calendar focus to today
- `Tab` / `Shift+Tab`: move between panes
- `h` / `l`: move by one day in calendar mode, move the sidebar calendar by one day when the right pane is focused, or switch review/now tabs / board columns otherwise
- `j` / `k`: move by one week in calendar mode, move the sidebar calendar by one week when the right pane is focused, or move rows in tree/board/review/now
- `[` / `]`: move by one month inside the current range
- `g` / `G`: jump to first / last day in range, or first / last row in the sections pane
- in `calendar` mode, the month grid is larger and the focused-day task list lives above the inspector
- in non-calendar modes, the right sidebar shows a compact calendar above the inspector; focus it with `Tab` to move the date directly
- `n` / `p`: switch focused-day tasks in calendar mode, switch cockpit tabs in review/now, or move board columns
- `now` mode shows `Focused Day`, `Overdue`, and `Today` side by side in the main pane
- the header and mode tabs stay fixed at the top
- `PgUp` / `PgDn` or `Ctrl+U` / `Ctrl+D`: scroll the body
- `Home` / `End`: jump to the top or bottom of the body
- `1..6`: jump directly to a visible section
- `v`: toggle a single multi-select mark in tree/board modes
- `u`: clear all marks in tree/board modes
- `V`: start/finish continuous range selection in tree/board modes; move to expand the marked range without clearing previously marked tasks
- `Ctrl+[` leaves transient modes and returns to the normal cockpit state without quitting
- `m`: in tree mode, move the current task or marked tasks under the currently highlighted task; move mode also exposes `(root)` as a target
- `a`: open inline add composer for the focused day
- `o` / `i` / `b` / `d` / `c`: set the selected task, or all marked tree/board tasks, to `open` / `in_progress` / `blocked` / `done` / `cancelled`
- `Enter`: toggle compact/detailed inspector
- `e`: open the selected task in the configured editor
- `z`: open snooze presets for the selected task
- `r`: reload task data
- `q`: close help first, otherwise quit
- `Esc` / `Ctrl+C`: quit
- moving beyond the current window shifts the calendar range automatically
