# Backup/restore orphan staging recovery — 2026-09-16

## Scope

Backup and restore staging directories now include a deterministic hash of the
caller-selected target. After the per-target managed lock is held, the next
retry removes only matching private staging directories that were left by an
interrupted writer. It does not follow symlinks or remove files, staging for a
different target, or a visible backup/runtime target.

## Result

- `repo_commit_before_change`: `e79e384` (`main`)
- `host`: local macOS arm64
- `TestBackupProcessExitBeforePublishLeavesNoTargetAndRetrySucceeds`: PASS
- `TestBackupProcessKillBeforePublishLeavesNoTargetAndRetrySucceeds`: PASS
- `TestBackupProcessKillAfterPartialStagingLeavesNoTargetAndRetrySucceeds`: PASS
- `TestCleanupStaleStagingOnlyRemovesTargetOwnedDirectories`: PASS

Each test runs the real learner Bot backup/restore command in a separate child
process, terminates it after target-owned staging creation and before the first
staged file write, confirms the final target is absent and exactly one matching
staging tree remains, then retries. The retry publishes a valid result and
leaves no target-owned staging tree. Staging for another target is not part of
this regression.

This is bounded local POSIX process-termination cleanup evidence. It does not
prove power-loss durability, atomic multi-file recovery, cleanup after a
filesystem crash, Windows native locking, distributed/multi-host behavior,
provider operation, business outcomes, pilot or deployment readiness.

## Verification rerun

- `python3 -m json.tool docs/plans/READINESS-MATRIX.json`: PASS
- `python3 -m json.tool docs/plans/READINESS-EVIDENCE-GRAPH.json`: PASS
- `python3 -m unittest discover -s scripts/tests -v`: PASS (112 tests)
- `python3 scripts/audit_readiness.py`: PASS; `NOT_READY_FOR_PRODUCTION`
- `GOWORK=off go test -count=1 ./...` and `GOWORK=off go vet ./...`: PASS in `lab/affiliate-bot`
- `GOWORK=off go test -race -count=1 ./...`: PASS in `lab/affiliate-bot`
- `git diff --check`: PASS
