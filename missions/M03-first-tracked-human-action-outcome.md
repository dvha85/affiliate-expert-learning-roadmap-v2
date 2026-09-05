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

`NO_OBSERVED_OUTCOME` chỉ được nhận tại/sau `measurement_window_end`; trước đó vẫn có thể ghi `PENDING` hoặc trạng thái giao dịch thực sự đã quan sát, không tự kết luận cả window. Kiểm `ValidateActionOutcomeLink` cùng ActionRecord, không chỉ kiểm OutcomeRecord riêng. Báo cáo cập nhật thêm record có ID mới và cùng `effect_ref`, giữ bằng chứng cũ theo M03.2.

## Authority ceiling

Kiểm file bằng [m03-check](../docs/architecture/M03-JSON-BOUNDARY.md) trước bàn giao: JSON sai schema bị chặn trước semantic. VALID của cặp file không xác minh decision tồn tại trong store hoặc nguồn báo cáo là thật; không thay Reality/Operated proof.

Chỉ người thực hiện external action; máy chỉ validate/record.

## PASS

`Capability + Reality + Operated`. Synthetic/offline eval chỉ tạo capability proof; learner vẫn cần evidence phù hợp Mission để claim Reality/Operated PASS.
