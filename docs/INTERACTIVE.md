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
- `shelf mv` (when `<id>` and/or `--parent` is omitted)
- `shelf done` (when `<id>` is omitted; `status!=done` tasks are prioritized)
- `shelf links` (when `<id>` is omitted)
- `shelf triage` (without `--auto`)

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

## show / explain / edit / set / done / links Interactive Flow

1. Select target task by ID/title
   - Uses hierarchical tree-style labels without IDs by default
2. (`set` only, when no update flags are passed) choose fields in a menu and edit interactively (`Title`/`Kind`/`Status`/`Tags`/`Due`/`Repeat`/`Parent`/`Body replace`/`Body append`)
3. `set` shows change preview before apply

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
