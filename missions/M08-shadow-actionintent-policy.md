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

Một `ActionIntent` được binding tới decision/evidence thật phù hợp, có expiry/correlation/idempotency/hash, `intent_mode=PROPOSAL_ONLY` và `execution_authorized=false`; một `PolicyDecision` deterministic có version/risk/reason, `policy_mode=NON_AUTHORIZING` và `policy_review_required`; operated evidence cho tamper/expiry/linkage/duplicate và **không có live execution**.

M08 gọi toàn bộ quá trình là shadow (chạy bóng) vì Mission không tạo execution authority. Không dùng `dry_run` làm thuộc tính bản chất của ActionIntent: dry-run thuộc executor simulation, còn intent chỉ là một đề xuất có cấu trúc.

## Authority ceiling

`ALLOW` trong M08 chỉ là shadow policy result (kết quả chính sách chạy bóng). `ActionIntent.execution_authorized=false` và `PolicyDecision.execution_authorized=false` luôn bắt buộc; không nối executor, không gọi write tool, không tạo side effect.

`policy_review_required` chỉ mô tả review ở policy layer; nó không phải `ApprovalRecord` và không được dùng thay execution approval của M09+.

## PASS

`Capability + Reality + Operated` với E4 context. Synthetic/offline eval chỉ chứng minh behavior/boundary; learner phải chạy shadow policy trên artifact thật phù hợp và lưu proof rằng không có external side effect.
