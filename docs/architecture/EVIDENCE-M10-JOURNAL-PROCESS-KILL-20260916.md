# M10 journal process-termination recovery (2026-09-16)

Trạng thái: **local automated verification / synthetic fixture / read-only**.

## Run record

- `run_id`: `m10-journal-process-kill-20260916`
- `implementation_commit`: `20a6225a5c17c53aeb15572198b2b0f1459f45be`
- `merge_commit`: to be recorded after the PR is merged
- `host`: local Darwin arm64
- `scope`: learner Bot M10 canary and trusted cost-bound journal transitions

## Scenario and result

The real learner Bot test binary starts each M10 command in a separate child
process and sends `SIGKILL` immediately after the synced recovery journal is
visible, before the first canonical artifact/state side effect:

```text
TestMissionM10ProcessKillAfterJournalPublishRequiresLockedReplay/canary      PASS
TestMissionM10ProcessKillAfterJournalPublishRequiresLockedReplay/cost-bound  PASS
```

For both transitions, the parent process observed the actual signal and the
journal remained present. A fresh read-only `mission status` returned
`RECOVERY_REQUIRED`, so it did not expose a partial M10 state. A later locked
writer retried the same request, replayed the exact journal once, removed the
journal, and returned the canonical acknowledgement (`ACK` for canary,
`EXACT_DUPLICATE` for cost-bound). The recovered cost-bound was also resolved
from the immutable M10 registry.

## Verification

```text
GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...   PASS
GOWORK=off go test -race -count=1 ./...                       PASS
```

The focused regression is
`lab/affiliate-bot/cmd/bot/m10_process_kill_test.go` and runs the production
command dispatch plus the real journal/recovery implementation. The kill hook
is test-only and cannot be selected through CLI input or environment by the
production command path.

## Limits

This is a bounded local POSIX process-termination seam. It does not prove
kernel power-loss durability, filesystem-crash behavior, atomicity across all
M10/M11 files, Windows native locking, distributed/multi-host coordination,
live executor behavior, provider operation, business outcome, pilot or
deployment readiness. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
