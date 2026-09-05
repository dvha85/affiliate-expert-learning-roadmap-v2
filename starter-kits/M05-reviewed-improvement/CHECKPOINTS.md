# M05 Checkpoints

- [ ] EvaluationRecord nối Decision→Action→Outcome: `decision_id` khớp quyết định; `effect_ref.effect_kind=HUMAN_ACTION`, `effect_ref.effect_id` khớp `ActionRecord.action_id`; các OutcomeRecord trong `outcome_ids` có cùng `effect_ref`.
- [ ] ImprovementProposal tham chiếu evaluation IDs có thật.
- [ ] current/proposed version khác nhau và có rollback.
- [ ] `auto_apply=false`.
- [ ] JSON schema kiểm trước semantic bằng m05-check; đã thử required/null/auto-apply và broken link, ghi lỗi mà không sửa expected.
- [ ] Outcome/evaluation ID không trùng; evaluation không trước outcome, review không trước evaluation; VALID không thay decision review hoặc cấp execution.
- [ ] ReviewRecord do human tạo và liên kết đúng proposal.
- [ ] một result không bị overfit thành thay đổi lớn.
- [ ] Reality + Operated evidence đã lưu.
