# Backup/restore process-exit regression — 2026-09-16

## Scope

This record covers one bounded local failure seam at implementation commit
`44e34b4abc79c91547ee193915f8f9e881b7c284`. A real learner Bot test process is
terminated while backup or restore output is still private staging. POSIX
advisory locks release when that process exits, the incomplete target remains
absent, and an exact retry creates/restores a valid target.

This is not a power-loss, kernel-crash, atomic multi-file, Windows-native-lock,
multi-host, or orphan-staging-cleanup proof. Provider, live executor, business
outcome, pilot, and deployment evidence are not involved.

## Verification

| Check | Result |
|---|---|
| `GOWORK=off go test ./cmd/bot -run '^(TestBackupProcessExitBeforePublishLeavesNoTargetAndRetrySucceeds|TestRuntimeGateRejectsAnotherProcess|TestRestoreTargetGateRejectsConcurrentManagedPublisher)$' -count=1 -v` | PASS |
| `GOWORK=off go test -count=1 ./...` and `GOWORK=off go vet ./...` in `contracts`, `core`, `lab/affiliate-bot`, and `lab/mission-runtime` | PASS |
| `python3 -m unittest discover -s scripts/tests -v` | PASS; 106 tests |
| `python3 scripts/audit_readiness.py` | `NOT_READY_FOR_PRODUCTION` |
| `python3 -m unittest scripts.tests.test_audit_readiness -v` | PASS; 91 tests |
| `python3 -m compileall -q scripts` and `git diff --check` | PASS |

## Interpretation

The local process-exit seam is now covered by the learner Bot test shard and
the structured readiness audit. The repository remains
`NOT_READY_FOR_PRODUCTION`; the broader crash/power-loss and operated gaps stay
open.
