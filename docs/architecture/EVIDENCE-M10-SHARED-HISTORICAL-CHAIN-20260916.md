# Evidence: shared M10 historical-chain validator — 2026-09-16

## Scope

The mission-runtime `m10-chain-check` path now performs strict input decoding
at its boundary and delegates the historical canary-chain decision to
`core/m10.ValidateHistoricalChain`. The shared implementation owns the
cross-artifact links, canary scope, timestamp ordering, pre-gate ledger
snapshot, blocked/duplicate checks and budget arithmetic. The harness no
longer contains a second copy of that cross-artifact validator.

This is a non-authorizing offline fixture audit. It does not call an executor,
reserve budget, contact a provider, establish a business outcome, or prove
crash/power-loss or multi-host behavior. Readiness remains
`NOT_READY_FOR_PRODUCTION`.

## Implementation

- `core/m10/historical_chain.go` defines the typed historical approval/ledger
  envelopes and `ValidateHistoricalChain`.
- `lab/mission-runtime/cmd/demo/m10_chain.go` keeps file/contract decoding at
  the harness boundary, converts the decoded values to core types, and calls
  the shared validator.
- The existing learner Bot uses the canonical `core/m10` artifact builders,
  decoders and graph registry; this change does not grant live execution.

## Commands and results

Executed from the repository checkout:

```text
(cd contracts && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)  PASS
(cd core && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)       PASS
(cd lab/affiliate-bot && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
(cd lab/mission-runtime && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
python3 scripts/validate_repo.py                                                   PASS
python3 scripts/validate_missions.py                                              PASS
python3 scripts/validate_agent_semantics.py                                       PASS
python3 scripts/validate_semantic_contracts.py                                    PASS
python3 scripts/audit_readiness.py                                                PASS (NOT_READY_FOR_PRODUCTION)
python3 -m unittest scripts.tests.test_audit_readiness -q                             PASS (88 tests)
git diff --check                                                                  PASS
```

The mission M10 chain regression covers valid RISK0/RISK1 and window-reset
chains plus rejection of broken links, scope, expiry, ledger snapshots,
blocked ledgers, schema/profile mismatches and near-`int64` cost overflow.

## Remaining evidence

The shared validator closes only the implementation-sharing gap. Provider and
ACCESSTRADE operation, authenticated/live execution, business outcomes,
clean-machine pilot, target-host deployment/recovery, distributed locking and
power-loss guarantees remain open in the readiness matrix.
