# M05 Checkpoints

- [ ] EvaluationRecord nối Decision→Action→Outcome: `decision_id` khớp quyết định; `effect_ref.effect_kind=HUMAN_ACTION`, `effect_ref.effect_id` khớp `ActionRecord.action_id`; các OutcomeRecord trong `outcome_ids` có cùng `effect_ref`.
- [ ] ImprovementProposal tham chiếu evaluation IDs có thật.
- [ ] current/proposed version khác nhau và có rollback.
- [ ] `auto_apply=false`.
- [ ] ReviewRecord do human tạo và liên kết đúng proposal.
- [ ] một result không bị overfit thành thay đổi lớn.
- [ ] Reality + Operated evidence đã lưu.
