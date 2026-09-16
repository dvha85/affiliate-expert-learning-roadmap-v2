# Local regression after PR #389 — 2026-09-16

## Baseline and scope

- Merged `main` before this record: `fb201c24243dd7bd901849e2bb3cdab15d2b3595`.
- This is maintainer-run synthetic, fixture/read-only local verification. It
  does not prove provider traffic, live executor behavior, business outcome,
  clean-machine self-service, target-host deployment, distributed locking, or
  power-loss / atomic multi-file durability.
- The n8n checks use disposable n8n `2.38.1`, Node `24.21.0`, loopback
  adapters, and an in-process OpenAI-compatible model stub. No ACCESSTRADE
  request, credential, publish, or live provider call was made.

## Commands and results

```text
for module in contracts core lab/affiliate-bot lab/mission-runtime; do
  (cd "$module" && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)
done: PASS
python3 -m unittest discover -s scripts/tests -p 'test_*.py' -q: 113 PASS
python3 scripts/smoke_br16a_offline.py --workspace <empty-temp-dir>: PASS
python3 scripts/smoke_br18b_backup_restore.py: PASS
python3 scripts/audit_readiness.py: NOT_READY_FOR_PRODUCTION
python3 -m unittest scripts.tests.test_audit_readiness -q: 98 PASS
python3 scripts/run_n8n_engine_regression.py \
  --n8n-node /Users/hadinh/.local/node-v24.21.0-darwin-arm64/bin/node \
  --n8n-cli /Users/hadinh/.local/n8n-2.38.1/node_modules/n8n/bin/n8n: PASS
python3 scripts/run_n8n_m06_schedule_regression.py \
  --n8n-node /Users/hadinh/.local/node-v24.21.0-darwin-arm64/bin/node \
  --n8n-cli /Users/hadinh/.local/n8n-2.38.1/node_modules/n8n/bin/n8n: PASS
git diff --check: PASS
```

The disposable n8n engine regression passed M06/M07 synthetic and sanitized
selected-source replay, strict policy rejection, grounded Agent output and
forged-commission rejection. The Schedule Trigger regression appended one
canonical record, exact-deduplicated retries before and after n8n/adapter
restart, and failed closed when the loopback adapter was unavailable.

The readiness audit resolved 71 scoped claims and intentionally keeps the
repository `NOT_READY_FOR_PRODUCTION`. The run records only local evidence;
provider, business outcome, live executor, pilot, deployment and stronger
crash/distributed durability claims remain open.
