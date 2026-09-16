# Evidence: local regression after PR #379

## Scope

- Checkout: `main` at `80a895ed69d339b7c66e69e70466aac52ded6783`.
- This is maintainer-run local fixture/read-only verification after the M11
  source-canary-grant lineage change.
- It does not prove ACCESSTRADE/provider traffic, business outcome, live
  executor, clean-machine pilot, target-host deployment, distributed locking,
  or power-loss/atomic multi-file durability.

## Results

```text
cd lab/affiliate-bot && GOWORK=off go test -count=1 ./...
PASS

cd lab/affiliate-bot && GOWORK=off go vet ./...
PASS

python3 -m unittest discover -s scripts/tests
Ran 113 tests ... OK

python3 scripts/smoke_br18b_backup_restore.py
BR-18b PASS

python3 scripts/audit_readiness.py
READINESS AUDIT: NOT_READY_FOR_PRODUCTION
```

The audit resolved 66 scoped evidence claims and confirmed that required
M00-M11 regressions remain wired to CI. The readiness boundary remains
`NOT_READY_FOR_PRODUCTION`.
