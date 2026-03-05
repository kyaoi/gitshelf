# INTERACTIVE (TTY Selection Spec)

Interactive mode is available only when stdin/stdout are TTY.
If not TTY, users must provide required flags.

## Supported Commands

- `shelf add` (when `--title` is omitted)
- `shelf link` (when `--from/--to/--type` are omitted)
- `shelf unlink` (when `--from/--to/--type` are omitted)
- `shelf show` (when `<id>` is omitted)
- `shelf set` (when `<id>` is omitted)
- `shelf mv` (when `<id>` and/or `--parent` is omitted)
- `shelf done` (when `<id>` is omitted; `status!=done` tasks are prioritized)
- `shelf links` (when `<id>` is omitted)

## Key Bindings

- `j` / `k`: move selection down/up
- `Enter`: confirm
- `/`: search mode
- `Esc`:
  - in search mode: clear search and leave search mode
  - otherwise: cancel selection
- `Ctrl+C`: cancel selection
- Arrow up/down are also supported.

## Search

- `/` enters incremental search.
- Search matches option label/search text (task title and short/full ID).
- Result list updates as user types.

## Display

Task candidate line format:

`[{short}] {title}  ({kind}/{status})`

## add Interactive Flow

1. Input `Title`
2. Select `Kind`
3. Select `Status`
4. Select `Parent` (`0: [root]` means no parent)

## link Interactive Flow

1. Select source task
2. Select destination task
3. Select link type

The type selection screen includes this warning:

`A depends_on B = B must be done before A`

## unlink Interactive Flow

1. Select source task
2. Select existing outbound edge to remove

## show / set / done / links Interactive Flow

1. Select target task by ID/title
2. (`set` only, when no update flags are passed) select `Status`

## mv Interactive Flow

1. Select target task by ID/title (when `<id>` omitted)
2. Select new parent (`0: [root]`) when `--parent` omitted
