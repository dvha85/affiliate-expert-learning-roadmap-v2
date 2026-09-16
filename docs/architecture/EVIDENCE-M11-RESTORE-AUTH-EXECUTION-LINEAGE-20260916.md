# M11 restore authorization/execution lineage evidence — 2026-09-16

## Scope

This record covers a narrow local restore-graph regression on the learner Bot.
It does not claim live execution, provider traffic, business outcome, pilot,
deployment, multi-host locking, or power-loss durability.

The fixture is created by `scripts/smoke_br18b_backup_restore.py` through the
real learner Bot, then copied into checksum-valid backup variants. The variants
change only immutable M11 cross-artifact references and update the backup
manifest/content hash, so a digest-only check cannot explain the rejection.

## Cases

The smoke now runs these two negative cases against `backup restore`:

1. `PRODUCTION_EXECUTION_AUTHORIZATION.production_health_snapshot_hash` is
   replaced with a different hash while the referenced health snapshot remains
   present.
2. `PRODUCTION_EXECUTION_RECORD.production_health_snapshot_hash` is replaced
   with a different hash while its authorization and execution remain present.

For each case the real learner Bot returns `GRAPH_FAILED` and the empty restore
target is not published. The positive backup is still restored and replayed by
the same smoke, so the negative cases are not a parser-only test.

## Verification

On branch `codex/m11-restore-lineage-guards-20260916`, before PR creation:

```text
python3 scripts/smoke_br18b_backup_restore.py: PASS
```

The check uses a typed v3 manifest, runtime-created M00–M11 fixture artifacts,
the learner Bot backup/restore commands, and the canonical M11 registry graph
validator. The existing readiness boundary remains:
`NOT_READY_FOR_PRODUCTION`.

## Remaining boundary

This proves only that the offline restore path does not publish a runtime when
these authorization/execution references drift. It does not prove arbitrary
filesystem crash recovery, atomicity across multiple stores, distributed
locking, live executor behavior, ACCESSTRADE/provider evidence, business
outcome, clean-machine self-service, or target-host deployment.
