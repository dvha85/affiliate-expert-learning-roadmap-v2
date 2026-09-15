# Evidence: learner M08/M09 schema identity

Date: 2026-09-16  
Scope: local offline learner Bot source and tests only

## Change

`lab/affiliate-bot/cmd/bot/mission_command.go` now aliases the learner
mission-state fields to the canonical `core/m08.Intent`,
`core/m08.PolicyDecision` and `core/m09.ApprovalRecord` types. The learner
keeps its own mission-state envelope and M10/M11 runtime state, but no longer
has a second local declaration for these M08/M09 artifacts.

## Verification

From `lab/affiliate-bot`:

```text
$ GOWORK=off go test ./cmd/bot -run '^TestLearnerMissionStateUsesCanonicalM08M09Types$' -count=1 -v
--- PASS: TestLearnerMissionStateUsesCanonicalM08M09Types (0.00s)
PASS

$ GOWORK=off go vet ./cmd/bot
PASS (no diagnostics)

$ GOWORK=off go test ./...
PASS (cmd/bot 33.771s; internal/app, internal/learning and internal/store PASS)
```

From the repository root:

```text
$ python3 scripts/audit_readiness.py
READINESS AUDIT: NOT_READY_FOR_PRODUCTION
... evidence graph resolves 51 scoped claims ...

$ python3 -m unittest scripts.tests.test_audit_readiness -v
Ran 88 tests in 7.872s
OK
```

The assignments in `TestLearnerMissionStateUsesCanonicalM08M09Types` require
type identity. Reintroducing a local learner struct makes the test fail at
compile time, instead of allowing JSON-compatible but divergent schemas.

## Boundary

This closes only the offline schema-drift seam. It does not prove broader
authorization/execution migration, provider operation, live execution,
business outcome, crash/power-loss recovery, distributed locking, pilot,
deployment or production readiness. The repository remains
`NOT_READY_FOR_PRODUCTION`.
