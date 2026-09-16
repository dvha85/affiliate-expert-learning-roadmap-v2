# M11 process-kill journal recovery — 2026-09-16

## Scope

The learner Bot now uses the managed lock primitive for `.mission.lock` on
supported POSIX systems. The lock is a regular control file with an advisory
kernel lock, so a process termination releases ownership; the file itself is
safe to retain and is excluded from backup inventory. The existing cooperative
directory marker remains supported on Windows/non-POSIX targets.

The real test-binary child runs `m11-record-failed` and `m11-record-unknown`,
publishes the exact replay journal, then is killed immediately before the new
M11 ledger append. A fresh read-only process must return
`RECOVERY_REQUIRED`. A later locked writer replays the journal, removes the
stale lock/journal boundary as appropriate, and preserves the exact FAILED
ledger or durable UNKNOWN→STOP state.

This is a bounded local POSIX process-termination seam. It is not proof of
power-loss durability, atomic multi-file transactions, orphan-staging cleanup,
Windows native locking, distributed/multi-host locking, provider operation,
business outcomes, pilot or deployment readiness.

## Result

- `repo_commit_before_change`: `34493a5` (`main`)
- `host`: local macOS arm64
- `test`: `TestM11ProcessKillAfterJournalBeforeLedgerAppendLeavesJournalForFreshRecovery`
- result: PASS for both `failed` and `unknown` cases
- `mission_state_lock_regression`: `TestMissionStateLockRejectsConcurrentMutationWithoutWriting` — PASS

The regression uses the actual learner command dispatch in a separately
executed test process, observes `SIGKILL`, verifies the journal remains
available for recovery, and checks that the next writer is not blocked by a
stale `.mission.lock` marker. The implementation remains explicitly partial
for crash/power-loss and multi-file guarantees.

## Verification rerun

Executed after the M11 process-boundary change on the local working tree:

- `python3 -m json.tool docs/plans/READINESS-MATRIX.json`: PASS
- `python3 -m json.tool docs/plans/READINESS-EVIDENCE-GRAPH.json`: PASS
- `python3 -m unittest discover -s scripts/tests -v`: PASS (111 tests)
- `python3 scripts/audit_readiness.py`: PASS; `NOT_READY_FOR_PRODUCTION`
- `git diff --check`: PASS
- `GOWORK=off go test -count=1 ./...` and `GOWORK=off go vet ./...`: PASS for `core`, `contracts`, `lab/mission-runtime`, and `lab/affiliate-bot`
- `GOWORK=off go test -race -count=1 ./...` in `lab/affiliate-bot`: PASS
- `GOOS=linux GOARCH=amd64 GOWORK=off go test -c ./cmd/bot`: PASS

These are local offline/fixture checks. They do not add evidence for
power-loss atomicity, provider or business outcomes, clean-machine pilot,
target-host deployment, or distributed locking.
