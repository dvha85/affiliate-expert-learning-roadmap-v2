# Local runtime re-run (2026-09-16, main `2511efd`)

Trạng thái: **local automated verification / fixture-only / read-only**.

## Run record

- `run_id`: `local-runtime-rerun-20260916-2`
- `repo_commit`: `2511efd` (`main` trước khi ghi record)
- `host`: local Darwin arm64
- `scope`: learner Bot runtime, canonical graph, audit and test wiring
- `result`: all commands listed below PASS

## Commands and results

Runtime smoke bằng implementation thật của learner Bot:

```text
python3 scripts/smoke_br16a_offline.py
BR-16a PASS: one shared runtime, M00→M11 artifacts, UNKNOWN reconciliation, backup/restore, restart/replay and durable STOP

python3 scripts/smoke_br18b_backup_restore.py
BR-18b PASS: runtime-created M10 graph, M11 fixture evaluation/cycle, and UNKNOWN-to-human-reconciliation chain use a typed v3 manifest; checksum, exact inventory, lease-window activation, activation-bound health, broken evaluation/cycle links, reversed cycle time, restart, and durable STOP are verified
```

Audit and negative regression suite:

```text
python3 scripts/audit_readiness.py
READINESS AUDIT: NOT_READY_FOR_PRODUCTION

python3 -m unittest scripts.tests.test_audit_readiness -v
Ran 88 tests ...
OK

git diff --check
PASS
```

Go module checks:

```text
for module in contracts core lab/affiliate-bot lab/mission-runtime; do
  (cd "$module" && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) || exit 1
done
PASS
```

The learner Bot package and all four module test/vet commands passed. The
backup/restore run created its artifacts through the real CLI, restored into a
fresh target, replayed the graph and exercised post-restore rejection; it was
not a JSON byte-copy check.

## Boundary

This record confirms only local synthetic/fixture runtime behavior and audit
consistency. It does not provide ACCESSTRADE/provider requests, Cockpit
credential evidence, live executor authorization, payout/business outcome,
clean-machine self-service pilot, target-host deployment/recovery,
multi-host locking, or crash/power-loss atomicity. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
