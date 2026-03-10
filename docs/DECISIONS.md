# DECISIONS

## Storage

- tasks: `.shelf/tasks/<id>.md`
- hierarchy: `parent` in task front matter
- links: `.shelf/edges/<src_id>.toml`
- kind and status are independent

## UI

- daily editing is centered on `Cockpit`
- `calendar/tree/board/review/now` are launcher presets for `Cockpit`
- read-only queries use `ls` / `next`
- scriptable link operations use `link` / `unlink` / `links`

## IDs

- ULID
- short IDs for display
- IDs are hidden by default and shown with `--show-id`

## Locking

- mutating operations use `.shelf/.write.lock`
- lock acquisition failures return a timeout error
