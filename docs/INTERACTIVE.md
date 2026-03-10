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
- `Tab` / `Shift+Tab`: move focus between panes in non-calendar modes
- `?`: toggle help overlay
- `q`: close help first, otherwise quit
- `Esc`: quit or leave transient state
- `Ctrl+[` : return to normal state from transient overlays

## Calendar Mode

Weeks are rendered Sunday through Saturday.

If a parent task has a due date, descendants without their own `due_on` are shown on that same date as inherited entries.

Main keys:

- `t`: jump to today
- `h` / `l`: move by day
- `j` / `k`: move by week
- `[` / `]`: move by month
- `n` / `p`: cycle tasks on the selected day
- `a`: create as a child of the selected task, or at root when nothing is selected
- `A`: create at root

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

- `K`: edit kind on the selected task
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

Link selectors use tree-style labels and a scrolling window so duplicate titles remain identifiable.
IDs are hidden there unless `--show-id` is enabled.
Transient pickers and composers are shown as centered popups.

## Add

- `a`: create using the current mode context as a child of the selected task
- when nothing is selected, `a` also creates at root
- `A`: create using the current mode context at root
- calendar / review / now keep using the focused day as the due date default
- board keeps using the selected column status as the status default
- the add composer now includes a title field and kind field
- `Tab` / `Shift+Tab` cycle between title and kind
- `j` / `k` cycle kinds while the kind field is active
- `Enter` confirms creation
- `Esc` / `Ctrl+[` cancel add mode
- `q` is normal text input inside the title field

## Filters

- `f`: open a popup filter editor
- include / exclude filters are available for both `status` and `kind`
- the applied filters affect every Cockpit mode

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
- `q` and `Esc` cancel plain selectors
- Link uses `/` to enter query input mode; while typing, movement keys are treated as text
- Tag enters text input mode from `+ Add new tag`; while typing, movement keys are treated as text
- `Selected Day` replaces the old focused-day panel name and stays synced with the main selection
