# Managed lock directory-handle/openat hardening — 2026-09-16

## Scope

POSIX managed lock acquisition now pins each existing directory component with
`O_DIRECTORY|O_NOFOLLOW`, then opens the final lock entry with `openat` against
the pinned parent descriptor and `O_NOFOLLOW`. This prevents a parent that is
replaced by a symlink after the initial preflight from redirecting lock creation
outside the requested tree.

The implementation canonicalizes only the already-existing prefix first, so
platform prefixes such as macOS `/var -> /private/var` remain usable. Missing
output components are not created during this step. The unresolved suffix is
then traversed by descriptors, so a workspace ancestor replaced after this
resolution is still rejected rather than followed.

This is a bounded local filesystem seam on supported POSIX targets. It does not
claim protection for every filesystem race outside the descriptor traversal,
distributed/multi-host locking, Windows native locking, power-loss, multi-file
atomicity, provider, business outcome, live executor, pilot or deployment.

## Result

- `repo_commit_before_change`: `7309887` (`main`)
- `host`: local macOS arm64
- `targeted_tests`: `TestManagedPathLockRejects*` — PASS
- `full_go_test`: `GOWORK=off go test -count=1 ./...` — PASS
- `go_vet`: `GOWORK=off go vet ./...` — PASS
- `full_go_race`: `GOWORK=off go test -race -count=1 ./...` — PASS
- `python_regression`: `python3 -m unittest discover -s scripts/tests -v` — PASS (110 tests)
- `readiness_audit`: `python3 scripts/audit_readiness.py` — PASS as an audit command;
  reported status remains `NOT_READY_FOR_PRODUCTION`
- `json_and_diff_checks`: JSON parse and `git diff --check` — PASS

The real regression first passes the parent preflight, replaces the requested
parent with a symlink to an external directory in the test seam, and then
attempts the lock acquisition. Descriptor-pinned traversal rejects the symlink;
the external sentinel remains byte-identical and the external lock file is not
created. The final-component and ordinary symlink-parent regressions also pass,
and the normal runtime/store suite passes on the macOS system prefix.
