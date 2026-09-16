# M10 execution journal process-termination recovery (2026-09-16)

Trạng thái: **local automated verification / synthetic fixture / read-only**.

## Run record

- `run_id`: `m10-execution-process-kill-20260916`
- `implementation_commit`: `2571ec74fb14e0aaf7271f2afb96fd4f6724cd63`
- `merge_commit`: to be recorded after the PR is merged
- `host`: local Darwin arm64
- `scope`: learner Bot M10 governed execution journal transition

## Scenario and result

The real learner Bot test binary prepared a governed M10 authorization and
reservation, then started a separate child process for `m10-record-failed`.
The child was terminated by `SIGKILL` immediately after the synced execution
recovery journal became visible and before the canonical execution registry or
reservation binding was written.

The parent observed the actual signal, the journal remained present, no
portable output was created, the reservation had no execution binding, and no
execution record was visible in the canonical registry. A fresh read-only Bot
returned `RECOVERY_REQUIRED`. A later locked writer replayed the exact journal
once, removed it, published the portable record, and restored the reservation
→ execution binding.

## Verification

```text
GOWORK=off go test ./cmd/bot -run TestMissionM10ExecutionProcessKillAfterJournalPublishRequiresLockedReplay -count=1  PASS
```

The regression is
`lab/affiliate-bot/cmd/bot/m10_process_kill_test.go` and invokes the real
learner Bot command dispatch in a separate process. The termination hook is
test-only and cannot be selected through production CLI input.

## Limits

This is a bounded local POSIX process-termination seam. It does not prove
kernel power-loss durability, filesystem-crash behavior, atomicity across all
M10/M11 files, Windows native locking, distributed/multi-host coordination,
live executor behavior, provider operation, business outcome, pilot or
deployment readiness. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
