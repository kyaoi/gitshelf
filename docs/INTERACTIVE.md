# INTERACTIVE

Current interactive behavior for the Cockpit-first `shelf` tool.

## Main Rule

Interactive daily work happens inside `Cockpit`.

You normally enter it through one of these commands:

- `shelf`
- `shelf cockpit`
- `shelf calendar`
- `shelf tree`
- `shelf board`
- `shelf review`
- `shelf now`

All of those open the same TUI workspace with different starting modes.

## Shared Cockpit Navigation

- `C`: calendar mode
- `T`: tree mode
- `B`: board mode
- `R`: review mode
- `N`: now mode
- `Ctrl+H` / `Ctrl+L`: previous / next mode
- `Tab` / `Shift+Tab`: move focus between panes
- `?`: toggle help overlay
- `q`: close help first, otherwise quit
- `Esc` / `Ctrl+C`: quit or leave transient state
- `Ctrl+[` : return to normal state from transient overlays

## Calendar Mode

Weeks are rendered Sunday through Saturday.

If a parent task has a due date, descendants without their own `due_on` are shown on that same date as inherited entries.

Main keys:

- `t`: jump to today
- `h` / `l`: move by day
- `j` / `k`: move by week
- `[` / `]`: move by month
- `n` / `p`: cycle tasks on the focused day
- `a`: create from the current context
- `A`: quick capture

## Tree Mode

Main keys:

- `h`: collapse current subtree, or move to parent
- `l`: expand current subtree
- `m`: move selected task or marked tasks
- `v`: toggle mark on the current task
- `V`: start or stop range marking
- `u`: clear all marks

## Board Mode

Main keys:

- `h` / `l`: move between columns
- `j` / `k`: move within a column
- `v`: toggle mark on the current task
- `V`: start or stop range marking
- `u`: clear all marks

## Review / Now Modes

These are compact operational views inside the same workspace.

- `review`: inbox / overdue / blocked / ready scan
- `now`: focused execution view for today

## Common Task Actions

These actions operate on the selected task, or on marked tasks when multi-select is active.

- `K`: edit kind, or choose the kind while the add composer is open
- `#`: edit tags for the selected task
- `y`: copy the selected title, or marked titles joined by the configured separator
- `o`: set `open`
- `i`: set `in_progress`
- `b`: set `blocked`
- `d`: set `done`
- `c`: set `cancelled`
- `x`: archive toggle
- `z`: snooze presets
- `e`: open the task file in `$VISUAL`, `$EDITOR`, then `vi`
- `L`: add a link from the selected task
- `U`: remove one outbound link from the selected task
- `Enter`: toggle compact / detailed inspector
- `r`: reload

Link selectors use tree/path labels so duplicate titles remain identifiable.
IDs are hidden there unless `--show-id` is enabled.

## Add / Capture

- `a`: create using the current mode context
  - calendar / review / now: use focused day as the due date default
  - tree: use selected task as the parent default
  - board: use selected column status as the status default
  - `K` inside the composer opens a kind picker before creation
- `A`: quick capture (`kind=inbox`, `status=open`)

## Scrolling

- fixed header stays on screen
- body scroll:
  - `PgUp` / `PgDn`
  - `Ctrl+U` / `Ctrl+D`
  - `Home` / `End`

## Selector Behavior

Long task selectors scroll automatically.

- tree-style labels are used where hierarchy matters
- `(root)` appears as an explicit move target where relevant
- `q`, `Esc`, and `Ctrl+C` cancel plain selectors
