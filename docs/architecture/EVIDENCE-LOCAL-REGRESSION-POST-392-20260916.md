# Local regression after PR #392

## Environment and result

The regression was run after PR #392 was squash-merged into `main` at
`74c1719ee6fc90014a9e6599ed0837160e8de6a5`. It used the repository's local
fixture/read-only paths. No provider, live executor, business outcome, clean
machine pilot or target-host deployment was used.

All commands below passed:

## Commands and results

- `cd core && GOWORK=off go test ./... && GOWORK=off go vet ./...`
- `cd contracts && GOWORK=off go test ./... && GOWORK=off go vet ./...`
- `cd lab/mission-runtime && GOWORK=off go test ./... && GOWORK=off go vet ./...`
- `cd lab/affiliate-bot && GOWORK=off go test ./... && GOWORK=off go vet ./...`
- `python3 -m unittest discover -s scripts/tests` — 114 tests, PASS
- `python3 scripts/smoke_br16a_offline.py` — BR-16a PASS
- `python3 scripts/smoke_br18b_backup_restore.py` — BR-18b PASS
- `python3 scripts/audit_readiness.py .` — `NOT_READY_FOR_PRODUCTION`, 72
  scoped evidence claims resolved

The M11 graph test included in the merged commit rejects an authorization that
switches to individually checksum-valid health and cost artifacts not evaluated
by its gate. The result closes only that bounded local lineage seam; remaining
crash/power-loss, atomic multi-file, distributed-locking, provider, business,
pilot and deployment evidence remains open.
