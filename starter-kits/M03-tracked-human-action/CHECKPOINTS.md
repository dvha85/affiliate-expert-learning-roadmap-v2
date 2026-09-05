# M03 Checkpoints

- [ ] ActionRecord liên kết tới DecisionPacket có thật.
- [ ] `performed_by=human`; không machine execution.
- [ ] compliance/disclosure/platform review đã ghi.
- [ ] OutcomeRecord có `effect_ref.effect_kind=HUMAN_ACTION`, `effect_ref.effect_id` khớp `ActionRecord.action_id` và không xảy ra trước action.
- [ ] `NO_OBSERVED_OUTCOME` trước window end bị `MEASUREMENT_WINDOW_OPEN`; đúng bằng/sau end được nhận nếu record/link hợp lệ. Đã thử timezone khác và outcome trước action.
- [ ] Phân biệt `PENDING`/trạng thái giao dịch quan sát được với kết luận cả window; không bắt mọi observation phải chờ window end.
- [ ] Báo cáo đến muộn tạo `outcome_id` mới, cùng `effect_ref`; giữ record cũ, nguồn/thời điểm và giải thích bổ sung/đính chính, không cộng trùng snapshot.
- [ ] zero observed value không bị nhầm với missing.
- [ ] Reality + Operated evidence đã lưu, không chứa secret.
