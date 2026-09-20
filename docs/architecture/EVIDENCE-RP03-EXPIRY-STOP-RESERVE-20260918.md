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

## Post-merge evidence

PR #421 was squash-merged into `main` at
`12c50f3337991b533e5ecf363860df1e27bc2217` from implementation head
`3620b24cc3b51d6490473c40103f328c83bc8c20`. The hosted acceptance runs passed
all 13 checks: Curriculum CI run `35323363318` (including Windows runtime job
`105530687600` and learner race job `105530687837`) and Mission Agent Path CI
run `35323363297` (3/3 jobs PASS). Post-merge readiness audit and 118 audit
unit tests pass; local Go execution remains unavailable because `go.exe` is not
installed in this worktree.

## Boundary

This is bounded local/offline/synthetic/read-only fail-closed evidence. It does
not prove multi-file crash or power-loss atomicity, distributed locking,
provider access, live execution, business outcome, pilot or deployment
readiness. RP-03 and the overall tracker remain `PARTIAL` and
`NOT_READY_FOR_PRODUCTION`.
