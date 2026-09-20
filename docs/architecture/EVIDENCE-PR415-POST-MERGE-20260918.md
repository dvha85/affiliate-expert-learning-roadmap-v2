# PR #415 post-merge evidence

## Scope

PR #415 (`ci: run learner bot runtime tests on Windows`) was squash-merged into
`main` at `707ac76aba0c1aac705ec3e52bcbd45e71edfb85`. The merge adds the hosted
Windows runtime gate and the platform-specific filesystem fixes recorded in
[the RP-01 Windows runtime evidence](EVIDENCE-RP01-WINDOWS-RUNTIME-CI-20260918.md).

This is bounded synthetic/read-only repository evidence. It does not prove
provider operation, live executor authority, business outcome, clean-machine
pilot, target-host deployment, distributed locking, power-loss durability or
atomic multi-file recovery. Readiness remains
`NOT_READY_FOR_PRODUCTION`.

## Verification

The merged `main` head completed both required push workflows successfully:

- [Curriculum CI run 35303587160](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35303587160): 10/10 jobs passed, including [the hosted `windows-runtime` job 105471115956](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35303587160/job/105471115956).
- [Mission Agent Path CI run 35303587091](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35303587091): 3/3 jobs passed.

The synchronized local checkout independently passed:

- `GOWORK=off go test -race ./...` in `lab/affiliate-bot`;
- `go test` and `go vet` for `contracts`, `core`, `lab/affiliate-bot` and
  `lab/mission-runtime`;
- `python -m unittest discover -s scripts/tests -v` (133 tests);
- `python scripts/smoke_br16a_offline.py` and
  `python scripts/smoke_br18b_backup_restore.py`;
- `python scripts/audit_readiness.py`, which still reports
  `NOT_READY_FOR_PRODUCTION`.

## Readiness boundary

The post-merge run confirms the Windows single-host runtime gate is present on
the current `main` head and that the offline regression suite remains green.
It does not close RP-01's arbitrary non-cooperating-writer or full
ancestor-race parity gaps, and it adds no provider, live-execution, business,
pilot or deployment evidence. RP-01 remains `PARTIAL` and overall readiness
remains `NOT_READY_FOR_PRODUCTION`.
