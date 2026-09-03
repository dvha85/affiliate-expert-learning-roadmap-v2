# Affiliate Bot v0.1 — M01 deterministic learner runtime

Workspace duy nhất của **M01 — Smallest Deterministic Bot v0.1**.

## Chạy

```bash
cd lab/affiliate-bot
go test ./...
go vet ./...
go run ./cmd/bot
```

`go test` đọc trực tiếp executable eval pack tại `../../evals/M01-deterministic-bot/cases.json`.

Fixture mặc định là `synthetic`; nó chỉ chứng minh behavior kỹ thuật.

## Contract

```text
canonical observations + formula_version
→ deterministic formula + stable tie-break
→ HUMAN_REVIEW > GET_MORE_DATA > RANK_SCENARIO
→ reason + missing/invalid evidence
→ NO external action
```

Baseline cố ý yếu: `price × commission_rate`. Nó chưa xét conversion potential, audience fit, refund/cancel risk, competition, compliance và nhiều yếu tố business khác.

```text
0 != missing
real evidence != RECOMMEND
RANK_SCENARIO != Approval != Execution
```

M00 evidence thật cung cấp reality boundary; M01 không fabricate field và không biến provenance thật thành recommendation.
