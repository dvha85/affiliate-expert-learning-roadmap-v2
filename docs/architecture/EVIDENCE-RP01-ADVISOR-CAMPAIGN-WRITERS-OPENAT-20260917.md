# RP-01 advisor campaign writers parent guard — 2026-09-17

## Scope

The remaining local writer inventory now uses the shared stable append boundary:
the advisor campaign manifest, reservation attempts, campaign results, the
offline fixture JSON writer, and backup staging files. Campaign report,
campaign result and canary paths also acquire the managed local lock before
reading or writing their campaign state.

On POSIX, the shared append helper pins every existing parent directory with
`O_DIRECTORY|O_NOFOLLOW`, rechecks that the pathname still names the pinned
directory, and opens the final entry with `openat`, `O_NOFOLLOW` and `O_EXCL`
for new files. The fallback retains the conservative `O_EXCL` regular-file
check and does not claim native descriptor-pinned parity on Windows or other
platforms.

## Regression

The real learner Bot tests cover campaign initialization, reservation, result
persistence, fixture JSON and backup staging. Each test swaps the selected
parent to an external symlink at the shared append seam and requires rejection
before the external sentinel or target file changes. The audit also checks
that every inventoried writer remains on the shared append/managed-lock path.

The regression is wired to the normal `go test -race ./...` CI path. The current
worktree has no Go executable available on PATH, so local Go PASS is not
claimed; remote CI is the acceptance gate.

## Boundary

This is bounded local pathname and single-host writer hardening. It does not
prove arbitrary non-cooperating writer races, Windows native-lock parity,
power-loss or filesystem-crash durability, atomicity across multiple files,
distributed locking, provider/source evidence, live execution, business
outcomes, clean-machine pilot, or deployment recovery.
