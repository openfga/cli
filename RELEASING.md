# Releasing

Releases are automated with [release-please](https://github.com/googleapis/release-please).

## Cutting a release

1. Go to **Actions → release-please → Run workflow**. Pick a bump type: `auto` follows conventional commits, `patch`/`minor`/`major` force that bump, and `explicit` uses the exact version entered in `release-version`.
2. release-please opens a PR titled `release: vX.Y.Z` with the version bump and updated `CHANGELOG.md` — review it.
3. Merge it. CI tags the commit and publishes the release automatically.

## Version bumps

Bumps follow [conventional commits](https://www.conventionalcommits.org/), with `bump-minor-pre-major` and `bump-patch-for-minor-pre-major` enabled:

| Commit | Before v1.0.0 | v1.0.0 and later |
|---|---|---|
| `feat:` | patch | minor |
| `fix:`, `perf:`, `refactor:`, `revert:`, `docs:` | patch | patch |
| `type!:` or `BREAKING CHANGE:` footer | minor | major |

## Changelog

Built from commit messages since the last release: `feat:` → Added, `fix:` → Fixed, `perf:`/`refactor:` → Changed, `revert:` → Removed, `docs:` → Documentation. `test:`, `ci:` and `chore:` commits are hidden.

### Overriding an entry

If a commit message isn't user-facing, add this block to that PR's description before running the release workflow (already-merged PRs work too). release-please uses it instead of the commit message:

```
BEGIN_COMMIT_OVERRIDE
feat: user-facing description of the change
END_COMMIT_OVERRIDE
```

### Breaking changes

Mark the type with `!` and add a `BREAKING CHANGE:` footer. PRs are squash-merged, so check the merge dialog before merging: the squash title needs the `!` (single-commit PRs default to the commit title, not the PR title) and the body needs the footer:

```
feat!: remove the Foo option

BREAKING CHANGE: the Foo option has been removed. Use Bar instead.
```

This gets its own changelog section and triggers a major bump (minor before v1.0.0).

## Pre-releases

Pre-releases (alpha/beta/rc) are cut from the `develop` branch: **Actions → release-please-develop → Run workflow**, supplying an explicit pre-release version (e.g. `0.9.0-beta.1`). Sync `develop` with `main` first. Stable releases only go through `main`. Note: pre-release tags currently run the same goreleaser publish as stable ones, so the Homebrew formula and the Docker `latest`/`vX`/`vX.Y` tags move too.
