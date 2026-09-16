# Evidence: M11 recovery admission concurrency

## Scope

- Base checkout before this change: `main` at `b9630a85c060e0b28401d3dc7e8e1ea5a5402586`.
- The test uses a real learner Bot binary in separate OS processes, one old
  stopped runtime, one prepared new runtime/lease, and one reviewed recovery
  handoff/admission input.
- This is local fixture/read-only concurrency evidence. It does not prove a
  live executor, provider traffic, business outcome, target deployment,
  distributed/multi-host locking, or power-loss/atomic multi-file durability.

## Verification

```text
cd lab/affiliate-bot
GOWORK=off go test ./cmd/bot -run TestMissionM11RecoveryAdmissionSingleWriterAcrossBotProcesses -count=1
PASS

GOWORK=off go test ./cmd/bot -run TestMissionM11RecoveryAdmissionSingleWriterAcrossBotProcesses -count=5
PASS
```

Each run starts eight separately executing Bot processes behind a common
barrier. Exactly one process appends the immutable recovery admission. Other
processes can only receive `BUSY` while the new-runtime gate is held or
`EXACT_DUPLICATE` after the first append. The test then loads the registry and
requires exactly one admission, and a fresh Bot exact retry returns
`EXACT_DUPLICATE`. The old runtime remains the reviewed stopped source.

The existing CI command `go test -race ./...` executes this regression on pull
requests and pushes to `main`.

## Boundary

This closes the tested local concurrent-writer seam for M11 recovery admission
only. RP-07 remains `PARTIAL`; crash/power-loss and multi-file atomicity,
distributed locking, live execution, provider/business outcome, pilot and
deployment evidence remain open. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
