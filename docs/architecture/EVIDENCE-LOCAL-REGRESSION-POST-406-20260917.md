# Evidence — local regression after PR #406

- Main head: `7eacb661d2083d97b42ea31207a3e677705d7fea` (PR #406).
- Scope: local synthetic/read-only fixture verification on macOS arm64.
- Readiness remains `NOT_READY_FOR_PRODUCTION`.

## Commands and results

The following passed on the merged head:

- `GOWORK=off go test -count=1 ./...` and `GOWORK=off go vet ./...` in each
  of `contracts`, `core`, `lab/affiliate-bot` and `lab/mission-runtime`.
- `python3 -m unittest discover -s scripts/tests -v`: 124 tests passed.
- Argument-free static validators that are runnable standalone.
- `python3 scripts/smoke_br12d.py`.
- `python3 scripts/smoke_br13b.py`.
- `python3 scripts/smoke_br16a_offline.py`.
- `python3 scripts/smoke_br18b_backup_restore.py`.
- `python3 scripts/audit_readiness.py`: `NOT_READY_FOR_PRODUCTION` as expected.
- `git diff --check`.

The standalone M06/M07 operated validators require an execution JSON and
runtime store produced by an operated run, so they were not invoked without
those artifacts. Node.js and n8n are not installed in this worktree; no local
n8n engine PASS is claimed. The PR's remote checks are tracked separately from
this local record.

This evidence verifies the offline learner/runtime and audit paths only. It
does not prove provider operation, live execution, business outcome, clean
machine pilot, target-host deployment, distributed locking, or
power-loss/atomic multi-file durability.
