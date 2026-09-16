# Local regression after PR #387 — 2026-09-16

## Baseline and scope

- Merged `main`: `58160fa1afee9652efc52fbdaddfc770ea922cc2` after [PR #387](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/387).
- PR #387 checks: 10/10 successful.
- Scope is synthetic, fixture/read-only local verification. It does not prove
  provider traffic, live executor behavior, business outcome, clean-machine
  self-service, target-host deployment, distributed locking, or power-loss /
  atomic multi-file durability.

## Commands and results

```text
for module in contracts core lab/affiliate-bot lab/mission-runtime; do
  (cd "$module" && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)
done: PASS
python3 -m unittest discover -s scripts/tests -p 'test_*.py': 113 PASS
python3 scripts/smoke_br16a_offline.py --workspace <empty-temp-dir>: PASS
python3 scripts/smoke_br18b_backup_restore.py: PASS
python3 scripts/audit_readiness.py: NOT_READY_FOR_PRODUCTION
python3 -m unittest scripts.tests.test_audit_readiness -q: 98 PASS
git diff --check: PASS
```

The BR-18b smoke includes checksum-valid M11 authorization/execution lineage
mutations and confirms `GRAPH_FAILED` without publishing a restore target. The
audit resolves 70 scoped claims and intentionally keeps the repository
`NOT_READY_FOR_PRODUCTION`.

## Re-run on main after PR #388

- Current merged `main`: `8baae92352d9165affffd45f6dfb85e7ebdfd7b9`.
- The same four Go module test/vet suites, 113 Python tests, BR-16a, BR-18b,
  readiness audit, 98 audit tests and `git diff --check` were run again locally.
- All rerun commands passed. The audit reported `NOT_READY_FOR_PRODUCTION`
  and resolved 70 scoped claims.
- This rerun remains synthetic, fixture/read-only evidence; it does not add
  provider traffic, live executor, business outcome, clean-machine pilot,
  target-host deployment, distributed-lock or power-loss / atomic multi-file
  durability proof.

## CI evidence

PR #387 ran both required workflows and all 10 checks passed:

- [Curriculum CI run](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35070142379)
- [Mission Agent Path CI run](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35070142539)

These checks validate the repository's fixture and offline regression paths;
they are not external deployment or business evidence.

## Re-run on main after PR #389

- Merged `main` before this record: `fb201c24243dd7bd901849e2bb3cdab15d2b3595`.
- Four Go module test/vet suites, 113 Python tests, BR-16a, BR-18b, readiness
  audit, 98 audit tests and `git diff --check` all passed.
- Disposable n8n `2.38.1` with Node `24.21.0` passed the M06/M07 engine and
  M06 Schedule Trigger regressions: one canonical append, exact-deduplicated
  retries across n8n/adapter restart, and fail-closed adapter-unavailable
  behavior.
- The audit remained `NOT_READY_FOR_PRODUCTION` and resolved 71 scoped claims.

This remains local synthetic fixture/read-only evidence. It does not prove
provider traffic, business outcome, live executor behavior, clean-machine
self-service, target-host deployment, distributed locking, or power-loss /
atomic multi-file durability. The standalone command record is also captured
in `EVIDENCE-LOCAL-REGRESSION-POST-389-20260916.md`.
