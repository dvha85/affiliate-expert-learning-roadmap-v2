# M11 process-kill recovery after ledger append — 2026-09-16

Trạng thái: **local automated verification / synthetic fixture / read-only**.

## Scope

This record covers the later side of the M11 execution transition. The real
learner Bot command runs in a separately executed test-binary child. The child
publishes the exact recovery journal, appends the execution/STOP ledger
transition, directory-syncs that registry append, and then receives POSIX
`SIGKILL` before journal cleanup.

The fresh read-only process must return `RECOVERY_REQUIRED`, even though the
execution and ledger transition are already visible. A later locked writer
must recognize that exact pair rather than append either side twice, remove the
journal, preserve the FAILED ledger or durable UNKNOWN→STOP state, and retain
the artifact cardinality.

This is a bounded local POSIX process-termination seam. It is not proof of
power-loss durability, filesystem-crash behavior, atomicity across multiple
runtime files, Windows native locking, distributed/multi-host coordination,
provider operation, business outcomes, pilot or deployment readiness.

## Scenario and result

- `implementation_commit_before_change`: `780bc6bf2ab0774d684f0eb3cc856e6531c32304`
- `host`: local macOS arm64
- `test`: `TestM11ProcessKillAfterLedgerAppendLeavesJournalForFreshRecovery`
- `child`: `TestM11ProcessTerminationChild`
- `injected_boundary`: after ledger `after_sync`, before journal cleanup
- `signal`: POSIX `SIGKILL`
- result: PASS for both `failed` and `unknown` (`UNKNOWN→STOP`) cases

The parent test observed the signal and the journal remained present. Before
recovery, a fresh `mission status` process returned `RECOVERY_REQUIRED`; the
registry already contained exactly one execution artifact and the expected
three ledger artifacts. A locked writer then recovered before processing an
invalid input: the FAILED case retained its exact failure ledger, while the
UNKNOWN case retained durable STOP. The recovery journal disappeared and the
execution/ledger counts stayed unchanged.

## Verification

```text
GOWORK=off go test -race -count=1 ./cmd/bot -run TestM11ProcessKillAfterLedgerAppendLeavesJournalForFreshRecovery  PASS
```

The regression uses the production learner command dispatch, managed lock,
registry journal and recovery implementation. The fault hook is test-only and
is not selectable through normal CLI input.

## Limits

The tested `after_sync` boundary confirms visibility and replay idempotency on
the local POSIX path; it does not establish that all filesystem metadata is
durable after power loss or that the state/registry/journal transition is an
atomic multi-file transaction. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
