# RP-03 exhausted-budget monotonicity boundary

## Scope

The learner M10 runtime now has a direct R08 regression after a cap=1 budget
has actually been consumed by a fresh Bot process. Exact M09 approval replay
and exact M10 grant re-import remain acknowledgement-only and do not reset the
usage counter. Checksum-valid attempts to increase the same grant's execution
cap, change its currency, or register a cost bound in another currency are
rejected without changing `mission-state.json`, the M10 artifact registry, or
the trusted cost-bound registry. A new reservation remains budget-denied.

The same sequence is repeated against a fresh runtime restored from a backup.
The regression compares the canonical state, usage counters, reservation list
and registries byte-for-byte after every rejected operation. This is a
monotonicity/replay boundary for the local synthetic learner runtime; it is not
a multi-file power-loss or distributed-ledger proof.

## Verification status

The regression is wired to the hosted learner race gate:

```text
go test -race ./...
```

The local targeted Go baseline and post-edit execution could not start because
the current Windows worktree has no `go.exe`. The local readiness audit and
118-test Python audit suite remain runnable; hosted Windows runtime and race CI
are the acceptance gate for the Go regression.

## Boundary

This is bounded local/offline/synthetic/read-only fail-closed evidence. It does
not prove multi-file crash or power-loss atomicity, distributed locking,
provider access, live execution, business outcome, pilot or deployment
readiness. RP-03 and the overall tracker remain `PARTIAL` and
`NOT_READY_FOR_PRODUCTION`.
