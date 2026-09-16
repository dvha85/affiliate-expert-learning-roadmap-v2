# Managed lock ancestor preflight — 2026-09-16

## Scope

POSIX `acquireManagedPathLock` now validates each existing ancestor of the
lock pathname with `Lstat` before opening the lock file. A symlinked parent is
rejected before `OpenFile` can resolve the lock outside the requested runtime
tree.

The regression is a bounded local preflight. It does not claim protection
against an ancestor being replaced concurrently between preflight and open;
that would require OS-specific directory-handle/openat hardening. It also does
not prove Windows native locking, power-loss, multi-host behavior, provider,
business outcome, live executor, pilot or deployment.

## Result

- `repo_commit_before_change`: `3d1a4a6` (`main`)
- `host`: local macOS arm64
- `test`: `TestManagedPathLockRejectsSymlinkedParentWithoutTouchingExternal`
- result: PASS

The real test creates `external/`, aliases it as the lock parent, and attempts
to acquire `alias/managed.lock`. Acquisition is rejected; the external
sentinel remains byte-identical and `external/managed.lock` is not created.
The existing final-component `O_NOFOLLOW` regression also remains PASS.
