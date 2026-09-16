# M10 execution process-kill recovery after canonical append — 2026-09-16

Trạng thái: **local automated verification / synthetic fixture / read-only**.

## Scope

This record covers the later sides of the M10 governed execution transition.
The real learner Bot command runs in a separately executed test-binary child.
The child publishes the exact execution recovery journal, appends the
immutable execution record, or binds the reservation to that record, and then
receives POSIX `SIGKILL` before journal cleanup.

A fresh read-only process must return `RECOVERY_REQUIRED` rather than expose a
transition whose journal is still pending. A later locked retry must recognize
the exact execution record and reservation link, complete only the remaining
side, remove the journal, publish the portable record, and leave exactly one
canonical execution record and one reservation-to-execution binding. This is a
bounded local POSIX process-termination seam; it is not proof of power-loss
durability, filesystem-crash behavior, atomicity across all runtime files,
Windows native locking, distributed/multi-host coordination, provider
operation, business outcomes, pilot or deployment readiness.

## Scenario and result

- `base_commit`: `65c70409830b0a789d877bf708a0daec6a2a9be6`
- `implementation_commit`: `3670c20237779d6412ad0122a4fb32d48dbf27d1`
- `host`: local macOS arm64
- `test`: `TestMissionM10ExecutionProcessKillAfterCanonicalAppendRequiresLockedReplay`
- `child`: `TestM10ProcessTerminationChild`
- `signal`: POSIX `SIGKILL`
- result: PASS for execution `after_artifact` and `after_state`

For `after_artifact`, the execution registry record is visible while the
reservation remains unbound. For `after_state`, both the record and its exact
reservation binding are visible. In both cases the parent observed the actual
signal, the journal remained present, a fresh `mission status` process returned
`RECOVERY_REQUIRED`, and the portable output was absent before replay. A
locked retry returned `APPENDED`, removed the journal, created the portable
record, and left exactly one canonical execution record with the expected
reservation binding.

## Verification

```text
GOWORK=off go test -count=1 ./cmd/bot -run TestMissionM10ExecutionProcessKillAfterCanonicalAppendRequiresLockedReplay  PASS
GOWORK=off go test -race -count=1 ./cmd/bot -run 'TestMissionM10(ProcessKillAfterJournalPublishRequiresLockedReplay|ProcessKillAfterCanonicalAppendRequiresLockedReplay|ExecutionProcessKillAfterCanonicalAppendRequiresLockedReplay)$'  PASS
```

The regression uses the production learner command dispatch, managed runtime
lock, M10 execution journal, canonical registry/state recovery and a fresh Bot
binary for read-only observation and locked retry. The termination hook is
test-only and cannot be selected through normal CLI input.

## Limits

The tested post-append boundaries confirm local visibility and replay
idempotency; they do not establish filesystem metadata durability after power
loss or an atomic transaction across M10/M11/runtime files. Overall readiness
remains `NOT_READY_FOR_PRODUCTION`.
