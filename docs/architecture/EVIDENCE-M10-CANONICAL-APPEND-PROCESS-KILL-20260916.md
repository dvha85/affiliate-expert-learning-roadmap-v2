# M10 process-kill recovery after canonical append — 2026-09-16

Trạng thái: **local automated verification / synthetic fixture / read-only**.

## Scope

This record covers the later sides of the two-store M10 recovery transitions.
The real learner Bot command runs in a separately executed test-binary child.
The child publishes the exact recovery journal, makes one canonical append
visible, and then receives POSIX `SIGKILL` before journal cleanup. The tested
boundaries are the immutable canary grant, the mutable canary state binding,
the immutable trusted cost-bound, and its compact index.

A fresh read-only process must return `RECOVERY_REQUIRED` rather than expose a
half-completed transition. A later locked writer must recognize the exact
canonical artifact or index entry, complete the remaining side, remove the
journal once, and leave exactly one artifact/index entry. This is a bounded
local POSIX process-termination seam; it is not proof of power-loss
durability, filesystem-crash behavior, atomicity across all runtime files,
Windows native locking, distributed/multi-host coordination, provider
operation, business outcomes, pilot or deployment readiness.

## Scenario and result

- `base_commit`: `1c071ca539de7e77050f43804d91f36884921b58`
- `implementation_commit`: `ff8ac97f3bbbe58c8d81bd293298f9f03ad51142`
- `host`: local macOS arm64
- `test`: `TestMissionM10ProcessKillAfterCanonicalAppendRequiresLockedReplay`
- `child`: `TestM10ProcessTerminationChild`
- `signal`: POSIX `SIGKILL`
- result: PASS for canary `after_artifact` and `after_state`, and trusted
  cost-bound `after_artifact` and `after_index`

For every scenario the parent observed the actual signal and the recovery
journal remained present. Before recovery, a fresh `mission status` process
returned `RECOVERY_REQUIRED`. The assertions also confirmed the expected
visible/absent side at the injected boundary. A locked retry converged to
`ACK` for canary or `EXACT_DUPLICATE` for cost-bound, removed the journal, and
left exactly one canonical artifact; the cost-bound retry also left exactly
one compact index entry.

## Verification

```text
GOWORK=off go test -count=1 ./cmd/bot -run TestMissionM10ProcessKillAfterCanonicalAppendRequiresLockedReplay  PASS
GOWORK=off go test -race -count=1 ./cmd/bot -run TestMissionM10ProcessKillAfterCanonicalAppendRequiresLockedReplay  PASS
```

The regression uses the production learner command dispatch, managed runtime
lock, M10 journals, canonical registry/state/index recovery and a fresh Bot
binary for the read-only observation and locked retry. The termination hook is
test-only and cannot be selected through normal CLI input.

## Limits

The tested post-append boundaries confirm local visibility and replay
idempotency; they do not establish filesystem metadata durability after power
loss or an atomic transaction across M10/M11/runtime files. Overall readiness
remains `NOT_READY_FOR_PRODUCTION`.
