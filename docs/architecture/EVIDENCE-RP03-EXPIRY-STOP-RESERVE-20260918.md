# RP-03 approval expiry and STOP/reserve fail-closed boundary

## Scope

The real learner Bot now has direct regressions for M10 reservation admission,
not only M10 authorization admission. For each intent, approval, canary grant
and trusted cost-bound expiry, `m10-reserve` succeeds immediately before the
boundary and rejects at and after the exact RFC3339 boundary without changing
the canonical runtime snapshot. A second regression persists durable M11 STOP,
then proves `m10-reserve` returns `STOPPED` without changing state, counters,
reservations or registries.

The existing synchronized offline smoke remains the cross-process STOP versus
M10 reserve race evidence: whichever writer acquires the local runtime gate
first may complete, STOP becomes durable, and no later reservation is accepted.

## Verification status

The new regressions are wired to the hosted learner race gate:

```text
go test -race ./...
python scripts/smoke_br16a_offline.py
```

The local pre-edit baseline and post-edit Go execution could not start because
the current Windows worktree has no `go.exe`; hosted CI is the acceptance gate.
Readiness audit and its Python unit suite remain separately runnable locally.

## Boundary

This is bounded local/offline/synthetic/read-only fail-closed evidence. It does
not prove multi-file crash or power-loss atomicity, distributed locking,
provider access, live execution, business outcome, pilot or deployment
readiness. RP-03 and the overall tracker remain `PARTIAL` and
`NOT_READY_FOR_PRODUCTION`.
