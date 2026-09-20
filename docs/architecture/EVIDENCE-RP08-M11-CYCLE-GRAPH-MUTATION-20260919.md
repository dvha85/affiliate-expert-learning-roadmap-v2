# RP-08 M11 cycle graph mutation proof

## Scope

The disposable mutation in `scripts/mutate_m11_cycle_graph.py` removes only
the core `ProductionCycleRecord` graph guard that binds a closed M11 cycle to
its exact evaluation outcome, execution, authorization, gate and lease
lineage. The real `core/m11` lifecycle regression must then fail at the
mismatched evaluation-outcome assertion.

This is offline, synthetic, fixture-backed mutation evidence. It does not
prove crash or power-loss durability, atomic multi-file publication,
distributed or multi-host locking, Windows traversal parity, provider/live
execution, deployment, pilot acceptance, business outcomes or production
readiness.

## Verification

- Exact-head hosted CI is required to execute the mutation because this
  workspace does not have `go.exe` installed.
- Expected failure marker: `cycle with mismatched evaluation outcome was
  accepted`.
- The implementation PR adds the mutation to the required
  `deterministic-smokes-backup-mutations` workflow job and the isolated
  readiness-audit fixture.
