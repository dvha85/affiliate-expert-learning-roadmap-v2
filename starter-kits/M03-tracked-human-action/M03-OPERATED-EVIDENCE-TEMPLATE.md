# M03 Operated Evidence (bằng chứng vận hành)

- DecisionPacket / decision_id (gói quyết định / mã quyết định):
- ActionRecord / action_id (bản ghi hành động / mã hành động):
- Hành động do người thực hiện:
- compliance/disclosure review (rà soát tuân thủ/công bố):
- performed_at (thời điểm thực hiện):
- measurement_window_end (cuối cửa sổ đo lường):
- OutcomeRecord / outcome_id (bản ghi kết quả / mã kết quả):
- OutcomeRecord.effect_ref (tham chiếu hành động): `effect_kind=HUMAN_ACTION`; điền `effect_id` khớp `ActionRecord.action_id`:
- observed_at + source_ref (thời điểm quan sát + nguồn):
- status (trạng thái): `PENDING | VALID | CANCELLED | REFUNDED | PAID | NO_OBSERVED_OUTCOME`; đây là dữ liệu tạm, trạng thái giao dịch hay kết luận cả window?
- So sánh observed_at với measurement_window_end (cùng thời điểm chuẩn); output ValidateActionOutcomeLink:
- Báo cáo bổ sung/đính chính: outcome_id cũ → outcome_id mới, cùng effect_ref; kỳ báo cáo, lý do thay đổi và cách tránh cộng trùng (không ghi đè record cũ):
- predicted result (kết quả dự đoán):
- observed result (kết quả quan sát):
- failure case đã thử (ca lỗi đã thử):
- limitation (giới hạn):
- next measurement (phép đo tiếp theo):
