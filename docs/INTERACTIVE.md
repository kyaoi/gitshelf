# INTERACTIVE

Current interactive behavior for the Cockpit-first `shelf` tool.

This is the single detailed keybinding reference for Cockpit.

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
- `Esc`: leave transient state, or close help if it is open
- `Ctrl+[` : leave popup or input mode and return to normal state

Transient pickers and composers are shown as centered popups.
Scrollable lists keep a fixed box height; overflow is handled by scrolling instead of resizing the layout.

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

The main calendar view does not use pane focus switching.

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

- `K`: edit kind on the selected task, or on marked tasks when multi-select is active
- `#`: edit tags for the selected task
- `y`: copy the selected title, or marked titles joined by the configured separator
- `Y`: copy the selected task subtree, or marked subtrees, as an indented title tree
- `P`: copy the selected task file path, or marked file paths, as absolute paths
- `O`: copy the selected task body, or marked task bodies
- `M`: open advanced copy presets with preview and save-command help
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

Inside advanced copy (`M`):

- `j` / `k`: choose `Custom` or one saved preset
- `Tab` / `Shift+Tab`: move focus between preset list and custom fields
- custom `template` / `join_with` fields use escaped text such as `\n`
- `subtree style` switches `{{subtree}}` between plain indentation and ASCII tree rendering
- `Enter`: copy the currently previewed payload
- `Ctrl+S`: save the current preset into `.shelf/config.toml`
- the generated `shelf config copy-preset set ...` command is shown in the popup footer
- `Esc` / `q`: close the popup

Link selectors use tree-style labels and a scrolling window so duplicate titles remain identifiable.
IDs are hidden there unless `--show-id` is enabled.
- In link pickers, `h` / `l` collapse and expand the hierarchy like Tree mode.
- Link type cycling uses `Tab` / `Shift+Tab`.

## Add

- `a`: create using the current mode context as a child of the selected task
- when nothing is selected, `a` also creates at root
- `A`: create using the current mode context at root
- calendar / review / now keep using the focused day as the due date default
- board keeps using the selected column status as the status default
- the add composer now includes a title field and kind field
- `Tab` / `Shift+Tab` cycle between title and kind
- `Left` / `Right` move the cursor inside the title field
- `j` / `k` cycle kinds while the kind field is active
- `Enter` confirms creation
- `Esc` / `Ctrl+[` cancel add mode
- `q` is normal text input inside the title field

## Filters

- `f`: open a popup filter editor
- include / exclude filters are available for both `status` and `kind`
- the applied filters affect every Cockpit mode

## Tags

- `Space`: toggle the highlighted tag
- `Enter` on `Done`: save and close
- `Enter` on `+ Add new tag`: enter text input mode
- `Ctrl+S`: save and close from anywhere in tag editing
- while typing a new tag, `Left` / `Right` move the cursor and typed text is inserted at the cursor position

## Non-Calendar Sidebar

- the right pane is split into `Calendar / Selected Day / Inspector`
- the height ratio is `Calendar 40% / gap 1% / Selected Day 28% / gap 1% / Inspector 30%`
- main selection syncs the sidebar date and `Selected Day`
- moving the sidebar calendar updates the main selection when that day has visible tasks
- `n` / `p` inside `Selected Day` updates the main selection in non-calendar modes
- the focused sidebar calendar uses a highlighted border, matching the main pane focus treatment

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
- Link uses `/` to enter query input mode; while typing, `Left` / `Right` move the cursor and typed text is inserted at the cursor position
- `Selected Day` replaces the old focused-day panel name and stays synced with the main selection
- `Selected Day` also syncs when the sidebar calendar changes the selected date
