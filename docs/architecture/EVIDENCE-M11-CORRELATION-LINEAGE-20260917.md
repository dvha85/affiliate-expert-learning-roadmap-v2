# M11 correlation-lineage evidence

## Scope

This record covers a bounded offline/read-only integrity seam. The immutable
M11 production lease is the correlation root. A checksum-valid cost bound,
authorization, execution record or closed-cycle record from another
correlation lineage must not be made admissible by copying downstream IDs and
recomputing their local hashes.

The same invariant is enforced by the registry graph and by the shared
historical-chain reader. The learner loader therefore rejects a mixed
correlation registry before runtime use, while the mission-runtime adapter
exercises the same core validator for both chain profiles.

This does not prove ledger durability, power-loss recovery, atomic multi-file
commit, distributed locking, live execution, provider operation, business
outcome, pilot or deployment readiness. The repository remains
`NOT_READY_FOR_PRODUCTION`.

## Implementation

- `core/m11/artifact_registry.go` requires the trusted cost bound to retain the
  lease correlation and requires authorization, execution and cycle records to
  retain that same root.
- `core/m11/historical_chain.go` applies the correlation check to both the
  resolved-stop path and the shared closed-cycle validator.
- `lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go` rewrites a disposable
  registry with checksum-valid foreign-correlation artifacts and confirms the
  real loader rejects it.
- `lab/mission-runtime/cmd/demo/m11_chain_test.go` mutates execution
  correlation on the real chain-check path and requires `BROKEN_LINK`.

## Verification

```text
cd core
GOWORK=off go test ./m11
ok   github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11

cd ../lab/affiliate-bot
GOWORK=off go test ./cmd/bot
ok   github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/cmd/bot

cd ../mission-runtime
GOWORK=off go test ./cmd/demo
ok   missionlab/cmd/demo
```

The positive graph remains valid. Negative cases reject a foreign
cost-bound correlation, a foreign authorization correlation, and a mismatched
execution correlation; no provider or executor is invoked.
