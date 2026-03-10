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
shelf next --format csv --fields id,title,path --no-header
shelf link --from 01AAA --to 01BBB --type depends_on
shelf links 01AAA
```

### 5. Compose with shell tools when needed

```bash
# preview ready tasks in jq
shelf next --json | jq '.[].path'

# sort ready tasks by due date before inspection
shelf next --format tsv --fields title,due_on --sort due_on

# choose a task interactively with fzf
shelf next --format tsv --fields id,title,path | fzf --with-nth=2,3 | cut -f1 | xargs -r shelf show

# the same flow using csv without a header row
shelf next --format csv --fields id,title,path --no-header | fzf --delimiter=, --with-nth=2,3 | cut -d, -f1 | xargs -r shelf show

# jump from search results to task files
shelf ls --format tsv --fields file,title,path | fzf --with-nth=2,3 | cut -f1 | xargs -r ${EDITOR:-vi}

# reverse-sort titles before piping elsewhere
shelf ls --format tsv --fields title,path --sort title --reverse

# inspect dependency targets with jq
shelf links 01AAA --json | jq '.outbound[] | {type, path, file}'

# read one task as a single tsv row
shelf show 01AAA --format tsv --fields id,title,file,body

# inspect saved copy presets as csv
shelf config copy-preset list --format csv --fields name,scope,subtree_style
```

## When to Use Which Mode

- `calendar`: date-first planning
- `tree`: parent-child structure and moves
- `board`: status-driven work management
- `review`: inbox / overdue / blocked / ready scan
- `now`: focused execution for today
