# TESTING

## Important checks

- `init` is idempotent
- `link` / `unlink` update edge files without duplicates
- `links` resolves outbound and inbound relations correctly
- Cockpit mode switches preserve selection where possible
- sidebar `Calendar` and `Selected Day` stay synchronized with the main pane
- inherited due dates appear consistently across calendar, tree, review, and now views

## Stable diffs

- edge ordering stays deterministic
- tree ordering stays deterministic

## Conflict resistance

- editing task bodies and editing links stay split across task files and edge files
