# M11 reverse-link ledger graph evidence

## Scope

This record covers a bounded offline/read-only integrity seam in the canonical
M11 artifact registry. A ledger that publishes a reconciliation resolution ID
must resolve that ID back to the exact UNKNOWN execution and immutable lease,
and only a completed stopped head may carry the reviewed resolution. A ledger
outcome link must also retain the same `OutcomeID` as the offline evaluation
for that execution when the evaluation is present.

The decoder rejects empty or duplicate `ReconciliationResolutionIDs` before a
registry entry is created. The graph validator allows the append-only
resolution-only intermediate, then rejects a checksum-valid dangling,
cross-lease, cross-execution or wrong-outcome link during complete inventory
validation. Because M11 evaluations are deliberately offline fixture
evaluations, each evaluation must carry exactly one evidence ID and it must be
the evaluated fixture `OutcomeID`; schema validation rejects an empty array and
graph validation rejects a non-singleton or unrelated ID. The learner Bot
loader regression exercises that rejection before runtime use.

This does not prove ledger durability, power-loss recovery, atomic multi-file
commit, distributed locking, live execution, provider operation, business
outcome, pilot or deployment readiness. The repository remains
`NOT_READY_FOR_PRODUCTION`.

## Implementation

- `core/m11/artifact.go` rejects empty and duplicate resolution IDs in the
  canonical ledger decoder.
- `core/m11/artifact_registry.go` indexes resolutions by immutable ID and
  reverse-checks every ledger resolution link against execution, lease and the
  resolved STOPPED state; it also compares ledger `OutcomeID` links with an
  available evaluation and rejects an evaluation whose evidence is not exactly
  the single fixture outcome ID.
- `core/m11/artifact_test.go` keeps positive coverage for a valid reviewed
  STOPPED ledger and negative coverage for dangling resolution IDs, duplicate
  resolution IDs, swapped outcome IDs, unrelated evaluation evidence and
  non-singleton evaluation evidence.
- `lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go` rewrites disposable
  registries with checksum-valid orphan resolution and mismatched evaluation
  links and requires the real learner loader to reject them before runtime use.

## Verification

The Go acceptance command is wired to the Curriculum CI race suite:

```text
go test -race ./...
```

The current Windows worktree does not contain `go.exe`, so local Go test/vet
execution is unavailable here. Python audit/tests and JSON/diff checks are run
locally; the remote Go CI result is required before treating the runtime
regression as verified. No provider or executor is invoked.
