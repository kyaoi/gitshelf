# TESTING

`gitshelf` treats the heavy suite as the default suite.

## Local hooks

Install hooks locally with:

```bash
lefthook install
```

Hook policy:

- `pre-commit`: fast local guardrails only
- `pre-push`: the same heavy suite used before merge
- `pre-commit` also checks that public docs do not drift back to removed CLI schema terms

## Required checks

Run all of these before merging:

- `gofmt -w .`
- `go test ./...`
- `bash scripts/check_coverage_ratchet.sh`
- `go test -race ./...`
- `go vet ./...`
- `bash scripts/check_public_contract_docs.sh`

You can run the same heavy suite locally with:

```bash
bash scripts/run_heavy_checks.sh
```

## Regression policy

- fix bugs with `repro test first`
- treat machine-readable CLI output as a public contract
- update golden fixtures only intentionally
- do not lower package coverage below [`scripts/coverage_baseline.txt`](../scripts/coverage_baseline.txt)

## Golden fixtures

CLI output snapshots live under:

- `internal/cli/testdata/outputs/`

Refresh them only when the public output contract is intentionally changed:

```bash
UPDATE_GOLDEN=1 go test ./internal/cli -run TestCLIMachineReadableOutputGoldens
```

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

## Release smoke

For a tagged release such as `v1.3.1`:

1. `go install github.com/kyaoi/gitshelf/cmd/shelf@v1.3.1`
2. `shelf --version` prints `v1.3.1`
3. `mise use -g go:github.com/kyaoi/gitshelf/cmd/shelf@latest`
4. `shelf --version` prints the latest tag, not `dev`

GitHub Actions also runs [`release-smoke.yml`](../.github/workflows/release-smoke.yml)
for tag pushes and manual release smoke checks.

GitHub Actions also runs [`release.yml`](../.github/workflows/release.yml)
to create a GitHub Release automatically on tag push.
The same workflow also supports manual backfill for any existing tags by
passing a JSON array to `tags_json`.

For future releases, prefer full semver tags such as `v1.3.1` if you want
exact `go install module@version` resolution through the Go module proxy.
