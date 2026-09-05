# M03 Checkpoints

- [ ] ActionRecord liên kết tới DecisionPacket có thật.
- [ ] `performed_by=human`; không machine execution.
- [ ] compliance/disclosure/platform review đã ghi.
- [ ] OutcomeRecord có `effect_ref.effect_kind=HUMAN_ACTION`, `effect_ref.effect_id` khớp `ActionRecord.action_id` và không xảy ra trước action.
- [ ] measurement window/late-arriving outcome được giải thích.
- [ ] zero observed value không bị nhầm với missing.
- [ ] Reality + Operated evidence đã lưu, không chứa secret.
