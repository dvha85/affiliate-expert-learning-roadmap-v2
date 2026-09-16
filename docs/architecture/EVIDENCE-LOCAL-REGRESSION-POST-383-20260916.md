# Evidence: local regression after PR #383

## Scope

- Merged `main` after PR #383 at `c910185c435cbbef576208047487ce1394ec9337`.
- This record covers local fixture/read-only regression only. It does not prove
  provider operation, ACCESSTRADE capture, live executor behavior, business
  outcome, clean-machine pilot, target deployment, distributed locking, or
  power-loss/atomic multi-file durability.

## Verification

```text
cd lab/affiliate-bot
GOWORK=off go test -count=1 ./...
GOWORK=off go vet ./...
PASS

python3 -m unittest discover -s scripts/tests -v
Ran 113 tests ... OK

python3 scripts/smoke_br18b_backup_restore.py
BR-18b PASS

python3 scripts/audit_readiness.py
READINESS AUDIT: NOT_READY_FOR_PRODUCTION
evidence graph resolves 68 scoped claims
```

The merged PR's 10 GitHub Actions checks also passed, including learner-bot
race and both learner test shards. The new M11 exact-expiry/no-mutation
regression passed five local repetitions before merge.

## Boundary

This confirms the post-merge local regression baseline and keeps the public
readiness boundary at `NOT_READY_FOR_PRODUCTION`. Remaining external and
operated gaps stay open.
