# Evidence: M11 source canary-grant lineage

## Scope

- Base checkout before this change: `main` at `6ec07065796933d1b9a05019739832e4da72378a`.
- Scope: the local learner Bot's populated runtime and typed backup/restore graph.
- The check is read-only/fixture-based. It does not exercise a live executor,
  ACCESSTRADE/provider traffic, a business outcome, a clean-machine pilot,
  target-host deployment, power-loss recovery, or distributed locking.

## Implementation

- `lab/affiliate-bot/cmd/bot/m11_registry.go`
  - registering a `PRODUCTION_LEASE` in a populated learner runtime now binds
    its source canary grant ID, version and domain hash to the active canonical
    M10 grant;
  - the exact M10 canary-grant registry entry is resolved and decoded before
    the lease can be appended.
- `lab/affiliate-bot/cmd/bot/backup_command.go`
  - restore validation applies the same cross-store binding when a mission
    state with an active M10 grant is present;
  - a checksum-valid M11 lease/approval pair with an unrelated source grant is
    rejected before the restored target is published.

Minimal M11-only fixtures without `mission-state.json` remain supported because
there is no M10 grant to resolve. A populated learner runtime is fail-closed.

## Verification

The following commands were run from this checkout:

```text
cd lab/affiliate-bot
GOWORK=off go test -run 'TestM11LeaseSourceGrantMustResolveActiveM10Grant|TestMissionM11RegistryUsesCanonicalCoreDecoder|TestBackupRestore' -count=1 ./cmd/bot
PASS

GOWORK=off go test -count=1 ./...
PASS

GOWORK=off go vet ./...
PASS

python3 scripts/smoke_br18b_backup_restore.py
BR-18b PASS

python3 -m unittest discover -s scripts/tests -v
Ran 113 tests ... OK

python3 scripts/audit_readiness.py
READINESS AUDIT: NOT_READY_FOR_PRODUCTION
```

The BR-18b smoke creates a real runtime-generated backup, mutates the M11
lease's source grant hash while recomputing the lease, approval, entry and
manifest checksums, then restores into a fresh target. Restore returns
`GRAPH_FAILED` and does not publish the target. The positive path uses the
actual canonical M10 grant hash and still restores/replays successfully.

## Boundary

This closes one local cross-store lineage seam only. RP-03 and RP-07 remain
`PARTIAL`; external provider/business evidence, live execution, deployment,
clean-machine pilot, multi-host guarantees and power-loss/atomic multi-file
durability remain open. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
