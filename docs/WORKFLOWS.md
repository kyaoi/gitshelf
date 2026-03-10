# WORKFLOWS

Practical usage for the current Cockpit-first toolset.

## Default Rule

Use this as the default mental model:

- start with `shelf`
- stay inside `Cockpit`
- use `ls` or `next` only when you need a direct textual answer or JSON
- use `link` / `unlink` / `links` only when you need direct scripting for relations

Detailed keybindings live only in [`INTERACTIVE.md`](INTERACTIVE.md).

## Main Entry Points

- `shelf`
  - on TTY, opens Cockpit directly
- `shelf cockpit`
  - same workspace, explicit form
- `shelf calendar`
  - open Cockpit in calendar mode
- `shelf tree`
  - open Cockpit in tree mode
- `shelf board`
  - open Cockpit in board mode
- `shelf review`
  - open Cockpit in review mode
- `shelf now`
  - open Cockpit in now mode

## Recommended Daily Flow

### 1. Open the workspace

```bash
shelf
```

### 2. Plan and inspect in Cockpit

- start in `calendar` for date-first planning
- switch to `tree` when you need hierarchy or moves
- use `board` for status-oriented work
- use `review` for inbox / overdue / blocked / ready scanning
- use `now` for today-focused execution

In non-calendar modes, the sidebar `Calendar` and `Selected Day` stay synchronized with the main selection.

### 3. Edit inside the TUI

- create tasks from the current context
- change status, kind, tags, due dates, and links in place
- use centered popups for add, link, tag, filter, and other transient editors

### 4. Use direct commands only for scripting or quick answers

```bash
shelf ls --status open --json
shelf next --json
shelf next --format tsv
shelf link --from 01AAA --to 01BBB --type depends_on
shelf links 01AAA
```

### 5. Compose with shell tools when needed

```bash
# preview ready tasks in jq
shelf next --json | jq '.[].path'

# choose a task interactively with fzf
shelf next --format tsv | fzf --with-nth=2,3 | cut -f1 | xargs -r shelf show

# jump from search results to task files
shelf ls --format tsv | fzf --with-nth=2,3 | cut -f12 | xargs -r ${EDITOR:-vi}
```

## When to Use Which Mode

- `calendar`: date-first planning
- `tree`: parent-child structure and moves
- `board`: status-driven work management
- `review`: inbox / overdue / blocked / ready scan
- `now`: focused execution for today
