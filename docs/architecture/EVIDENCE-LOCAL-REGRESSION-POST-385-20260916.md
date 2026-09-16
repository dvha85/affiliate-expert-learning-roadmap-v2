# Evidence: local regression after PR #385

## Scope

- `main` is at `4942ffaf0ef8c7a76ba9edafb6654670449d5521` after PR #385.
- This record covers the learner Bot and repository regression suites after the
  M11 expiry authority no-mutation test was merged.
- The result is fixture/read-only local evidence. It does not prove provider,
  business outcome, live executor, clean-machine pilot, target deployment,
  distributed locking or power-loss/atomic multi-file durability.

## Verification

```text
cd lab/affiliate-bot
GOWORK=off go test ./...
PASS

GOWORK=off go vet ./...
PASS

cd ../..
python3 -m unittest discover -s scripts/tests -p 'test_*.py'
Ran 113 tests: PASS

python3 scripts/audit_readiness.py
READINESS AUDIT: NOT_READY_FOR_PRODUCTION

python3 scripts/smoke_br18b_backup_restore.py
PASS
```

The post-merge audit resolves 68 scoped claims and confirms the required
M00-M11 regressions remain wired to CI. The expiry-specific regression also
passed five repetitions before merge and all ten PR checks passed.

## Boundary

The repository remains `NOT_READY_FOR_PRODUCTION`. External provider/source
operation, business outcome capture, live execution, pilot and deployment
evidence remain intentionally open.
