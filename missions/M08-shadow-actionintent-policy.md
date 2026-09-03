---
mission_id: M08
title: Shadow ActionIntent + Deterministic Policy
status: ready
minimum_evidence: E4
authority: A3-shadow
external_side_effects: false
runtime: lab/mission-runtime/
eval_pack: evals/M08-shadow-policy/
---

# Mission M08 — Shadow ActionIntent + Deterministic Policy (ActionIntent chạy bóng + chính sách tất định)

## Contract bàn giao

Một `ActionIntent` chạy bóng được binding tới decision/evidence thật phù hợp, có expiry/correlation/idempotency/hash; một `PolicyDecision` deterministic có version/risk/reason; operated evidence cho tamper/expiry/linkage/duplicate và **không có live execution**.

## Authority ceiling

`ALLOW` trong M08 chỉ là shadow policy result (kết quả chính sách chạy bóng). `execution_authorized=false` luôn bắt buộc; không nối executor, không gọi write tool, không tạo side effect.

## PASS

`Capability + Reality + Operated` với E4 context. Synthetic/offline eval chỉ chứng minh behavior/boundary; learner phải chạy shadow policy trên artifact thật phù hợp và lưu proof rằng không có external side effect.
