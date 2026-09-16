# Backup/restore post-publish process-kill boundary — 2026-09-16

## Scope

The real backup and restore command paths are exercised in a separate child
process that is killed immediately after the final staging rename and before
the parent-directory acknowledgement. The visible target must be a complete,
verifiable snapshot/runtime, and a same-target retry must be rejected rather
than risking an overwrite.

## Result

- `repo_commit_before_change`: `231254a` (`main`)
- `host`: local macOS arm64
- `TestBackupProcessKillAfterPublishLeavesVisibleCompleteTarget`: PASS
- backup target after `SIGKILL`: visible and `verifyBackup` PASS
- restored runtime after `SIGKILL`: visible and `loadMissionState` PASS
- same-target backup and restore retries: `TARGET_NOT_EMPTY`

The test covers the real `os.Rename` boundary, not only an in-process error
return. It proves safe visible-target behavior for this local process-kill
seam; it does not prove that a parent-directory sync survives power loss,
atomicity across multiple files, filesystem-crash recovery, Windows native
locking, distributed/multi-host behavior, provider operation, business
outcomes, pilot or deployment readiness.
