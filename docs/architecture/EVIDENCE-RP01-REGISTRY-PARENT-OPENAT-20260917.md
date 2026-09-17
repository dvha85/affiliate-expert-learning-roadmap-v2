# RP-01 registry append parent guard — 2026-09-17

## Scope

The M10 and M11 artifact registries share `openStableRegularFileForAppend`.
Before this change, a missing registry used `os.OpenFile` after a final-name
preflight. Replacing the parent directory with a symlink in that interval
could send the first registry creation through an external tree.

The POSIX implementation now pins each existing parent directory with
`O_DIRECTORY|O_NOFOLLOW`, verifies that the selected parent name still points
to the pinned inode after the test-only replacement seam, and opens the final
entry with `openat` plus `O_NOFOLLOW` and `O_EXCL` for a new file. The fallback
implementation retains the conservative regular-file and `O_EXCL` checks; it
does not claim native descriptor-pinned parity on Windows or other platforms.

## Regression

`TestM10RegistryAppendRejectsParentSwapBeforeOpenat` and
`TestM11RegistryAppendRejectsParentSwapBeforeOpenat` call the real registry
registration paths. The test swaps the selected parent to a symlink after the
parent descriptor is pinned and asserts that the external sentinel remains
unchanged and no registry is created below the external directory.

The regression is added to the learner Bot test suite and wired to the normal
`go test -race ./...` CI path. It is not marked as verified here because the
current worktree has no `go` executable available on PATH; the required CI
run is still pending.

## Boundary

This is bounded local pathname hardening for the shared M10/M11 append helper.
It does not prove arbitrary non-cooperating writer races, power-loss,
atomicity across multiple files, distributed locking, live execution,
provider/source evidence, business outcome, clean-machine pilot, or deployment
recovery.
