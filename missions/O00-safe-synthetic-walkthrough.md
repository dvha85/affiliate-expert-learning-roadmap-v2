---
mission_id: O00
title: Safe Synthetic Walkthrough
status: ready
minimum_evidence: E0
authority: no side effect
external_side_effects: false
runtime: lab/mission-runtime/
eval_pack: evals/O00-safe-walkthrough/
---

# O00 — Safe Synthetic Walkthrough (mô phỏng tổng thể an toàn)

Orientation để learner thấy toàn hệ thống. Không PASS Mission và không tạo Reality credit.

```bash
cd lab/mission-runtime
go test ./...
go run ./cmd/demo O00
```

Bắt buộc `external_side_effects=false` và `DRY_RUN_ONLY`.
