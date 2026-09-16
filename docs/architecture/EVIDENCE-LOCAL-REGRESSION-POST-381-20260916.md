# Evidence: local regression after PR #381

## Scope

- Checkout: `main` at `8229ad3c8d3490a8780ea987c3a406a3f08cd96c`.
- Maintainer-run local fixture/read-only verification after the M11 recovery
  admission concurrency regression was merged.
- No provider traffic, business outcome, live executor, clean-machine pilot,
  target-host deployment, distributed/multi-host guarantee, or power-loss/
  atomic multi-file durability is claimed.

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

The audit resolved 67 scoped claims and confirmed the M00-M11 regressions are
wired to CI. Readiness remains `NOT_READY_FOR_PRODUCTION`.
