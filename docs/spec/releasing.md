# Releasing

**Status:** accepted

FX releases are cut with [`platform`](https://platform.prodigy9.co) using the `semver`
strategy. The repo ships a `./platform` wrapper that pins a known-good `platform`
version (`go run platform.prodigy9.co@<version>`), so contributors don't need a global
install.

## Pre-release checklist

1. **Working tree is clean.** `platform release` refuses to proceed otherwise. Use
   `--force` only when intentionally releasing on top of uncommitted local-only files
   (rare; usually you should commit or stash first).
2. **All commits are pushed.** `platform release` pushes the new tag, but does not push
   commits — push `gh main` (or the release branch) yourself first. A tag pointing at
   a commit that isn't on the remote will fail to fetch for anyone running
   `go get fx.prodigy9.co@vX.Y.Z`.
3. **Tests pass.** `go test ./...` from the repo root. Examples under `examples/` are
   not in the default test run; if the release touches code an example exercises,
   verify it manually too.
4. **`CHANGELOG.md` updated.** Add the new version section at the top (see below).
   Releases without a changelog entry are valid but unhelpful.

## Versioning

Pre-1.0 conventions (subject to change at 1.0):

- **Patch** (`--patch`) — bug fixes, internal refactors, and small additive features
  inside an existing package. Default choice. v0.8.x has been shipping new packages
  (`app/files`, `fxtest`) as patches.
- **Minor** (`--minor`) — large new packages, behavior changes that downstream apps
  may need to react to, or breaking changes within a package that isn't yet considered
  stable.
- **Major** (`--major`) — reserved for 1.0 and beyond. Don't use pre-1.0.
- **Explicit version** (`vX.Y.Z`) — when none of the above fits, name it.

When in doubt: patch. Pre-1.0 callers should expect breakage at minor boundaries
anyway, so the cost of under-bumping is low and the cost of burning a minor on a
trivial change is real (fewer minors available before 1.0).

**Docs-only changes don't warrant a release.** A tag should carry a code change; fold
spec/doc/changelog updates into the next code-bearing release rather than tagging them
alone. If a change touches nothing under a consumer's `go get`, it does not need a version.

## CHANGELOG conventions

Update `CHANGELOG.md` only at release time, not per-commit. Format mirrors prior
entries:

```markdown
## v0.8.5

* **package:** One-line description of what changed and why a user cares.
* **package:** Group multiple related changes under the same scope.
```

What to include:

- User-facing additions, behavior changes, bug fixes
- New packages or commands
- Docs changes that downstream readers should know about (e.g. moved canonical
  reference, new spec section)

What to omit:

- Pure-internal chore (`.gitignore` edits, ACE/tooling config, CI tweaks)
- Refactors with no behavior change
- Typo fixes in code comments

The scope label (`**package:**`) is the package or area, not the commit's
conventional-commit type. `**migrator:**`, not `**fix:**`.

## Running the release

```sh
./platform release --patch    # most common
./platform release --minor
./platform release v0.9.0     # explicit
./platform release --force    # dirty tree (avoid)
```

`platform release` prompts `create this release? [y/N]` before tagging. To confirm
non-interactively (automation, or driving it from an agent), set `ALWAYS_YES=1` — a
piped `y` (`echo y | …`) does **not** work, since the prompt reads the tty directly:

```sh
ALWAYS_YES=1 ./platform release --patch
```

What `platform release` does:

1. Computes the next version from the strategy (`semver` + flag).
2. Updates any version files configured in `platform.toml` (none for FX).
3. Creates and pushes a git tag (`vX.Y.Z`) to the configured remote.

What it does **not** do:

- Push commits — that's your job, pre-release.
- Edit `CHANGELOG.md` — that's your job, pre-release.
- Publish to a registry — Go modules use git tags directly, no separate publish.

## After the release

- Verify the tag landed: `git fetch --tags && git tag | tail -3`.
- Downstream consumers pick it up via `go get fx.prodigy9.co@vX.Y.Z` or
  `go get fx.prodigy9.co@latest`.
- If the release introduced a notable behavior change, update affected downstream apps
  in their own PRs — don't bulk-bump from this repo.
