---
mission_id: M05
title: First Reviewed Improvement
status: ready
minimum_evidence: E4
authority: A1 propose only
external_side_effects: false
runtime: lab/mission-runtime/
eval_pack: evals/M05-reviewed-improvement/
---

# Mission M05 — First Reviewed Improvement (cải tiến đầu tiên có review)

## Contract bàn giao

`ImprovementProposal` phải có version (phiên bản), liên kết evaluation (đánh giá), human review (người review), rollback (quay lui) và không được tự áp dụng.

## Authority ceiling

Không machine execution có side effect trong Mission này.

## PASS

`Capability + Reality + Operated`. Synthetic/offline eval chỉ tạo capability proof; learner vẫn cần evidence phù hợp Mission để claim Reality/Operated PASS.
