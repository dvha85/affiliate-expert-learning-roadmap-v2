# M11 authorization lineage mutation evidence

## Scope

This is a bounded RP-08 mutation proof for the offline M11 artifact graph. It
checks that the real graph regression fails when authorization-to-gate exact
lineage enforcement is removed from a disposable copy of `core/m11`.

It does not prove ledger durability, live execution, provider operation,
business outcome, crash/power-loss recovery, atomic multi-file commit,
distributed locking, pilot or deployment readiness.

## Verification

```text
python3 scripts/mutate_m11_authorization_lineage.py
M11 authorization-to-gate exact lineage mutation detected by the real graph regression
```

The runner removes the complete set of authorization-to-gate lease,
health-snapshot and cost-bound identity checks, runs
`TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks`, and requires
the forged-lineage assertion to be the failure. The source repository is not
mutated by the runner.
