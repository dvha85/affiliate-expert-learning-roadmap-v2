# RP-08 M11 outcome-link graph mutation evidence

## Scope

This bounded offline/read-only proof removes only the core M11 graph guard that
requires a checksum-valid `PRODUCTION_LEDGER.OutcomeLinks` `OutcomeID` to match
the offline evaluation reached through the same execution. The disposable copy
then runs the real
`TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks` regression and
requires failure at the swapped outcome-ID assertion.

This does not establish ledger durability, crash or power-loss recovery,
atomic multi-file publication, distributed or multi-host safety, Windows
traversal parity, provider/live execution, deployment, pilot acceptance,
business outcomes, or production readiness.

## Verification

- Implementation branch: `codex/rp08-m11-outcome-link-mutation`.
- Runner: `scripts/mutate_m11_outcome_link_graph.py`.
- Core guard: `core/m11/artifact_registry.go`.
- Regression: `core/m11/artifact_test.go` / `TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks`.
- Exact-head hosted CI verification: pending until the implementation PR is opened.
- Readiness boundary: `NOT_READY_FOR_PRODUCTION`.

