# Evidence: shared M11 historical-chain validator — 2026-09-16

## Scope

The mission-runtime `m11-chain-check` path keeps strict contract/schema
decoding at its boundary and delegates the cross-artifact audit to
`core/m11.ValidateHistoricalChain`. The shared implementation covers both
`closed_cycle` and `resolved_stop`: exact lineage, machine `EffectRef`, scope,
timestamp ordering, health freshness, ledger snapshots/transitions, replay and
budget guards, durable STOP/reconciliation, and reviewed improvement links.
The canonical closed-cycle adapter also delegates to
`core/m11.ValidateClosedCycle`, so the harness does not own a second canonical
cross-artifact implementation.

This is a non-authorizing offline fixture audit. It does not call an executor,
reserve budget, contact a provider, establish a business outcome, or prove
crash/power-loss or multi-host behavior. Readiness remains
`NOT_READY_FOR_PRODUCTION`.

## Implementation

- `core/m11/historical_chain.go` defines the typed historical input and shared
  validators.
- `lab/mission-runtime/cmd/demo/m11_chain.go` decodes untrusted files, converts
  them to core types, and calls the shared validator.
- `lab/mission-runtime/cmd/demo/m11_effect_ref.go` reuses the shared
  machine-effect closed-cycle validator for the canonical adapter boundary.

## Commands and results

Executed from the repository checkout on `main` `f2c01a1` plus this change:

```text
(cd contracts && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
(cd core && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
(cd lab/affiliate-bot && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
(cd lab/mission-runtime && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
python3 scripts/validate_repo.py PASS
python3 scripts/validate_missions.py PASS
python3 scripts/validate_agent_semantics.py PASS
python3 scripts/validate_semantic_contracts.py PASS
python3 scripts/audit_readiness.py PASS (NOT_READY_FOR_PRODUCTION)
python3 -m unittest scripts.tests.test_audit_readiness -q PASS (88 tests)
python3 scripts/smoke_br16a_offline.py PASS
python3 scripts/smoke_br18b_backup_restore.py PASS
git diff --check PASS
```

The mission M11 regressions cover valid closed-cycle/resolved-stop fixtures and
reject forged links, invalid EffectRefs, scope changes, stale/degraded health,
ledger snapshot/transition drift, replay, budget overflow, malformed profiles,
and invalid STOP/review transitions. The existing learner smoke independently
confirms the real Bot M00–M11 and backup/restore paths remain passing.

## Remaining evidence

The shared validator closes only the implementation-sharing gap. Provider and
ACCESSTRADE operation, authenticated/live execution, business outcomes,
clean-machine pilot, target-host deployment/recovery, distributed locking and
power-loss guarantees remain open in the readiness matrix.
