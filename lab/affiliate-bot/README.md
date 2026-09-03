# Affiliate Bot v0.1 — M01 deterministic learner runtime

Workspace này thuộc **M01 — Smallest Deterministic Bot v0.1**.

## Chạy

```bash
cd lab/affiliate-bot
go run ./cmd/bot
go test ./...
```

Fixture mặc định là `synthetic`; nó chỉ chứng minh behavior kỹ thuật.

## Contract

```text
known observations
→ deterministic formula + stable tie-break
→ RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW
→ NO external action
```

Baseline cố ý yếu: `price × commission_rate`. Nó chưa xét conversion potential, audience fit, refund/cancel risk, competition, compliance và nhiều yếu tố business khác.

```text
real evidence != RECOMMEND
RANK_SCENARIO != Approval != Execution
```

M00 evidence thật cung cấp context; M01 không tự biến provenance thật thành recommendation.
