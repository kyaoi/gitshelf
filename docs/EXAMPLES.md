# EXAMPLES

## Start the main workspace

```sh
shelf
shelf review
shelf tree
```

## Read-only listing

```sh
shelf ls --kind todo --status open
shelf ls --tag backend --json
shelf next
```

## Direct link management

```sh
shelf link --from 01AAA --to 01BBB --type depends_on
shelf unlink --from 01AAA --to 01BBB --type depends_on
shelf links 01AAA
```
