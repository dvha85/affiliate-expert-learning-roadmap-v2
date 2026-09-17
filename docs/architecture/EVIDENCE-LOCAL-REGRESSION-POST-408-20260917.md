# Evidence — regression and CI baseline after PR #408

- Merged main head: `c6debee4b1ef82626eb3e22469886b1f55b45e54`.
- PR head before squash: `4360858b0d3c04620838fcd94cfd7df8f0f2e287`.
- Scope: repository-local synthetic/read-only verification and remote CI evidence.
- Readiness remains `NOT_READY_FOR_PRODUCTION`.

## Remote CI

All 12 PR check-runs for head `4360858b0d3c04620838fcd94cfd7df8f0f2e287`
completed successfully before merge. They were split across these workflow
runs:

- [curriculum CI run 35202587298](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35202587298):
  foundations, backup/mutation smokes, learner race/shards, deterministic
  runtime and quickstart.
- [mission-agent path CI run 35202587341](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35202587341):
  mission semantics/blueprints, mission runtime and n8n engine regression.

The remote checks provide the Go gate for RP-01's M10/M11 registry parent-swap
regressions and the existing offline/runtime CI coverage. They do not turn the
fixture path into provider, live-executor, business-outcome, clean-machine
pilot, target-host deployment, distributed-locking or power-loss/atomic
multi-file evidence.

## Local commands and results

The current Windows worktree has no Go executable, so no local Go test or vet
PASS is claimed. The following repository checks were run on the synced branch:

- `python scripts/audit_readiness.py`: PASS; `NOT_READY_FOR_PRODUCTION`.
- `python -m unittest discover -s scripts/tests -q`: PASS; 125 tests.
- JSON parse for `docs/plans/READINESS-MATRIX.json` and
  `docs/plans/READINESS-EVIDENCE-GRAPH.json`: PASS.
- `git diff --check`: PASS.

This record is a baseline/evidence sync for the merged PR. RP-10 remains
`OPEN` because selected environment/authority, provider operation, live
executor, clean-machine pilot and target-host deployment evidence are outside
this repository.
