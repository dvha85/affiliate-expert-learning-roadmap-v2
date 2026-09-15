# Deterministic runtime single-build re-run (2026-09-16)

Trạng thái: **local automated verification / fixture-only / read-only**.

## Run record

- `run_id`: `deterministic-runtime-single-build-20260916`
- `repo_commit`: `4461e69` (`main`, sau khi merge PR #344)
- `host`: local Darwin arm64
- `scope`: workflow-equivalent deterministic-runtime checks for the learner Bot
- `result`: all commands listed below PASS

## Commands and results

The learner Bot was compiled once into a fresh temporary directory and that
same binary was used for the M01 baseline plus M02 capture, list, replay and
decision smoke commands.

```text
go test ./internal/...                              PASS (0.10s)
cd ../../contracts && go test ./... && go vet ./... PASS (0.18s)
go vet ./...                                        PASS (0.26s)
go build -o <tmp>/learner-bot ./cmd/bot              PASS
<tmp>/learner-bot                                    PASS (M01 baseline)
<tmp>/learner-bot history capture ...               PASS
<tmp>/learner-bot history list ...                  PASS
<tmp>/learner-bot history replay ...                PASS (replay=MATCH)
<tmp>/learner-bot history decision ...              PASS (action=null)
python3 scripts/audit_readiness.py                  PASS
```

The audit result was:

```text
READINESS AUDIT: NOT_READY_FOR_PRODUCTION
- structured partial/open criteria: BR-13, BR-14, BR-15, BR-16a, BR-17, BR-18b, BR-19
```

## Boundary

This record verifies the CI efficiency path and unchanged M01/M02 smoke
coverage on the merged `main` revision. It does not measure GitHub runner
wall-clock improvement, and it does not provide ACCESSTRADE/provider,
business-outcome, live-executor, clean-machine pilot, target-host deployment,
multi-host or crash/power-loss evidence. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
