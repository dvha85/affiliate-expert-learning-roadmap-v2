# RP-01 Windows runtime CI — 2026-09-18

## Scope

PR #415 adds a real Windows runner to Curriculum CI for the learner Bot's
filesystem boundaries. The `windows-runtime` job runs on `windows-latest` from
`lab/affiliate-bot` and exercises the native Windows managed lock, stable
append/read paths, backup/restore staging, and the full learner test package.

## Verification

Final head: `dad7d4c` (`fix: validate direct ACCESSTRADE outcome paths`).

- [Curriculum CI run 35302232344](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35302232344): overall `Success`.
- [windows-runtime job 105467080110](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35302232344/job/105467080110): `succeeded`, 2m48s.
- `go test ./...`: PASS.
- `go vet ./...`: PASS.
- `go test -count=1 ./cmd/bot -run "Test(ManagedPathLock|RuntimeGate|Backup|Restore)"`: PASS.

The new gate first exposed real runtime defects before the final green run:

- run `35301234630`: Windows directory sync and stable-reader sharing failed;
- run `35301582663`: Windows append wrote from offset zero and corrupted JSONL
  history;
- run `35301922153`: direct `outcomes.jsonl` validation accepted an incomplete
  record;
- run `35302232344`: all Curriculum CI jobs passed after the fixes in
  `822cd83`, `f043d3f` and `dad7d4c`.

## Readiness boundary

This is verified hosted-Windows, single-machine runtime evidence for the local
lock and backup/restore filesystem paths. It does not prove arbitrary
non-cooperating writer races, full ancestor-race parity, distributed or
multi-host locking, crash/power-loss durability, atomicity across multiple
files, provider operation, live execution, business outcomes, a clean-machine
pilot, or target-host deployment recovery. RP-01 remains `PARTIAL` and the
overall readiness boundary remains `NOT_READY_FOR_PRODUCTION`.
