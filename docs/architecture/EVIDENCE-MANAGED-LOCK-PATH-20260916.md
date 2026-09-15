# Managed lock path guard — 2026-09-16

## Scope

This record covers the POSIX managed-lock path guard at implementation commit
`40b20103d79283624eca55ae3cd1a90d80bc2d17`. The learner Bot uses
`O_NOFOLLOW` when opening runtime, backup-target, and restore-target advisory
lock paths. A real test replaces a lock pathname with a symlink to an external
file; acquisition is rejected and the external bytes remain unchanged.

This is a bounded local pathname check. It does not prove ancestor-path TOCTOU
protection, a native Windows lock, power-loss or atomic multi-file recovery,
multi-host coordination, provider operation, business outcome, pilot, or
deployment readiness.

## Verification

| Check | Result |
|---|---|
| `GOWORK=off go test ./cmd/bot -run '^(TestManagedPathLockRejectsSymlinkWithoutTouchingExternal|TestBackupProcessExitBeforePublishLeavesNoTargetAndRetrySucceeds|TestRuntimeGateRejectsAnotherProcess|TestRestoreTargetGateRejectsConcurrentManagedPublisher)$' -count=1 -v` | PASS |
| `python3 scripts/audit_readiness.py` | `NOT_READY_FOR_PRODUCTION` |
| `python3 -m unittest discover -s scripts/tests -v` | PASS; 107 tests |
| `python3 -m compileall -q scripts` and `git diff --check` | PASS |

## Interpretation

The lock pathname regression is wired to the learner Bot test shards and
required by the structured readiness audit. The repository remains
`NOT_READY_FOR_PRODUCTION`; broader filesystem and operated evidence remains
open.
