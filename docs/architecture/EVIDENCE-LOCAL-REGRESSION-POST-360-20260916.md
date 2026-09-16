# Local regression after PR #360 (2026-09-16)

Trạng thái: **local automated verification / fixture-only / read-only**.

## Run record

- `run_id`: `local-regression-post-360-20260916`
- `repo_commit`: `930a040c80d4295c7dac60980b96030d9b6439eb` (`main` sau khi merge PR #360)
- `host`: local Darwin arm64
- `result`: các lệnh kiểm chứng được chạy trong record này đều PASS

## Commands and results

Go modules:

```text
(cd contracts && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)       PASS
(cd core && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)            PASS
(cd lab/mission-runtime && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
(cd lab/affiliate-bot && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)  PASS
```

Repository tests and smoke:

```text
python3 -m unittest discover -s scripts/tests -v        PASS (113 tests)
python3 scripts/smoke_br16a_offline.py                  PASS
python3 scripts/smoke_br18b_backup_restore.py           PASS
```

Static validators also passed: repository/continuity/artifact-spine/language/
mission/agent/semantic-contract validators, M06 validators and M07 validators
including the real adversarial/output-case implementations. The two
`*_operated_execution.py` scripts were not run standalone because they require
an execution JSON plus store paths; the disposable n8n runner is the command
that supplies those artifacts when that path is specifically re-verified.

The readiness audit was rerun after the merge:

```text
python3 scripts/audit_readiness.py
READINESS AUDIT: NOT_READY_FOR_PRODUCTION
```

## Scope and limits

This record confirms the local learner/runtime and structured-audit baseline at
the post-PR #360 commit. It does not claim ACCESSTRADE/provider requests,
Cockpit operation, live executor activity, business outcome/payout, clean-
machine self-service pilot, target-host deployment/recovery, distributed
locking or power-loss/atomic multi-file durability. Those remain outside this
run and the readiness status remains `NOT_READY_FOR_PRODUCTION`.
