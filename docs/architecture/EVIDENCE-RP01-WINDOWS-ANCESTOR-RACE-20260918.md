# RP-01 Windows ancestor-race hardening — 2026-09-18

## Scope

PR #416 hardens the remaining Windows pathname readers and writers at head
`bbd8aec14e56ba0b484376014b2ab5ec775ce3ec`:

- `output_parent_windows.go` pins the output-parent directory chain before
  backup/restore writes; missing components are created one at a time and
  pinned before traversal continues.
- `runtime_gate_windows.go` pins existing parents before the managed-lock
  handle and rejects reparse-point ancestors before the final file open.
- `internal/store/stable_read_windows.go` uses the same native parent-handle
  boundary for JSONL readers.
- Native directory handles use `FILE_FLAG_OPEN_REPARSE_POINT` and omit delete
  sharing, so an ancestor replacement is blocked or rejected before the final
  pathname open. The final file itself is still opened by pathname after the
  parent chain has been pinned.

## Verification

- [PR #416](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/416)
  is open from the clean branch `codex/rp01-windows-ancestor-race-v2`.
- [Curriculum CI run 35306191800](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35306191800): overall `Success` on the evidence-bearing head.
- [windows-runtime job 105478750919](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35306191800/job/105478750919): `succeeded`.
- `go test ./...`: PASS on `windows-latest`.
- `go vet ./...`: PASS on `windows-latest`.
- `go test -count=1 ./cmd/bot -run "Test(ManagedPathLock|RuntimeGate|Backup|Restore)"`: PASS on `windows-latest`.
- [learner-bot-race job 105478751048](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35306191800/job/105478751048): `succeeded`; `go test -race ./...` PASS.
- [Mission Agent Path CI run 35306191794](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35306191794): all three jobs PASS.

The Windows full package run includes these controlled post-preflight
ancestor-replacement regressions:

- `TestWindowsBackupRestoreRejectsAncestorJunctionAfterPreflight`
- `TestWindowsStableReaderRejectsAncestorJunctionAfterPreflight`
- `TestWindowsStableAppendRejectsAncestorJunctionAfterPreflight`
- `TestWindowsJSONLReaderRejectsAncestorJunctionAfterPreflight`

Each test replaces the preflighted directory with a Windows junction to an
external directory and requires the operation to fail closed without changing
the external sentinel/tree.

## Conservative boundary

This evidence verifies the bounded Windows single-host seam for an ancestor
replacement after pathname preflight: native parent handles are pinned with
reparse traversal disabled, and the regression fails before the external tree
is read or written. Windows does not expose a portable `openat`/`mkdirat`
equivalent for arbitrary multi-component creation in this code path, so this
implementation deliberately does not claim POSIX traversal parity. It also
does not prove distributed or multi-host locking, crash/power-loss durability,
atomicity across multiple files, provider operation, live execution, business
outcomes, pilot readiness, or target-host deployment recovery.

RP-01 remains `PARTIAL`; overall readiness remains
`NOT_READY_FOR_PRODUCTION`. RP-02 shared M08 decoder/policy and provider/live
execution remain out of scope.
