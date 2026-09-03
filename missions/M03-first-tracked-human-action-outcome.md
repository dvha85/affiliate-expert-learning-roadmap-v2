---
mission_id: M03
title: First Tracked Human Action + Outcome Context
status: ready
minimum_evidence: E2→E3
authority: human executes
external_side_effects: true-human-only
runtime: lab/mission-runtime/
eval_pack: evals/M03-tracked-human-action/
---

# Mission M03 — First Tracked Human Action + Outcome Context (hành động thật đầu tiên do người thực hiện + ngữ cảnh kết quả)

## Contract bàn giao

`ActionRecord` + `OutcomeRecord`: hành động bên ngoài chỉ do người thực hiện, có measurement window (cửa sổ đo lường) và compliance review (rà soát tuân thủ).

## Authority ceiling

Chỉ người thực hiện external action; máy chỉ validate/record.

## PASS

`Capability + Reality + Operated`. Synthetic/offline eval chỉ tạo capability proof; learner vẫn cần evidence phù hợp Mission để claim Reality/Operated PASS.
