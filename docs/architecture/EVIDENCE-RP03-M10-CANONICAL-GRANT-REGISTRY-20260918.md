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

## Post-merge evidence

PR #419 was squash-merged into `main` at
`2acfcfc5a80ad590ce06813940ce9bbff83fd839` from implementation head
`104bd186b0546dd4c3e7d8c6a3eebc44fe639ecc`. The hosted acceptance runs passed
all 13 checks: Curriculum CI run `35321402705` (including Windows runtime job
`105524487971` and learner race job `105524487897`) and Mission Agent Path CI
run `35321402712` (3/3 jobs PASS). The local audit and 118 readiness-audit
unit tests also pass on the post-merge evidence branch; the local Go baseline
remains unavailable because `go.exe` is not installed in this worktree.

## Boundary

This is a bounded offline/read-only M10 lineage guard. It does not prove
multi-file crash or power-loss atomicity, distributed locking, trusted external
time, provider access, live execution, business outcome, pilot or deployment
readiness. RP-03 and the overall tracker remain `PARTIAL` and
`NOT_READY_FOR_PRODUCTION`.
