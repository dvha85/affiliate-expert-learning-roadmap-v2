# Backup/restore SIGKILL regression — 2026-09-16

## Scope

Đây là regression cục bộ chạy learner Bot thật trong process con. Child bị
`SIGKILL` trong lúc backup hoặc restore còn ở private staging; parent xác nhận
process kết thúc do tín hiệu, target chưa xuất hiện, rồi dùng Bot thật retry
để tạo/restore snapshot hợp lệ.

Run chỉ chứng minh kernel advisory lock được nhả sau process termination trên
POSIX và không có target dở dang được publish trong seam này. Nó không phải
power-loss proof, không chứng minh atomic multi-file, orphan-staging cleanup,
Windows native lock, multi-host, provider, business outcome, live executor,
pilot hay deployment.

## Environment and result

- `repo_commit_before_change`: `814b081` (`main`)
- `host`: local macOS arm64
- `test`: `TestBackupProcessKillBeforePublishLeavesNoTargetAndRetrySucceeds`
- `result`: PASS (`0.13s` targeted run)

## Checks

| Check | Result |
|---|---|
| Child enters real `backup create` staging and receives `SIGKILL` | PASS; parent observes a signaled process |
| Backup target remains absent after termination | PASS |
| Fresh learner Bot backup retry | PASS; `BACKED_UP` |
| Child enters real `backup restore` staging and receives `SIGKILL` | PASS; parent observes a signaled process |
| Restore target remains absent after termination | PASS |
| Fresh learner Bot restore retry | PASS; `RESTORED` |
| Existing `os.Exit` process-boundary regression | PASS |

The same implementation is exercised by `go test ./cmd/bot`; the CI workflow
continues to run the full learner Bot race suite and the deterministic smoke
shard. The test intentionally remains POSIX-only because the repository's
Windows fallback uses cooperative stale-lock recovery rather than claiming a
native process-lock proof.
