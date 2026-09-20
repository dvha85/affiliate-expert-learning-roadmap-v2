# F-09 offline verification entrypoint — 2026-09-20

Status: `PARTIAL`, bounded offline/read-only runbook evidence.

## Scope

`python scripts/run_offline_checks.py` is the version-controlled entrypoint
for the complete local offline contract. It runs the four Go module test/vet
suites, learner race regression, Python regression, static validators, offline
smokes, the readiness audit and `git diff --check` in a fixed order. It
explicitly excludes operated validators that require an execution artifact and
the real n8n engine/Schedule Trigger regressions that require a pinned Node/n8n
runtime.

## Verification

- `python -m unittest scripts.tests.test_run_offline_checks -v`: 3/3 PASS.
- `python scripts/run_offline_checks.py --list`: PASS; prints 37 ordered
  offline steps and the two excluded operated validators plus the two real n8n
  runners.
- `python -m unittest discover -s scripts/tests -q`: 174 tests PASS.
- `python -m compileall -q scripts`: PASS.
- JSON parsing for `READINESS-MATRIX.json` and
  `READINESS-EVIDENCE-GRAPH.json`: PASS; the graph resolves 154 scoped claims.
- `python scripts/run_offline_checks.py`: blocked before any check started
  because this host has no `go` executable. The runner returned its explicit
  missing-tool diagnostic; this is not a partial test result.

## Boundaries and next action

This record does not claim hosted Go/Windows execution, real n8n execution,
operated execution artifacts, provider/live execution, deployment recovery,
beginner pilot, business outcome, distributed locking, power-loss or
production readiness. Run the same entrypoint in hosted CI or on a machine
with the declared Go/Python/Git toolchain before recording full offline PASS.
The repository remains `NOT_READY_FOR_PRODUCTION`.
