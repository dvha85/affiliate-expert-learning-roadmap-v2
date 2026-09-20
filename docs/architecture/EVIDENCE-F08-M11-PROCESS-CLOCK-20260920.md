# Evidence F-08 — M11 process-clock authority

## Semantics

The learner M11 adapter separates historical provenance from authority time.
`missionNowUTC()` is process-owned and controls whether a lease, activation,
ledger, cost bound, health snapshot, authorization or reservation is usable now.
CLI timestamps remain inputs to chronology checks and deterministic artifact IDs;
they cannot backdate an authority decision or make a historical artifact active.
Historical chain validation continues to evaluate recorded timestamps without
consulting the current wall clock.

The implementation is in `lab/affiliate-bot/cmd/bot/m11_registry.go` on head
`35394a3`. It also guards activation and ledger initialization so an expired or
not-yet-started lease cannot be used to create a new authority state.

## Verification

`TestMissionM11ExpiryRejectsAuthorityWritesWithoutMutation` builds the real
learner fixture, checkpoints gate/authorization/reservation/execution state with
backup/restore, and runs a fresh Bot process at exact lease expiry. It repeats the
checks with timestamps backdated into the original lease window. All four
authority writes reject before mutation; the restored state, M10/M11 registries
and journal remain unchanged.

Local checks on this head:

- `go test -count=1 ./...` and `go vet ./...` pass in `contracts`, `core`,
  `lab/mission-runtime` and `lab/affiliate-bot`.
- `go test -race -count=1 ./...` passes in `lab/affiliate-bot`.
- `python3 -m unittest discover -s scripts/tests -v` passes 167 tests.
- `python3 -m py_compile scripts/run_n8n_m06_schedule_regression.py` passes.

This is offline/fixture/read-only evidence. Hosted CI, real n8n execution,
trusted external time, live executor, provider operation, deployment, pilot,
business outcomes, power-loss durability and distributed locking remain open.
