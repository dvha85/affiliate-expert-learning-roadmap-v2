# RP-03 M10 canonical grant registry binding

## Scope

The learner M10 gate and reservation entrypoints now require the mutable mission
state's canary grant to resolve byte-for-byte to a registered canonical
`CANARY_GRANT` artifact before evaluating the gate or changing the reservation
ledger. A schema-valid grant that is only present in `mission-state.json` is
therefore rejected fail-closed.

The regression mutates a populated offline fixture with a newly hashed but
unregistered grant, then calls the real learner `m10-gate` and legacy
`m10-reserve` paths. Both reject before writing a portable gate, registry entry,
reservation or usage-counter change.

## Verification status

The implementation regression is wired to the hosted learner Go test/race gate:

```text
go test -race ./...
```

The current Windows worktree has no `go.exe`; the attempted local baseline
`go test -count=1 ./...` from `lab/affiliate-bot` could not start because the
binary is absent. Hosted CI remains the acceptance evidence for this change.

## Boundary

This is a bounded offline/read-only M10 lineage guard. It does not prove
multi-file crash or power-loss atomicity, distributed locking, trusted external
time, provider access, live execution, business outcome, pilot or deployment
readiness. RP-03 and the overall tracker remain `PARTIAL` and
`NOT_READY_FOR_PRODUCTION`.
